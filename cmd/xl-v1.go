/*
 * MinIO Cloud Storage, (C) 2016, 2017, 2018 MinIO, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package cmd

import (
	"context"
	"math/rand"
	"sort"
	"strings"
	"time"

	"github.com/minio/minio/cmd/logger"
	"github.com/minio/minio/pkg/bpool"
	"github.com/minio/minio/pkg/dsync"
	"github.com/minio/minio/pkg/madmin"
	xnet "github.com/minio/minio/pkg/net"
	"github.com/minio/minio/pkg/sync/errgroup"
)

// XL constants.
const (
	// XL metadata file carries per object metadata.
	xlMetaJSONFile = "xl.json"
)

// OfflineDisk represents an unavailable disk.
var OfflineDisk StorageAPI // zero value is nil

// xlObjects - Implements XL object layer.
type xlObjects struct {
	// getDisks returns list of storageAPIs.
	getDisks func() []StorageAPI

	// getLockers returns list of remote and local lockers.
	getLockers func() []dsync.NetLocker

	// Locker mutex map.
	nsMutex *nsLockMap

	// Byte pools used for temporary i/o buffers.
	bp *bpool.BytePoolCap

	// TODO: Deprecated only kept here for tests, should be removed in future.
	storageDisks []StorageAPI

	// TODO: ListObjects pool management, should be removed in future.
	listPool *TreeWalkPool
}

// NewNSLock - initialize a new namespace RWLocker instance.
func (xl xlObjects) NewNSLock(ctx context.Context, bucket string, object string) RWLocker {
	return xl.nsMutex.NewNSLock(ctx, xl.getLockers(), bucket, object)
}

// Shutdown function for object storage interface.
func (xl xlObjects) Shutdown(ctx context.Context) error {
	// Add any object layer shutdown activities here.
	closeStorageDisks(xl.getDisks())
	closeLockers(xl.getLockers())
	return nil
}

// byDiskTotal is a collection satisfying sort.Interface.
type byDiskTotal []DiskInfo

func (d byDiskTotal) Len() int      { return len(d) }
func (d byDiskTotal) Swap(i, j int) { d[i], d[j] = d[j], d[i] }
func (d byDiskTotal) Less(i, j int) bool {
	return d[i].Total < d[j].Total
}

// getDisksInfo - fetch disks info across all other storage API.
func getDisksInfo(disks []StorageAPI) (disksInfo []DiskInfo, onlineDisks, offlineDisks madmin.BackendDisks) {
	disksInfo = make([]DiskInfo, len(disks))

	g := errgroup.WithNErrs(len(disks))
	for index := range disks {
		index := index
		g.Go(func() error {
			if disks[index] == nil {
				// Storage disk is empty, perhaps ignored disk or not available.
				return errDiskNotFound
			}
			info, err := disks[index].DiskInfo()
			if err != nil {
				if IsErr(err, baseErrs...) {
					return err
				}
				reqInfo := (&logger.ReqInfo{}).AppendTags("disk", disks[index].String())
				ctx := logger.SetReqInfo(context.Background(), reqInfo)
				logger.LogIf(ctx, err)
			}
			disksInfo[index] = info
			return nil
		}, index)
	}

	getPeerAddress := func(diskPath string) (string, error) {
		hostPort := strings.Split(diskPath, SlashSeparator)[0]
		// Host will be empty for xl/fs disk paths.
		if hostPort == "" {
			return "", nil
		}
		thisAddr, err := xnet.ParseHost(hostPort)
		if err != nil {
			return "", err
		}
		return thisAddr.String(), nil
	}

	onlineDisks = make(madmin.BackendDisks)
	offlineDisks = make(madmin.BackendDisks)
	// Wait for the routines.
	for i, err := range g.Wait() {
		peerAddr, pErr := getPeerAddress(disksInfo[i].RelativePath)

		if pErr != nil {
			continue
		}
		if _, ok := offlineDisks[peerAddr]; !ok {
			offlineDisks[peerAddr] = 0
		}
		if _, ok := onlineDisks[peerAddr]; !ok {
			onlineDisks[peerAddr] = 0
		}
		if err != nil {
			offlineDisks[peerAddr]++
			continue
		}
		onlineDisks[peerAddr]++
	}

	// Success.
	return disksInfo, onlineDisks, offlineDisks
}

// returns sorted disksInfo slice which has only valid entries.
// i.e the entries where the total size of the disk is not stated
// as 0Bytes, this means that the disk is not online or ignored.
func sortValidDisksInfo(disksInfo []DiskInfo) []DiskInfo {
	var validDisksInfo []DiskInfo
	for _, diskInfo := range disksInfo {
		if diskInfo.Total == 0 {
			continue
		}
		validDisksInfo = append(validDisksInfo, diskInfo)
	}
	sort.Sort(byDiskTotal(validDisksInfo))
	return validDisksInfo
}

// Get an aggregated storage info across all disks.
func getStorageInfo(disks []StorageAPI) StorageInfo {
	disksInfo, onlineDisks, offlineDisks := getDisksInfo(disks)

	// Sort so that the first element is the smallest.
	validDisksInfo := sortValidDisksInfo(disksInfo)
	// If there are no valid disks, set total and free disks to 0
	if len(validDisksInfo) == 0 {
		return StorageInfo{}
	}

	// Combine all disks to get total usage
	usedList := make([]uint64, len(validDisksInfo))
	totalList := make([]uint64, len(validDisksInfo))
	availableList := make([]uint64, len(validDisksInfo))
	mountPaths := make([]string, len(validDisksInfo))

	for i, di := range validDisksInfo {
		usedList[i] = di.Used
		totalList[i] = di.Total
		availableList[i] = di.Free
		mountPaths[i] = di.RelativePath
	}

	storageInfo := StorageInfo{
		Used:       usedList,
		Total:      totalList,
		Available:  availableList,
		MountPaths: mountPaths,
	}

	storageInfo.Backend.Type = BackendErasure
	storageInfo.Backend.OnlineDisks = onlineDisks
	storageInfo.Backend.OfflineDisks = offlineDisks

	return storageInfo
}

// StorageInfo - returns underlying storage statistics.
func (xl xlObjects) StorageInfo(ctx context.Context) StorageInfo {
	return getStorageInfo(xl.getDisks())
}

func objHisto(usize uint64) string {
	size := int64(usize)

	var interval objectHistogramInterval
	for _, interval = range ObjectsHistogramIntervals {
		var cond1, cond2 bool
		if size >= interval.start || interval.start == -1 {
			cond1 = true
		}
		if size <= interval.end || interval.end == -1 {
			cond2 = true
		}
		if cond1 && cond2 {
			return interval.name
		}
	}
	// This would be the last element
	return interval.name
}

func (xl xlObjects) walkObjLayerInfo(ctx context.Context, disk StorageAPI, info *ObjectLayerInfo, bucketName, prefix string) {
	entries, err := disk.ListDir(bucketName, prefix, -1, xlMetaJSONFile)
	if err != nil {
		return
	}
	for _, entry := range entries {
		if strings.HasSuffix(entry, "/") {
			xl.walkObjLayerInfo(ctx, disk, info, bucketName, pathJoin(prefix, entry))
		} else {
			info.ObjectsCount += 1
			meta, err := readXLMeta(ctx, disk, bucketName, pathJoin(prefix, entry))
			if err != nil {
				continue
			}
			info.TotalSize += uint64(meta.Stat.Size)
			info.BucketsSizes[bucketName] += uint64(meta.Stat.Size)
			info.ObjectsSizesHistogram[objHisto(uint64(meta.Stat.Size))] += 1
		}
	}
	return
}

func (xl xlObjects) ObjectLayerInfo(ctx context.Context) ObjectLayerInfo {
	var disks []StorageAPI
	for _, d := range xl.getDisks() {
		if d == nil {
			continue
		}
		if d.IsLocal() && d.IsOnline() {
			disks = append(disks, d)
		}
	}
	if len(disks) == 0 {
		return ObjectLayerInfo{}
	}

	// Pick a random disk
	rand.Seed(time.Now().Unix())
	disk := disks[rand.Intn(len(disks))]

	volsInfo, err := disk.ListVols()
	if err != nil {
		return ObjectLayerInfo{}
	}

	var info = ObjectLayerInfo{
		BucketsSizes:          make(map[string]uint64),
		ObjectsSizesHistogram: make(map[string]uint64),
	}

	for _, vol := range volsInfo {
		if isReservedOrInvalidBucket(vol.Name, false) {
			continue
		}
		info.BucketsCount++
		xl.walkObjLayerInfo(ctx, disk, &info, vol.Name, "")
	}

	info.LastUpdate = UTCNow()
	return info
}
