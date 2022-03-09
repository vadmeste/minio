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
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/minio/madmin-go"
	"github.com/minio/minio-go/v7/pkg/set"
	"github.com/minio/minio/internal/logger"
)

const (
	defaultMonitorNewDiskInterval = time.Second * 3
	healingTrackerFilename        = ".healing.bin"
)

var newDiskHealingTimeout = newDynamicTimeout(30*time.Second, 10*time.Second)

//go:generate msgp -file $GOFILE -unexported

// healingTracker is used to persist healing information during a heal.
type healingTracker struct {
	disk StorageAPI `msg:"-"`

	ID         string
	PoolIndex  int
	SetIndex   int
	DiskIndex  int
	Path       string
	Endpoint   string
	Started    time.Time
	LastUpdate time.Time

	ObjectsTotalCount uint64
	ObjectsTotalSize  uint64

	ItemsHealed uint64
	ItemsFailed uint64

	BytesDone   uint64
	BytesFailed uint64

	// Last object scanned.
	Bucket string `json:"-"`
	Object string `json:"-"`

	// Numbers when current bucket started healing,
	// for resuming with correct numbers.
	ResumeItemsHealed uint64 `json:"-"`
	ResumeItemsFailed uint64 `json:"-"`
	ResumeBytesDone   uint64 `json:"-"`
	ResumeBytesFailed uint64 `json:"-"`

	// Filled on startup/restarts.
	QueuedBuckets []string

	// Filled during heal.
	HealedBuckets []string

	UntilBucket string
	UntilObject string

	// Add future tracking capabilities
	// Be sure that they are included in toHealingDisk
}

// loadHealingTracker will load the healing tracker from the supplied disk.
// The disk ID will be validated against the loaded one.
func loadHealingTracker(ctx context.Context, disk StorageAPI) (*healingTracker, error) {
	if disk == nil {
		return nil, errors.New("loadHealingTracker: nil disk given")
	}
	diskID, err := disk.GetDiskID()
	if err != nil {
		return nil, err
	}
	b, err := disk.ReadAll(ctx, minioMetaBucket,
		pathJoin(bucketMetaPrefix, slashSeparator, healingTrackerFilename))
	if err != nil {
		return nil, err
	}
	var h healingTracker
	_, err = h.UnmarshalMsg(b)
	if err != nil {
		return nil, err
	}
	if h.ID != diskID && h.ID != "" {
		return nil, fmt.Errorf("loadHealingTracker: disk id mismatch expected %s, got %s", h.ID, diskID)
	}
	h.disk = disk
	h.ID = diskID
	return &h, nil
}

// newHealingTracker will create a new healing tracker for the disk.
func newHealingTracker(disk StorageAPI) *healingTracker {
	diskID, _ := disk.GetDiskID()
	h := healingTracker{
		disk:     disk,
		ID:       diskID,
		Path:     disk.String(),
		Endpoint: disk.Endpoint().String(),
		Started:  time.Now().UTC(),
	}
	h.PoolIndex, h.SetIndex, h.DiskIndex = disk.GetDiskLoc()
	return &h
}

// update will update the tracker on the disk.
// If the tracker has been deleted an error is returned.
func (h *healingTracker) update(ctx context.Context) error {
	if h.disk.Healing() == nil {
		return fmt.Errorf("healingTracker: disk %q is not marked as healing", h.ID)
	}
	if h.ID == "" || h.PoolIndex < 0 || h.SetIndex < 0 || h.DiskIndex < 0 {
		h.ID, _ = h.disk.GetDiskID()
		h.PoolIndex, h.SetIndex, h.DiskIndex = h.disk.GetDiskLoc()
	}
	return h.save(ctx)
}

// save will unconditionally save the tracker and will be created if not existing.
func (h *healingTracker) save(ctx context.Context) error {
	if h.PoolIndex < 0 || h.SetIndex < 0 || h.DiskIndex < 0 {
		// Attempt to get location.
		if api := newObjectLayerFn(); api != nil {
			if ep, ok := api.(*erasureServerPools); ok {
				h.PoolIndex, h.SetIndex, h.DiskIndex, _ = ep.getPoolAndSet(h.ID)
			}
		}
	}
	h.LastUpdate = time.Now().UTC()
	htrackerBytes, err := h.MarshalMsg(nil)
	if err != nil {
		return err
	}
	globalBackgroundHealState.updateHealStatus(h)
	return h.disk.WriteAll(ctx, minioMetaBucket,
		pathJoin(bucketMetaPrefix, slashSeparator, healingTrackerFilename),
		htrackerBytes)
}

