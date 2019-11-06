/*
 * MinIO Cloud Storage, (C) 2018 MinIO, Inc.
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
)

func getObjectLayerInfo() ObjectLayerInfo {
	globalObjectLayerInfoMu.RLock()
	defer globalObjectLayerInfoMu.RUnlock()

	if globalObjectLayerInfo == nil {
		return ObjectLayerInfo{}
	}

	return *globalObjectLayerInfo
}

func initObjLayerStats() {
	go runObjLayerStatsRoutine()
}

func runObjLayerStatsRoutine() {
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

	// Hold a lock so only one server performs auto-healing
	leaderLock := objAPI.NewNSLock(ctx, minioMetaBucket, "leader-disk-stats")
	for {
		err := leaderLock.GetLock(leaderLockTimeout)
		if err == nil {
			break
		}
		time.Sleep(leaderTick)
	}

	doObjLayerStatsRoutine(ctx, objAPI)
}

func doObjLayerStatsRoutine(ctx context.Context, objAPI ObjectLayer) {
	for {
		info := objAPI.ObjectLayerInfo(ctx)

		globalObjectLayerInfoMu.Lock()
		globalObjectLayerInfo = &info
		globalObjectLayerInfoMu.Unlock()

		time.Sleep(1 * time.Minute) // should be 24 hours
	}
}
