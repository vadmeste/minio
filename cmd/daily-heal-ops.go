/*
 * MinIO Cloud Storage, (C) 2019 MinIO, Inc.
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
	"sync"
	"time"

	"github.com/minio/minio/cmd/logger"
	"github.com/minio/minio/pkg/madmin"
)

const (
	bgHealingUUID     = "0000-0000-0000-0000"
	bgHealingInterval = 30 * 24 * time.Hour
)

type healListener struct {
	ch           chan sweepEntry
	lastActivity time.Time
}

func (h *healListener) Send(elem sweepEntry) {
	h.ch <- elem
}

func (h *healListener) SignalEnd() {
	h.lastActivity = time.Now()
}

func (h *healListener) Interested(bucketName string) bool {
	var states = []madmin.BgHealState{getLocalBackgroundHealStatus()}
	if globalIsDistXL {
		peerStates := globalNotificationSys.BackgroundHealStatus()
		states = append(states, peerStates...)
	}

	var lastActivity time.Time
	for _, state := range states {
		if state.LastHealActivity.After(lastActivity) {
			lastActivity = state.LastHealActivity
		}
	}

	return time.Since(lastActivity) > bgHealingInterval
}

// NewBgHealSequence creates a background healing sequence
// operation which crawls all objects and heal them.
func newBgHealSequence(numDisks int) *healSequence {

	reqInfo := &logger.ReqInfo{API: "BackgroundHeal"}
	ctx := logger.SetReqInfo(context.Background(), reqInfo)

	hs := madmin.HealOpts{
		// Remove objects that do not have read-quorum
		Remove:   true,
		ScanMode: madmin.HealNormalScan,
	}

	return &healSequence{
		sourceCh:    make(chan sweepEntry),
		startTime:   UTCNow(),
		clientToken: bgHealingUUID,
		settings:    hs,
		currentStatus: healSequenceStatus{
			Summary:      healNotStartedStatus,
			HealSettings: hs,
			NumDisks:     numDisks,
			updateLock:   &sync.RWMutex{},
		},
		traverseAndHealDoneCh: make(chan error),
		stopSignalCh:          make(chan struct{}),
		ctx:                   ctx,
		reportProgress:        false,
	}
}

func getLocalBackgroundHealStatus() madmin.BgHealState {
	backgroundSequence, ok := globalSweepHealState.getHealSequenceByToken(bgHealingUUID)
	if !ok {
		return madmin.BgHealState{}
	}

	return madmin.BgHealState{
		ScannedItemsCount: backgroundSequence.scannedItemsCount,
		LastHealActivity:  backgroundSequence.lastHealActivity,
	}
}

func initDailyHeal() {
	go startDailyHeal()
}

func startDailyHeal() {
	var objAPI ObjectLayer
	var ctx = context.Background()

	// Wait until the object API is ready
	for {
		objAPI = newObjectLayerFn()
		if objAPI == nil {
			time.Sleep(time.Second)
			continue
		}
		break
	}

	// Find number of disks in the setup
	info := objAPI.StorageInfo(ctx)
	numDisks := info.Backend.OnlineDisks + info.Backend.OfflineDisks

	nh := newBgHealSequence(numDisks)
	globalSweepHealState.LaunchNewHealSequence(nh)

	l := &healListener{ch: nh.sourceCh}
	registerDailySweepListener(l)
}
