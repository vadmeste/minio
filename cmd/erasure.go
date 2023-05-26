// Copyright (c) 2015-2021 MinIO, Inc.
//
// This file is part of MinIO Object Storage stack
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package cmd

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/minio/madmin-go/v3"
	"github.com/minio/minio/internal/bpool"
	"github.com/minio/minio/internal/dsync"
	"github.com/minio/minio/internal/logger"
	"github.com/minio/pkg/sync/errgroup"
)

// list all errors that can be ignore in a bucket operation.
var bucketOpIgnoredErrs = append(baseIgnoredErrs, errDiskAccessDenied, errUnformattedDisk)

// list all errors that can be ignored in a bucket metadata operation.
var bucketMetadataOpIgnoredErrs = append(bucketOpIgnoredErrs, errVolumeNotFound)

// OfflineDisk represents an unavailable disk.
var OfflineDisk StorageAPI // zero value is nil

// erasureObjects - Implements ER object layer.
type erasureObjects struct {
	setDriveCount      int
	defaultParityCount int

	setIndex  int
	poolIndex int

	// getDisks returns list of storageAPIs.
	getDisks func() []StorageAPI

	// getLockers returns list of remote and local lockers.
	getLockers func() ([]dsync.NetLocker, string)

	// getEndpoints returns list of endpoint strings belonging this set.
	// some may be local and some remote.
	getEndpoints func() []Endpoint

	// Locker mutex map.
	nsMutex *nsLockMap

	// Byte pools used for temporary i/o buffers.
	bp *bpool.BytePoolCap

	// Byte pools used for temporary i/o buffers,
	// legacy objects.
	bpOld *bpool.BytePoolCap
}

// NewNSLock - initialize a new namespace RWLocker instance.
func (er erasureObjects) NewNSLock(bucket string, objects ...string) RWLocker {
	return er.nsMutex.NewNSLock(er.getLockers, bucket, objects...)
}

// Shutdown function for object storage interface.
func (er erasureObjects) Shutdown(ctx context.Context) error {
	// Add any object layer shutdown activities here.
	closeStorageDisks(er.getDisks()...)
	return nil
}

// defaultWQuorum write quorum based on setDriveCount and defaultParityCount
func (er erasureObjects) defaultWQuorum() int {
	dataCount := er.setDriveCount - er.defaultParityCount
	if dataCount == er.defaultParityCount {
		return dataCount + 1
	}
	return dataCount
}

func diskErrToDriveState(err error) (state string) {
	switch {
	case errors.Is(err, errDiskNotFound) || errors.Is(err, context.DeadlineExceeded):
		state = madmin.DriveStateOffline
	case errors.Is(err, errCorruptedFormat):
		state = madmin.DriveStateCorrupt
	case errors.Is(err, errUnformattedDisk):
		state = madmin.DriveStateUnformatted
	case errors.Is(err, errDiskAccessDenied):
		state = madmin.DriveStatePermission
	case errors.Is(err, errFaultyDisk):
		state = madmin.DriveStateFaulty
	case err == nil:
		state = madmin.DriveStateOk
	default:
		state = fmt.Sprintf("%s (cause: %s)", madmin.DriveStateUnknown, err)
	}

	return
}

func getOnlineOfflineDisksStats(disksInfo []madmin.Disk) (onlineDisks, offlineDisks madmin.BackendDisks) {
	onlineDisks = make(madmin.BackendDisks)
	offlineDisks = make(madmin.BackendDisks)

	for _, disk := range disksInfo {
		ep := disk.Endpoint
		if _, ok := offlineDisks[ep]; !ok {
			offlineDisks[ep] = 0
		}
		if _, ok := onlineDisks[ep]; !ok {
			onlineDisks[ep] = 0
		}
	}

	// Wait for the routines.
	for _, disk := range disksInfo {
		ep := disk.Endpoint
		state := disk.State
		if state != madmin.DriveStateOk && state != madmin.DriveStateUnformatted {
			offlineDisks[ep]++
			continue
		}
		onlineDisks[ep]++
	}

	rootDiskCount := 0
	for _, di := range disksInfo {
		if di.RootDisk {
			rootDiskCount++
		}
	}

	// Count offline disks as well to ensure consistent
	// reportability of offline drives on local setups.
	if len(disksInfo) == (rootDiskCount + offlineDisks.Sum()) {
		// Success.
		return onlineDisks, offlineDisks
	}

	// Root disk should be considered offline
	for i := range disksInfo {
		ep := disksInfo[i].Endpoint
		if disksInfo[i].RootDisk {
			offlineDisks[ep]++
			onlineDisks[ep]--
		}
	}

	return onlineDisks, offlineDisks
}