// delete the tracker on disk.
func (h *healingTracker) delete(ctx context.Context) error {
	return h.disk.Delete(ctx, minioMetaBucket,
		pathJoin(bucketMetaPrefix, slashSeparator, healingTrackerFilename),
		false)
}

func (h *healingTracker) isHealed(bucket string) bool {
	for _, v := range h.HealedBuckets {
		if v == bucket {
			return true
		}
	}
	return false
}

// resume will reset progress to the numbers at the start of the bucket.
func (h *healingTracker) resume() {
	h.ItemsHealed = h.ResumeItemsHealed
	h.ItemsFailed = h.ResumeItemsFailed
	h.BytesDone = h.ResumeBytesDone
	h.BytesFailed = h.ResumeBytesFailed
}

// bucketDone should be called when a bucket is done healing.
// Adds the bucket to the list of healed buckets and updates resume numbers.
func (h *healingTracker) bucketDone(bucket string) {
	h.ResumeItemsHealed = h.ItemsHealed
	h.ResumeItemsFailed = h.ItemsFailed
	h.ResumeBytesDone = h.BytesDone
	h.ResumeBytesFailed = h.BytesFailed

	h.HealedBuckets = append(h.HealedBuckets, bucket)
	for i, b := range h.QueuedBuckets {
		if b == bucket {
			// Delete...
			h.QueuedBuckets = append(h.QueuedBuckets[:i], h.QueuedBuckets[i+1:]...)
		}
	}

	h.Bucket = ""
	h.Object = ""
}

// setQueuedBuckets will add buckets, but exclude any that is already in h.HealedBuckets.
// Order is preserved.
func (h *healingTracker) setQueuedBuckets(buckets []BucketInfo) {
	s := set.CreateStringSet(h.HealedBuckets...)
	h.QueuedBuckets = make([]string, 0, len(buckets))
	for _, b := range buckets {
		if !s.Contains(b.Name) {
			h.QueuedBuckets = append(h.QueuedBuckets, b.Name)
		}
	}
}

func (h *healingTracker) printTo(writer io.Writer) {
	b, err := json.MarshalIndent(h, "", "  ")
	if err != nil {
		writer.Write([]byte(err.Error()))
	}
	writer.Write(b)
}

// toHealingDisk converts the information to madmin.HealingDisk
func (h *healingTracker) toHealingDisk() madmin.HealingDisk {
	return madmin.HealingDisk{
		ID:                h.ID,
		Endpoint:          h.Endpoint,
		PoolIndex:         h.PoolIndex,
		SetIndex:          h.SetIndex,
		DiskIndex:         h.DiskIndex,
		Path:              h.Path,
		Started:           h.Started.UTC(),
		LastUpdate:        h.LastUpdate.UTC(),
		ObjectsTotalCount: h.ObjectsTotalCount,
		ObjectsTotalSize:  h.ObjectsTotalSize,
		ItemsHealed:       h.ItemsHealed,
		ItemsFailed:       h.ItemsFailed,
		BytesDone:         h.BytesDone,
		BytesFailed:       h.BytesFailed,
		Bucket:            h.Bucket,
		Object:            h.Object,
		QueuedBuckets:     h.QueuedBuckets,
		HealedBuckets:     h.HealedBuckets,

		ObjectsHealed: h.ItemsHealed, // Deprecated July 2021
		ObjectsFailed: h.ItemsFailed, // Deprecated July 2021

	}
}

func initAutoHeal(ctx context.Context, objAPI ObjectLayer) {
	z, ok := objAPI.(*erasureServerPools)
	if !ok {
		return
	}

	initBackgroundHealing(ctx, objAPI) // start quick background healing

	bgSeq := mustGetHealSequence(ctx)

	globalBackgroundHealState.pushHealLocalDisks(getLocalDisksToHeal()...)

	if drivesToHeal := globalBackgroundHealState.healDriveCount(); drivesToHeal > 0 {
		logger.Info(fmt.Sprintf("Found drives to heal %d, waiting until %s to heal the content - use 'mc admin heal alias/ --verbose' to check the status",
			drivesToHeal, defaultMonitorNewDiskInterval))

		// Heal any disk format and metadata early, if possible.
		// Start with format healing
		if err := bgSeq.healDiskFormat(); err != nil {
			if newObjectLayerFn() != nil {
				// log only in situations, when object layer
				// has fully initialized.
				logger.LogIf(bgSeq.ctx, err)
			}
		}
	}

	go monitorLocalDisksAndHeal(ctx, z, bgSeq)
}

