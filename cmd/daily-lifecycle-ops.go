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
	"time"

	"github.com/minio/minio/pkg/lifecycle"
)

const (
	bgLifecycleInterval = 24 * time.Hour
)

type lifecycleListener struct {
	ch           chan sweepEntry
	lastActivity time.Time
}

func (l *lifecycleListener) Send(elem sweepEntry) {
	l.ch <- elem
}

func (l *lifecycleListener) SignalEnd() {
	l.lastActivity = time.Now()
}

func (l *lifecycleListener) Interested(bucketName string) bool {
	var states = []bgOpsStatus{
		bgOpsStatus{
			LifecycleOps: getLocalBgLifecycleOpsStatus(),
		},
	}
	if globalIsDistXL {
		peerStates := globalNotificationSys.BackgroundOpsStatus()
		states = append(states, peerStates...)
	}

	var lastActivity time.Time
	for _, state := range states {
		if state.LifecycleOps.LastActivity.After(lastActivity) {
			lastActivity = state.LifecycleOps.LastActivity
		}
	}

	if time.Since(lastActivity) < bgLifecycleInterval {
		return false
	}

	_, ok := globalLifecycleSys.Get(bucketName)
	return ok
}

// Register to the daily objects listing
var globalLifecycleListener = &lifecycleListener{ch: make(chan sweepEntry)}

func getLocalBgLifecycleOpsStatus() bgLifecycleOpsStatus {
	return bgLifecycleOpsStatus{
		LastActivity: globalLifecycleListener.lastActivity,
	}
}

// initDailyLifecycle starts the routine that receives the daily
// listing of all objects and applies any matching bucket lifecycle
// rules.
func initDailyLifecycle() {
	go startDailyLifecycle()
}

func startDailyLifecycle() {
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

	registerDailySweepListener(globalLifecycleListener)

	// Iterate over all received objects, find any matching lifecycle rules
	// and apply it
	for elem := range globalLifecycleListener.ch {
		// Ignore if bucket or object name is empty
		if elem.bucket.Name == "" || elem.object.Name == "" {
			continue
		}

		// Find the current bucket lifecycle
		l, ok := globalLifecycleSys.Get(elem.bucket.Name)
		if ok {
			// Find the action that need to be executed
			action := l.ComputeAction(elem.object.Name, elem.object.ModTime)
			switch action {
			case lifecycle.DeleteAction:
				objAPI.DeleteObject(ctx, elem.bucket.Name, elem.object.Name)
			default:
				// Nothing

			}
		}
	}
}
