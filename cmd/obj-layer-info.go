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
	"fmt"
	"time"
)

const (
	objectLayerInfoUpdateInterval = 24 * time.Hour
)

// Safe copy of the global object layer info
func safeCopyObjectLayerInfo() map[int]ObjectLayerInfo {
	globalObjectLayerInfoMu.RLock()
	defer globalObjectLayerInfoMu.RUnlock()

	var objInfoCopy = make(map[int]ObjectLayerInfo)
	for setIndex, info := range globalObjectLayerInfo {
		newObjInfo := ObjectLayerInfo{}
		newObjInfo.ObjectsCount = info.ObjectsCount
		newObjInfo.TotalSize = info.TotalSize
		newObjInfo.BucketsCount = info.BucketsCount
		newObjInfo.LastUpdate = info.LastUpdate
		newObjInfo.ObjectsSizesHistogram = make(map[string]uint64)
		for k, v := range info.ObjectsSizesHistogram {
			newObjInfo.ObjectsSizesHistogram[k] = v
		}
		newObjInfo.BucketsSizes = make(map[string]uint64)
		for k, v := range info.BucketsSizes {
			newObjInfo.BucketsSizes[k] = v
		}
		objInfoCopy[setIndex] = newObjInfo
	}
	return objInfoCopy
}

// Safe update of the object layer info
func safeUpdateObjectLayerInfo(setIndex int, info ObjectLayerInfo) {
	newObjInfo := ObjectLayerInfo{}
	newObjInfo.ObjectsCount = info.ObjectsCount
	newObjInfo.TotalSize = info.TotalSize
	newObjInfo.BucketsCount = info.BucketsCount
	newObjInfo.LastUpdate = info.LastUpdate
	newObjInfo.ObjectsSizesHistogram = make(map[string]uint64)
	for k, v := range info.ObjectsSizesHistogram {
		newObjInfo.ObjectsSizesHistogram[k] = v
	}
	newObjInfo.BucketsSizes = make(map[string]uint64)
	for k, v := range info.BucketsSizes {
		newObjInfo.BucketsSizes[k] = v
	}

	globalObjectLayerInfoMu.Lock()
	defer globalObjectLayerInfoMu.Unlock()

	if globalObjectLayerInfo == nil {
		globalObjectLayerInfo = make(map[int]ObjectLayerInfo)
	}
	globalObjectLayerInfo[setIndex] = newObjInfo
}

func initObjLayerStats() {
	go runObjLayerInfoUpdateRoutine()
}

func runObjLayerInfoUpdateRoutine() {
	// Wait until the object layer is ready
	var objAPI ObjectLayer
	for {
		objAPI = newObjectLayerWithoutSafeModeFn()
		if objAPI == nil {
			time.Sleep(time.Second)
			continue
		}
		break
	}

	ctx := context.Background()
	for {
		updateObjectLayerInfo(ctx, objAPI)
		time.Sleep(objectLayerInfoUpdateInterval)
	}
}

func updateObjectLayerInfo(ctx context.Context, objAPI ObjectLayer) {
	allSets, ok := objAPI.(*xlSets)
	if !ok {
		// FS mode (or gateway in the futre)
		objInfo := objAPI.ObjectLayerInfo(ctx)
		if !objInfo.LastUpdate.IsZero() {
			safeUpdateObjectLayerInfo(0, objInfo)
		}
		return
	}

	for idx, xlObj := range allSets.sets {
		disks := xlObj.getDisks()
		for _, d := range disks {
			if d == nil {
				continue
			}
			if !d.IsLocal() || !d.IsOnline() {
				continue
			}

			locker := allSets.NewNSLock(ctx, minioMetaBucket, fmt.Sprintf("obj-info-%d", idx))
			err := locker.GetLock(newDynamicTimeout(time.Millisecond, time.Millisecond))
			if err != nil {
				continue
			}

			objInfo := xlObj.ObjectLayerInfo(ctx)
			if !objInfo.LastUpdate.IsZero() {
				safeUpdateObjectLayerInfo(idx, objInfo)
			} else {
				// Go to the next disk
				locker.Unlock()
				continue
			}

			locker.Unlock()
			// Go to the next set
			break
		}
	}
}