func getLocalDisksToHeal() (disksToHeal Endpoints) {
	for _, disk := range globalLocalDrives {
		_, err := disk.GetDiskID()
		if errors.Is(err, errUnformattedDisk) {
			disksToHeal = append(disksToHeal, disk.Endpoint())
			continue
		}
		if disk.Healing() != nil {
			disksToHeal = append(disksToHeal, disk.Endpoint())
		}
	}
	if len(disksToHeal) == globalEndpoints.NEndpoints() {
		// When all disks == all command line endpoints
		// this is a fresh setup, no need to trigger healing.
		return Endpoints{}
	}
	return disksToHeal
}

type freshDisksHealer struct {
	ctx context.Context

	healCtx    context.Context
	healCancel context.CancelFunc
	stopCh     chan struct{}

	z *erasureServerPools
}

func (e *freshDisksHealer) updateHealingTrack(endpoint Endpoint) error {
	disk, format, err := connectEndpoint(endpoint)
	if err != nil {
		return err
	}

	poolIndex := globalEndpoints.GetLocalPoolIdx(disk.Endpoint())
	if poolIndex < 0 {
		return errors.New("unexpected pool index")
	}

	// Calculate the set index where the current endpoint belongs
	e.z.serverPools[poolIndex].erasureDisksMu.RLock()
	// Protect reading reference format.
	setIndex, _, err := findDiskIndex(e.z.serverPools[poolIndex].format, format)
	e.z.serverPools[poolIndex].erasureDisksMu.RUnlock()
	if err != nil {
		return err
	}

	// So someone changed the drives underneath, healing tracker missing.
	tracker, err := loadHealingTracker(e.ctx, disk)
	if err != nil {
		logger.LogIf(e.ctx, fmt.Errorf("Healing tracker missing on '%s', disk was swapped again on %s pool: %w",
			disk, humanize.Ordinal(poolIndex+1), err))
		tracker = newHealingTracker(disk)
	}

	if len(tracker.QueuedBuckets) > 0 {
		// This is a new tracker, update it
		return nil
	}

	buckets, _ := e.z.ListBuckets(e.ctx)

	// Buckets data are dispersed in multiple zones/sets, make
	// sure to heal all bucket metadata configuration.
	buckets = append(buckets, BucketInfo{
		Name: pathJoin(minioMetaBucket, minioConfigPrefix),
	}, BucketInfo{
		Name: pathJoin(minioMetaBucket, bucketMetaPrefix),
	})

	// Heal latest buckets first.
	sort.Slice(buckets, func(i, j int) bool {
		a, b := strings.HasPrefix(buckets[i].Name, minioMetaBucket), strings.HasPrefix(buckets[j].Name, minioMetaBucket)
		if a != b {
			return a
		}
		return buckets[i].Created.After(buckets[j].Created)
	})

	// Load bucket totals
	cache := dataUsageCache{}
	if err := cache.load(e.ctx, e.z.serverPools[poolIndex].sets[setIndex], dataUsageCacheName); err == nil {
		dataUsageInfo := cache.dui(dataUsageRoot, nil)
		tracker.ObjectsTotalCount = dataUsageInfo.ObjectsTotalCount
		tracker.ObjectsTotalSize = dataUsageInfo.ObjectsTotalSize
	}

	tracker.PoolIndex, tracker.SetIndex, tracker.DiskIndex = disk.GetDiskLoc()
	tracker.setQueuedBuckets(buckets)
	return tracker.save(e.ctx)
}

func healingResumeMiddleHealing(trackers healingTrackers) (pool, set int, bucket, object string, until bool) {
	for _, t := range trackers {
		if t.Bucket == "" {
			continue
		}
		if bucket == "" {
			pool, set, bucket, object = t.PoolIndex, t.SetIndex, t.Bucket, t.Object
			continue
		}
		if t.Bucket < bucket {
			pool, set, bucket, object = t.PoolIndex, t.SetIndex, t.Bucket, t.Object
		}
	}

	if bucket != "" {
		return
	}

	until = true

	for _, t := range trackers {
		if t.UntilBucket == "" {
			continue
		}
		if bucket == "" {
			pool, set, bucket, object = t.PoolIndex, t.SetIndex, t.UntilBucket, t.UntilObject
			continue
		}
		if t.UntilBucket < bucket {
			pool, set, bucket, object = t.PoolIndex, t.SetIndex, t.UntilBucket, t.UntilObject
		}
	}
	return
}

