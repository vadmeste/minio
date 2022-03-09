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
	"sort"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/minio/madmin-go"
	"github.com/minio/minio/internal/color"
	"github.com/minio/minio/internal/config/storageclass"
	"github.com/minio/minio/internal/logger"
	"github.com/minio/pkg/console"
	"github.com/minio/pkg/wildcard"
)

const (
	bgHealingUUID = "0000-0000-0000-0000"
)

// NewBgHealSequence creates a background healing sequence
// operation which scans all objects and heal them.
func newBgHealSequence() *healSequence {
	reqInfo := &logger.ReqInfo{API: "BackgroundHeal"}
	ctx, cancelCtx := context.WithCancel(logger.SetReqInfo(GlobalContext, reqInfo))

	hs := madmin.HealOpts{
		// Remove objects that do not have read-quorum
		Remove: healDeleteDangling,
	}

	return &healSequence{
		respCh:      make(chan healResult),
		startTime:   UTCNow(),
		clientToken: bgHealingUUID,
		// run-background heal with reserved bucket
		bucket:   minioReservedBucket,
		settings: hs,
		currentStatus: healSequenceStatus{
			Summary:      healNotStartedStatus,
			HealSettings: hs,
		},
		cancelCtx:          cancelCtx,
		ctx:                ctx,
		reportProgress:     false,
		scannedItemsMap:    make(map[madmin.HealItemType]int64),
		healedItemsMap:     make(map[madmin.HealItemType]int64),
		healFailedItemsMap: make(map[string]int64),
	}
}

// getBackgroundHealStatus will return the
func getBackgroundHealStatus(ctx context.Context, o ObjectLayer) (madmin.BgHealState, bool) {
	if globalBackgroundHealState == nil {
		return madmin.BgHealState{}, false
	}

	bgSeq, ok := globalBackgroundHealState.getHealSequenceByToken(bgHealingUUID)
	if !ok {
		return madmin.BgHealState{}, false
	}

	status := madmin.BgHealState{
		ScannedItemsCount: bgSeq.getScannedItemsCount(),
	}

	if globalMRFState.initialized() {
		status.MRF = map[string]madmin.MRFStatus{
			globalLocalNodeName: globalMRFState.getCurrentMRFRoundInfo(),
		}
	}

	healDisksMap := map[string]struct{}{}
	for _, ep := range getLocalDisksToHeal() {
		healDisksMap[ep.String()] = struct{}{}
	}

	if o == nil {
		healing := globalBackgroundHealState.getLocalHealingDisks()
		for _, disk := range healing {
			status.HealDisks = append(status.HealDisks, disk.Endpoint)
		}

		return status, true
	}

	// ignores any errors here.
	si, _ := o.StorageInfo(ctx)

	indexed := make(map[string][]madmin.Disk)
	for _, disk := range si.Disks {
		setIdx := fmt.Sprintf("%d-%d", disk.PoolIndex, disk.SetIndex)
		indexed[setIdx] = append(indexed[setIdx], disk)
	}

	for id, disks := range indexed {
		ss := madmin.SetStatus{
			ID:        id,
			SetIndex:  disks[0].SetIndex,
			PoolIndex: disks[0].PoolIndex,
		}
		for _, disk := range disks {
			ss.Disks = append(ss.Disks, disk)
			if disk.Healing {
				ss.HealStatus = "Healing"
				ss.HealPriority = "high"
				status.HealDisks = append(status.HealDisks, disk.Endpoint)
			}
		}
		sortDisks(ss.Disks)
		status.Sets = append(status.Sets, ss)
	}
	sort.Slice(status.Sets, func(i, j int) bool {
		return status.Sets[i].ID < status.Sets[j].ID
	})

	backendInfo := o.BackendInfo()
	status.SCParity = make(map[string]int)
	status.SCParity[storageclass.STANDARD] = backendInfo.StandardSCParity
	status.SCParity[storageclass.RRS] = backendInfo.RRSCParity

	return status, true
}

