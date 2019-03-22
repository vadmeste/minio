/*
 * Minio Cloud Storage, (C) 2019 Minio, Inc.
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

func init() {
	go crawlHealingOp()
}

var bgHealingUUID = "0000-0000-0000-0000"

// NewBgHealSequence
func newBgHealSequence(numDisks int) *healSequence {

	reqInfo := &logger.ReqInfo{API: "BackgroundHeal"}
	ctx := logger.SetReqInfo(context.Background(), reqInfo)

	hs := madmin.HealOpts{
		Recursive: true,
		Remove:    true,
		ScanMode:  madmin.HealDeepScan,
	}

	return &healSequence{
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

func crawlHealingOp() {
	var lastCrawl = time.Time{}
	var objectAPI ObjectLayer

	var ctx = context.Background()

	//
	for {
		objectAPI = newObjectLayerFn()
		if objectAPI == nil {
			time.Sleep(time.Second)
			continue
		}
	}

	// find number of disks in the setup
	info := objectAPI.StorageInfo(ctx)
	numDisks := info.Backend.OfflineDisks + info.Backend.OnlineDisks

	//
	for {
		if time.Now().Sub(lastCrawl) < 24*time.Hour {
			time.Sleep(time.Hour)
			continue
		}
		nh := newBgHealSequence(numDisks)
		globalAllHealState.LaunchNewHealSequence(nh)

		lastCrawl = time.Now()
	}
}
