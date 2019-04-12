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
	"time"

	"github.com/minio/minio/cmd/logger"
	"github.com/minio/minio/pkg/madmin"
)

// healTask represents what to heal along with options
//   path: '/' =>  Heal disk formats along with metadata
//   path: 'bucket/' or '/bucket/' => Heal bucket
//   path: 'bucket/object' => Heal object
type healTask struct {
	path string
	opts madmin.HealOpts
	// Healing response will be sent here
	responseCh chan healResult
}

// healResult represents a healing result with a possible error
type healResult struct {
	result madmin.HealResultItem
	err    error
}

type healRoutine struct {
	tasks chan healTask
}

func (h *healRoutine) queueHealTask(task healTask) {
	h.tasks <- task
}

func (h *healRoutine) run() {
	ctx := context.Background()
	for task := range h.tasks {

		if globalHTTPServer != nil {
			// Wait at max 1 minute for an inprogress request
			// before proceeding to heal
			waitCount := 60
			// Any requests in progress, delay the heal.
			for globalHTTPServer.GetRequestCount() > 2 && waitCount > 0 {
				waitCount--
				time.Sleep(1 * time.Second)
			}
		}

		var res madmin.HealResultItem
		var err error
		bucket, object := urlPath2BucketObjectName(task.path)
		switch {
		case bucket == "" && object == "":
			res, err = bgHealDiskFormat(ctx, task.opts)
		case bucket != "" && object == "":
			res, err = bgHealBucket(ctx, bucket, task.opts)
		case bucket != "" && object != "":
			res, err = bgHealObject(ctx, bucket, object, task.opts)
		}
		task.responseCh <- healResult{result: res, err: err}
	}
}

func initBackgroundHealing() {
	healBg := &healRoutine{
		tasks: make(chan healTask),
	}
	go healBg.run()

	globalBackgroundHealing = healBg
}

// bgHealDiskFormat - heals format.json, return value indicates if a
// failure error occurred.
func bgHealDiskFormat(ctx context.Context, opts madmin.HealOpts) (madmin.HealResultItem, error) {
	// Get current object layer instance.
	objectAPI := newObjectLayerFn()
	if objectAPI == nil {
		return madmin.HealResultItem{}, errServerNotInitialized
	}

	res, err := objectAPI.HealFormat(ctx, opts.DryRun)

	// return any error, ignore error returned when disks have
	// already healed.
	if err != nil && err != errNoHealRequired {
		return madmin.HealResultItem{}, err
	}

	// Healing succeeded notify the peers to reload format and re-initialize disks.
	// We will not notify peers only if healing succeeded.
	if err == nil {
		for _, nerr := range globalNotificationSys.ReloadFormat(opts.DryRun) {
			if nerr.Err != nil {
				logger.GetReqInfo(ctx).SetTags("peerAddress", nerr.Host.String())
				logger.LogIf(ctx, nerr.Err)
			}
		}
	}

	return res, nil
}

// bghealBucket - traverses and heals given bucket
func bgHealBucket(ctx context.Context, bucket string, opts madmin.HealOpts) (madmin.HealResultItem, error) {
	// Get current object layer instance.
	objectAPI := newObjectLayerFn()
	if objectAPI == nil {
		return madmin.HealResultItem{}, errServerNotInitialized
	}

	return objectAPI.HealBucket(ctx, bucket, opts.DryRun, opts.Remove)
}

// bgHealObject - heal the given object and record result
func bgHealObject(ctx context.Context, bucket, object string, opts madmin.HealOpts) (madmin.HealResultItem, error) {
	// Get current object layer instance.
	objectAPI := newObjectLayerFn()
	if objectAPI == nil {
		return madmin.HealResultItem{}, errServerNotInitialized
	}
	return objectAPI.HealObject(ctx, bucket, object, opts.DryRun, opts.Remove, opts.ScanMode)
}