func mustGetHealSequence(ctx context.Context) *healSequence {
	// Get background heal sequence to send elements to heal
	for {
		globalHealStateLK.RLock()
		hstate := globalBackgroundHealState
		globalHealStateLK.RUnlock()

		if hstate == nil {
			time.Sleep(time.Second)
			continue
		}

		bgSeq, ok := hstate.getHealSequenceByToken(bgHealingUUID)
		if !ok {
			time.Sleep(time.Second)
			continue
		}
		return bgSeq
	}
}

func (er *erasureObjects) healBucketAndContents(ctx context.Context, bucket, startObject, endObject string, trackers healingTrackers) error {
	fmt.Println("!! healbucketAndContents, bucket=", bucket, ", startObject=", startObject, ", endObject=", endObject)

	bgSeq := mustGetHealSequence(ctx)
	scanMode := madmin.HealNormalScan

	// Heal current bucket
	if _, err := er.HealBucket(ctx, bucket, madmin.HealOpts{
		ScanMode: scanMode,
	}); err != nil {
		return err
	}

	if serverDebugLog {
		console.Debugf(color.Green("healDisk:")+" healing bucket %s content on %s erasure set\n",
			bucket, humanize.Ordinal(trackers.getSetIndex()+1))
	}

	disks := er.getOnlineDisks()
	if len(disks) == 0 {
		return errors.New("no online disks found")
	}

	// Limit listing to 3 drives.
	if len(disks) > 3 {
		disks = disks[:3]
	}

	healEntry := func(entry metaCacheEntry) {
		if entry.name == "" && len(entry.metadata) == 0 {
			// ignore entries that don't have metadata.
			return
		}
		if entry.isDir() {
			// ignore healing entry.name's with `/` suffix.
			return
		}
		// We might land at .metacache, .trash, .multipart
		// no need to heal them skip, only when bucket
		// is '.minio.sys'
		if bucket == minioMetaBucket {
			if wildcard.Match("buckets/*/.metacache/*", entry.name) {
				return
			}
			if wildcard.Match("tmp/.trash/*", entry.name) {
				return
			}
			if wildcard.Match("multipart/*", entry.name) {
				return
			}
		}

		fivs, err := entry.fileInfoVersions(bucket)
		if err != nil {
			err := bgSeq.queueHealTask(healSource{
				bucket:    bucket,
				object:    entry.name,
				versionID: "",
			}, madmin.HealItemObject)
			if err != nil {
				trackers.incItemsFailed()
				logger.LogIf(ctx, fmt.Errorf("unable to heal object %s/%s: %w", bucket, entry.name, err))
			} else {
				trackers.incItemsFailed()
			}
			bgSeq.logHeal(madmin.HealItemObject)
			return
		}

		for _, version := range fivs.Versions {
			if _, err := er.HealObject(ctx, bucket, version.Name,
				version.VersionID, madmin.HealOpts{
					ScanMode: scanMode,
					Remove:   healDeleteDangling,
				}); err != nil {
				// If not deleted, assume they failed.
				trackers.incItemsFailed()
				trackers.incBytesFailed(uint64(version.Size))
				if version.VersionID != "" {
					logger.LogIf(ctx, fmt.Errorf("unable to heal object %s/%s-v(%s): %w", bucket, version.Name, version.VersionID, err))
				} else {
					logger.LogIf(ctx, fmt.Errorf("unable to heal object %s/%s: %w", bucket, version.Name, err))
				}
			} else {
				trackers.incItemsHealed()
				trackers.incBytesDone(uint64(version.Size))
			}
			bgSeq.logHeal(madmin.HealItemObject)
		}
		trackers.setBucket(bucket)
		trackers.setObject(entry.name)
		// if time.Since(trackers.getLastUpdate()) > time.Minute {
		logger.LogIf(ctx, trackers.update(ctx))
		// }

		// Wait and proceed if there are active requests
		waitForLowHTTPReq()
	}

	// How to resolve partial results.
	resolver := metadataResolutionParams{
		dirQuorum: 1,
		objQuorum: 1,
		bucket:    bucket,
	}

	return listPathRaw(ctx, listPathRawOptions{
		disks:          disks,
		bucket:         bucket,
		recursive:      true,
		forwardTo:      startObject,
		endAt:          endObject,
		minDisks:       1,
		reportNotFound: false,
		agreed:         healEntry,
		partial: func(entries metaCacheEntries, nAgreed int, errs []error) {
			entry, ok := entries.resolve(&resolver)
			if !ok {
				// check if we can get one entry atleast
				// proceed to heal nonetheless.
				entry, _ = entries.firstFound()
			}
			healEntry(*entry)
		},
		finished: nil,
	})
}

