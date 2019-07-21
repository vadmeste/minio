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
)

// General objects listing operation sends to listener
// a sweepEntry for each bucket and object
type sweepEntry struct {
	bucket BucketInfo
	object ObjectInfo
}

// Listeners to objects listing needs to satisfy
// sweepListener
type sweepListener interface {
	// Receive a new bucket/object entry
	Send(sweepEntry)
	// Interested to receive objects listing
	Interested(string) bool
	// Signal the start of the current sweep operation
	SignalStart()
}

// The list of modules listening for the daily listing of all objects
// such as the daily heal ops, disk usage and bucket lifecycle management.
var globalDailySweepListeners = make([]sweepListener, 0)
var globalDailySweepListenersMu = sync.Mutex{}

// Add a new listener to the daily objects listing
func registerDailySweepListener(listener sweepListener) {
	globalDailySweepListenersMu.Lock()
	defer globalDailySweepListenersMu.Unlock()

	globalDailySweepListeners = append(globalDailySweepListeners, listener)
}

// Safe copy of globalDailySweepListeners content
func copyDailySweepListeners() []sweepListener {
	globalDailySweepListenersMu.Lock()
	defer globalDailySweepListenersMu.Unlock()

	var listenersCopy = make([]sweepListener, len(globalDailySweepListeners))
	copy(listenersCopy, globalDailySweepListeners)

	return listenersCopy
}

// sweepRound will list all objects, having read quorum or not and
// feeds to all listeners, such as the background healing
func sweepRound(ctx context.Context, objAPI ObjectLayer) error {
	zeroDuration := time.Millisecond
	zeroDynamicTimeout := newDynamicTimeout(zeroDuration, zeroDuration)

	// General lock so we avoid parallel daily sweep by different instances.
	sweepLock := globalNSMutex.NewNSLock(ctx, "system", "daily-sweep")
	if err := sweepLock.GetLock(zeroDynamicTimeout); err != nil {
		return err
	}
	defer sweepLock.Unlock()

	buckets, err := objAPI.ListBuckets(ctx)
	if err != nil {
		return err
	}

	allListeners := copyDailySweepListeners()

	// List all objects, having read quorum or not in all buckets
	// and send them to all the registered sweep listeners
	for _, bucket := range buckets {
		var listeners []sweepListener
		for _, l := range allListeners {
			if l.Interested(bucket.Name) {
				l.SignalStart()
				listeners = append(listeners, l)
			}
		}

		if len(listeners) == 0 {
			// No listener for this bucket, move on
			continue
		}

		// Send bucket names to all listeners
		for _, l := range listeners {
			l.Send(sweepEntry{bucket: bucket})
		}

		marker := ""
		for {
			res, err := objAPI.ListObjectsAll(ctx, bucket.Name, "", marker, "", 1000)
			if err != nil {
				continue
			}
			// Send all objects to listeners
			for _, obj := range res.Objects {
				for _, l := range listeners {
					l.Send(sweepEntry{bucket: bucket, object: obj})
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

// initDailySweeper creates a go-routine which will list all
// objects in all buckets in a daily basis
func initDailySweeper() {
	go dailySweeper()
}

// List all objects in all buckets in a daily basis
func dailySweeper() {
	var lastSweepTime time.Time
	var objAPI ObjectLayer

	var ctx = context.Background()

	// Wait until the object layer is ready
	for {
		objAPI = newObjectLayerFn()
		if objAPI == nil {
			time.Sleep(time.Second)
			continue
		}
		break
	}

	// Perform a sweep round each month
	for {
		if time.Since(lastSweepTime) < 6*time.Hour {
			time.Sleep(time.Hour)
			continue
		}

		err := sweepRound(ctx, objAPI)
		if err != nil {
			switch err.(type) {
			// Unable to hold a lock means there is another
			// instance doing the sweep round
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