// getDisksInfo - fetch disks info across all other storage API.
func getDisksInfo(disks []StorageAPI, endpoints []Endpoint) (disksInfo []madmin.Disk) {
	disksInfo = make([]madmin.Disk, len(disks))

	g := errgroup.WithNErrs(len(disks))
	for index := range disks {
		index := index
		g.Go(func() error {
			diskEndpoint := endpoints[index].String()
			if disks[index] == OfflineDisk {
				logger.LogOnceIf(GlobalContext, fmt.Errorf("%s: %s", errDiskNotFound, endpoints[index]), "get-disks-info-offline-"+diskEndpoint)
				disksInfo[index] = madmin.Disk{
					State:    diskErrToDriveState(errDiskNotFound),
					Endpoint: diskEndpoint,
				}
				return nil
			}
			info, err := disks[index].DiskInfo(context.TODO())
			di := madmin.Disk{
				Endpoint:       diskEndpoint,
				DrivePath:      info.MountPath,
				TotalSpace:     info.Total,
				UsedSpace:      info.Used,
				AvailableSpace: info.Free,
				UUID:           info.ID,
				Major:          info.Major,
				Minor:          info.Minor,
				RootDisk:       info.RootDisk,
				Healing:        info.Healing,
				Scanning:       info.Scanning,
				State:          diskErrToDriveState(err),
				FreeInodes:     info.FreeInodes,
				UsedInodes:     info.UsedInodes,
			}
			di.PoolIndex, di.SetIndex, di.DiskIndex = disks[index].GetDiskLoc()
			if info.Healing {
				if hi := disks[index].Healing(); hi != nil {
					hd := hi.toHealingDisk()
					di.HealInfo = &hd
				}
			}
			di.Metrics = &madmin.DiskMetrics{
				LastMinute: make(map[string]madmin.TimedAction, len(info.Metrics.LastMinute)),
				APICalls:   make(map[string]uint64, len(info.Metrics.APICalls)),
			}
			for k, v := range info.Metrics.LastMinute {
				if v.N > 0 {
					di.Metrics.LastMinute[k] = v.asTimedAction()
				}
			}
			for k, v := range info.Metrics.APICalls {
				di.Metrics.APICalls[k] = v
			}
			if info.Total > 0 {
				di.Utilization = float64(info.Used / info.Total * 100)
			}
			disksInfo[index] = di
			return nil
		}, index)
	}

	g.Wait()
	return disksInfo
}

// Get an aggregated storage info across all disks.
func getStorageInfo(disks []StorageAPI, endpoints []Endpoint) StorageInfo {
	disksInfo := getDisksInfo(disks, endpoints)

	// Sort so that the first element is the smallest.
	sort.Slice(disksInfo, func(i, j int) bool {
		return disksInfo[i].TotalSpace < disksInfo[j].TotalSpace
	})

	storageInfo := StorageInfo{
		Disks: disksInfo,
	}

	storageInfo.Backend.Type = madmin.Erasure
	return storageInfo
}

// StorageInfo - returns underlying storage statistics.
func (er erasureObjects) StorageInfo(ctx context.Context) StorageInfo {
	disks := er.getDisks()
	endpoints := er.getEndpoints()
	return getStorageInfo(disks, endpoints)
}

// LocalStorageInfo - returns underlying local storage statistics.
func (er erasureObjects) LocalStorageInfo(ctx context.Context) StorageInfo {
	disks := er.getDisks()
	endpoints := er.getEndpoints()

	var localDisks []StorageAPI
	var localEndpoints []Endpoint

	for i, endpoint := range endpoints {
		if endpoint.IsLocal {
			localDisks = append(localDisks, disks[i])
			localEndpoints = append(localEndpoints, endpoint)
		}
	}

	return getStorageInfo(localDisks, localEndpoints)
}

func (er erasureObjects) getOnlineDisksWithHealing() (newDisks []StorageAPI, healing bool) {
	var wg sync.WaitGroup
	disks := er.getDisks()
	infos := make([]DiskInfo, len(disks))
	for _, i := range hashOrder(UTCNow().String(), len(disks)) {
		i := i
		wg.Add(1)
		go func() {
			defer wg.Done()

			disk := disks[i-1]

			if disk == nil {
				infos[i-1].Error = "nil drive"
				return
			}

			di, err := disk.DiskInfo(context.Background())
			if err != nil {
				// - Do not consume disks which are not reachable
				//   unformatted or simply not accessible for some reason.
				//
				//
				// - Future: skip busy disks
				infos[i-1].Error = err.Error()
				return
			}

			infos[i-1] = di
		}()
	}
	wg.Wait()

	for i, info := range infos {
		// Check if one of the drives in the set is being healed.
		// this information is used by scanner to skip healing
		// this erasure set while it calculates the usage.
		if info.Healing || info.Error != "" {
			healing = true
			continue
		}
		newDisks = append(newDisks, disks[i])
	}

	return newDisks, healing
}

