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
)

var globalDailySweepListeners = make([]chan string, 0)
var globalDailySweepListenersMu = sync.Mutex{}

func registerDailySweepListener(ch chan string) {
	globalDailySweepListenersMu.Lock()
	defer globalDailySweepListenersMu.Unlock()

	globalDailySweepListeners = append(globalDailySweepListeners, ch)
}

func copyDailySweepListeners() []chan string {
	globalDailySweepListenersMu.Lock()
	defer globalDailySweepListenersMu.Unlock()

	var listenersCopy = make([]chan string, len(globalDailySweepListeners))
	copy(listenersCopy, globalDailySweepListeners)

	return listenersCopy
}

func sweepRound(ctx context.Context, objAPI ObjectLayer) error {

	zeroDuration := time.Duration(0)
	zeroDynamicTimeout := newDynamicTimeout(zeroDuration, zeroDuration)

	// Lock the object before reading.
	sweepLock := globalNSMutex.NewNSLock("system", "daily-sweep")
	if err := sweepLock.GetLock(zeroDynamicTimeout); err != nil {
		return err
	}
	defer sweepLock.Unlock()

	buckets, err := objAPI.ListBuckets(ctx)
	if err != nil {
		return err
	}

	for _, bucket := range buckets {
		marker := ""
		for {
			res, err := objAPI.ListObjects(ctx, bucket.Name, "", marker, "", 1000)
			if err != nil {
				continue
			}
			for _, obj := range res.Objects {
				for _, l := range copyDailySweepListeners() {
					l <- pathJoin(bucket.Name, obj.Name)
				}
			}
			if !res.IsTruncated {
				break
			} else {
				marker = res.NextMarker
			}
		}
	}

	return nil
}

func initDailySweeper() {
	go dailySweeper()
}

func dailySweeper() {
	var lastSweepTime time.Time
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

	for {
		if time.Since(lastSweepTime) < 1*time.Minute {
			time.Sleep(time.Second)
			continue
		}

		err := sweepRound(ctx, objAPI)
		if err != nil {
			switch err.(type) {
			case OperationTimedOut:
				lastSweepTime = time.Now()
			default:
				logger.LogIf(ctx, err)
				time.Sleep(time.Minute)
				continue
			}
		} else {
			lastSweepTime = time.Now()
		}
	}
}