type healingTrackers []*healingTracker

func (ht healingTrackers) isHealed(bucket string) bool {
	for _, h := range ht {
		if !h.isHealed(bucket) {
			return false
		}
	}
	return true
}

func (ht healingTrackers) bucketDone(bucket string) bool {
	for _, h := range ht {
		if h.UntilBucket != "" {
			continue
		}
		h.bucketDone(bucket)
	}
	return true
}

func (ht healingTrackers) cleanup(ctx context.Context) {
	for _, h := range ht {
		if len(h.HealedBuckets) > 0 && len(h.QueuedBuckets) == 0 {
			fmt.Println("remove", h.disk)
			h.delete(ctx)
		}
	}
}

func (ht healingTrackers) update(ctx context.Context) (err error) {
	for _, h := range ht {
		logger.LogIf(ctx, h.update(ctx))
	}
	return
}

func (ht healingTrackers) setLastUpdate(t time.Time) {
	for _, h := range ht {
		h.LastUpdate = t
	}
}

func (ht healingTrackers) getLastUpdate() time.Time {
	return ht[0].LastUpdate
}

func (ht healingTrackers) setObject(obj string) {
	for _, h := range ht {
		h.Object = obj
	}
}

func (ht healingTrackers) setBucket(bucket string) {
	for _, h := range ht {
		h.Bucket = bucket
	}
}

func (ht healingTrackers) getSetIndex() int {
	return ht[0].SetIndex
}

func (ht healingTrackers) incItemsFailed() {
	for _, h := range ht {
		h.ItemsFailed++
	}
}

func (ht healingTrackers) incItemsHealed() {
	for _, h := range ht {
		h.ItemsHealed++
	}
}

func (ht healingTrackers) incBytesFailed(bytes uint64) {
	for _, h := range ht {
		h.BytesFailed++
	}
}

func (ht healingTrackers) incBytesDone(bytes uint64) {
	for _, h := range ht {
		h.BytesDone++
	}
}

func (ht healingTrackers) updateStartAt(ctx context.Context, bucket, object string) {
	if bucket == "" || object == "" {
		return
	}
	fmt.Println("update until at", bucket, object)
	for _, h := range ht {
		if h.Bucket != "" || h.UntilBucket != "" {
			continue
		}

		fmt.Println("set middle to", h.Endpoint)
		h.UntilBucket = bucket
		h.UntilObject = object
	}

	ht.update(ctx)
}

// healObject heals given object path in deep to fix bitrot.
func healObject(bucket, object, versionID string, scan madmin.HealScanMode) {
	// Get background heal sequence to send elements to heal
	bgSeq, ok := globalBackgroundHealState.getHealSequenceByToken(bgHealingUUID)
	if ok {
		bgSeq.queueHealTask(healSource{
			bucket:    bucket,
			object:    object,
			versionID: versionID,
			opts: &madmin.HealOpts{
				Remove:   healDeleteDangling, // if found dangling purge it.
				ScanMode: scan,
			},
		}, madmin.HealItemObject)
	}
}