// Clean-up previously deleted objects. from .minio.sys/tmp/.trash/
func (er erasureObjects) cleanupDeletedObjects(ctx context.Context) {
	// run multiple cleanup's local to this server.
	var wg sync.WaitGroup
	for _, disk := range er.getLoadBalancedLocalDisks() {
		if disk != nil {
			wg.Add(1)
			go func(disk StorageAPI) {
				defer wg.Done()
				diskPath := disk.Endpoint().Path
				readDirFn(pathJoin(diskPath, minioMetaTmpDeletedBucket), func(ddir string, typ os.FileMode) error {
					wait := deletedCleanupSleeper.Timer(ctx)
					removeAll(pathJoin(diskPath, minioMetaTmpDeletedBucket, ddir))
					wait()
					return nil
				})
			}(disk)
		}
	}
	wg.Wait()
}

// nsScanner will start scanning buckets and send updated totals as they are traversed.
// Updates are sent on a regular basis and the caller *must* consume them.
func (er erasureObjects) nsScanner(ctx context.Context, disk StorageAPI, bucket string, healing bool, wantCycle uint32, updates chan<- dataUsageCache, healScanMode madmin.HealScanMode) error {
	cacheName := pathJoin(bucket, dataUsageCacheName)

	// Load bucket totals
	cache := dataUsageCache{}
	if err := cache.load(ctx, er, cacheName); err != nil {
		return err
	}

	cache.Info = dataUsageCacheInfo{
		Name:        dataUsageRoot,
		NextCycle:   wantCycle,
		SkipHealing: healing,
	}

	bucketResults := make(chan dataUsageEntryInfo, 1)

	// Start async collector/saver.
	// This goroutine owns the cache.
	var saverWg sync.WaitGroup
	saverWg.Add(1)
	go func() {
		// Add jitter to the update time so multiple sets don't sync up.
		updateTime := 30*time.Second + time.Duration(float64(10*time.Second)*rand.Float64())
		t := time.NewTicker(updateTime)
		defer t.Stop()
		defer saverWg.Done()
		var lastSave time.Time

		for {
			select {
			case <-t.C:
				if cache.Info.LastUpdate.Equal(lastSave) {
					continue
				}
				logger.LogOnceIf(ctx, cache.save(ctx, er, cacheName), "nsscanner-cache-update")
				updates <- cache.clone()
				lastSave = cache.Info.LastUpdate
			case v, ok := <-bucketResults:
				if !ok {
					// Save final state...
					cache.Info.NextCycle = wantCycle
					cache.Info.LastUpdate = time.Now()
					logger.LogOnceIf(ctx, cache.save(ctx, er, cacheName), "nsscanner-channel-closed")
					updates <- cache
					return
				}
				cache.replace(v.Name, v.Parent, v.Entry)
				cache.Info.LastUpdate = time.Now()
			}
		}
	}()

	// Collect updates.
	diskUpdates := make(chan dataUsageEntry, 1)
	var wg sync.WaitGroup
	wg.Add(1)
	go func(name string) {
		defer wg.Done()
		for update := range diskUpdates {
			select {
			case <-ctx.Done():
			case bucketResults <- dataUsageEntryInfo{
				Name:   name,
				Parent: dataUsageRoot,
				Entry:  update,
			}:
			}
		}
	}(bucket)

	newCache := dataUsageCache{}

	err := newCache.load(ctx, er, cacheName)
	if err != nil {
		return err
	}

	newCache.Info.Name = bucket
	newCache.Info.SkipHealing = healing
	newCache.Info.NextCycle = wantCycle

	newCache, err = disk.NSScanner(ctx, newCache, diskUpdates, healScanMode)
	// This ensures that we don't close
	// bucketResults channel while the
	// updates-collector goroutine still
	// holds a reference to this.
	wg.Wait()
	if err != nil {
		logger.LogIf(ctx, err)
		return err
	}

	var root dataUsageEntry
	if r := newCache.root(); r != nil {
		root = newCache.flatten(*r)
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case bucketResults <- dataUsageEntryInfo{
		Name:   bucket,
		Parent: dataUsageRoot,
		Entry:  root,
	}:
	}

	close(bucketResults)
	saverWg.Wait()

	return nil
}