func healingResumeNextBucket(trackers healingTrackers) (pool, set int, bucket string) {
	for _, t := range trackers {
		for _, b := range t.QueuedBuckets {
			if bucket == "" {
				pool, set, bucket = t.PoolIndex, t.SetIndex, b
				continue
			}
			if b < bucket {
				pool, set, bucket = t.PoolIndex, t.SetIndex, b
			}
		}
	}

	return
}

func (e *freshDisksHealer) stop() {
	if e.healCancel != nil {
		e.healCancel()
		<-e.stopCh
	}
}

func (e *freshDisksHealer) resume() {
	e.healCtx, e.healCancel = context.WithCancel(e.ctx)
	e.stopCh = make(chan struct{})
	go func() {
		defer close(e.stopCh)

		for {
			finished, err := e.run()
			if finished || contextCanceled(e.healCtx) {
				break
			}

			logger.LogIf(e.ctx, err)
		}
	}()
}

func (e *freshDisksHealer) run() (bool, error) {
	var trackers healingTrackers
	for _, disk := range globalLocalDrives {
		// So someone changed the drives underneath, healing tracker missing.
		tracker, err := loadHealingTracker(e.healCtx, disk)
		if err != nil {
			if errors.Is(err, errFileNotFound) {
				continue
			}
			tracker = newHealingTracker(disk)
		}
		trackers = append(trackers, tracker)
	}

	if len(trackers) == 0 {
		return true, nil
	}

	fmt.Println("scan tracking files")
	for _, t := range trackers {
		fmt.Println(t.disk, t.Bucket, t.Object, "|", t.UntilBucket, t.UntilObject)
	}

	var start, end string

	poolIndex, setIndex, bucket, object, reverse := healingResumeMiddleHealing(trackers)
	if bucket != "" {
		start = object
		if reverse {
			start = ""
			end = object
		} else {
			trackers.updateStartAt(e.ctx, bucket, object)
		}
	} else {
		poolIndex, setIndex, bucket = healingResumeNextBucket(trackers)
		trackers.updateStartAt(e.ctx, "", "")
	}

	// Prevent parallel erasure set healing
	locker := e.z.NewNSLock(minioMetaBucket, fmt.Sprintf("new-disk-healing/%d/%d", poolIndex, setIndex))
	lkctx, err := locker.GetLock(e.healCtx, newDiskHealingTimeout)
	if err != nil {
		return false, err
	}
	ctx := lkctx.Context()
	defer locker.Unlock(lkctx.Cancel)

	err = e.z.serverPools[poolIndex].sets[setIndex].healBucketAndContents(ctx, bucket, start, end, trackers)
	if err != nil {
		trackers.update(e.ctx)
		return false, err
	}

	select {
	// If context is canceled don't mark as done...
	case <-ctx.Done():
		trackers.update(e.ctx)
		return false, e.healCtx.Err()
	default:
		trackers.bucketDone(bucket)
		trackers.update(e.ctx)
		trackers.cleanup(e.ctx)
	}

	// fmt.Println("finished healing", bucket, object)

	return false, nil
}

func newFreshDisksHealer(ctx context.Context, z *erasureServerPools) *freshDisksHealer {
	return &freshDisksHealer{ctx: ctx, z: z}
}

// monitorLocalDisksAndHeal - ensures that detected new disks are healed
//  1. Only the concerned erasure set will be listed and healed
//  2. Only the node hosting the disk is responsible to perform the heal
func monitorLocalDisksAndHeal(ctx context.Context, z *erasureServerPools, bgSeq *healSequence) {
	// Perform automatic disk healing when a disk is replaced locally.
	diskCheckTimer := time.NewTimer(defaultMonitorNewDiskInterval)
	defer diskCheckTimer.Stop()

	fdh := newFreshDisksHealer(ctx, z)

	for {
		select {
		case <-ctx.Done():
			return
		case <-diskCheckTimer.C:
			healDisks := globalBackgroundHealState.getHealLocalDiskEndpoints()
			if len(healDisks) == 0 {
				// Should not happen
				break
			}
			globalBackgroundHealState.popHealLocalDisks(healDisks...)

			fdh.stop()

			// Reformat disks immediately
			_, err := z.HealFormat(context.Background(), false)
			if err != nil && !errors.Is(err, errNoHealRequired) {
				logger.LogIf(ctx, err)
				continue
			}

			for _, disk := range healDisks {
				logger.LogIf(ctx, fdh.updateHealingTrack(disk))
			}

			fdh.resume()
		}

		// Reset for next interval.
		diskCheckTimer.Reset(defaultMonitorNewDiskInterval)

	}
}
