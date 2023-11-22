// Copyright (c) 2015-2023 MinIO, Inc.
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
	"sync"
	"time"
)

const (
	// the interval to discover if there are new buckets created in the cluster
	bucketsListInterval = time.Minute
)

type bucketScanStat struct {
	ongoing   bool   // this bucket is currently being scanned
	cycle     uint32 // the last cycle of this scan
	lifecycle bool   // This bucket has lifecycle set
}

type bucketsScanMgr struct {
	ctx context.Context

	// A registered function which knows how to list buckets
	bucketsLister func(context.Context, BucketOptions) ([]BucketInfo, error)

	mu           sync.RWMutex
	knownBuckets map[string]struct{}               // Current buckets in the S3 namespace
	bucketsCh    map[int]chan string               // A map of an erasure set identifier and a channel of buckets to scan
	internal     map[int]map[string]bucketScanStat // A map of an erasure set identifier and bucket scan stats
}

func newBucketsScanMgr(s3 ObjectLayer) *bucketsScanMgr {
	mgr := &bucketsScanMgr{
		ctx:           GlobalContext,
		bucketsLister: s3.ListBuckets,
		internal:      make(map[int]map[string]bucketScanStat),
		bucketsCh:     make(map[int]chan string),
	}
	return mgr
}

func (mgr *bucketsScanMgr) getKnownBuckets() []string {
	mgr.mu.Lock()
	defer mgr.mu.Unlock()

	ret := make([]string, 0, len(mgr.knownBuckets))
	for k := range mgr.knownBuckets {
		ret = append(ret, k)
	}
	return ret
}

func (mgr *bucketsScanMgr) isKnownBucket(bucket string) bool {
	mgr.mu.Lock()
	defer mgr.mu.Unlock()

	_, ok := mgr.knownBuckets[bucket]
	return ok
}

func (mgr *bucketsScanMgr) start() {
	// A routine that discovers new buckets and initialize scan stats for each new bucket
	go func() {
		t := time.NewTimer(bucketsListInterval)
		defer t.Stop()

		for {
			select {
			case <-t.C:
				buckets, err := mgr.bucketsLister(mgr.ctx, BucketOptions{})
				if err == nil {
					mgr.mu.Lock()
					mgr.knownBuckets = make(map[string]struct{}, len(buckets))
					for _, b := range buckets {
						mgr.knownBuckets[b.Name] = struct{}{}
					}
					for bucket := range mgr.knownBuckets {
						for _, set := range mgr.internal {
							st := set[bucket]
							if l, err := globalLifecycleSys.Get(bucket); err == nil && l.HasActiveRules("") {
								st.lifecycle = true
							}
							set[bucket] = st
						}
					}
					mgr.mu.Unlock()
				}
				t.Reset(bucketsListInterval)
			case <-mgr.ctx.Done():
				return
			}
		}
	}()

	// Clean up internal data when a deleted bucket is found
	go func() {
		const cleanInterval = 30 * time.Second

		t := time.NewTimer(cleanInterval)
		defer t.Stop()

		for {
			select {
			case <-t.C:
				mgr.mu.Lock()
				for _, set := range mgr.internal {
					for bkt := range set {
						if _, ok := mgr.knownBuckets[bkt]; !ok {
							delete(set, bkt)
						}
					}
				}
				mgr.mu.Unlock()

				t.Reset(cleanInterval)
			case <-mgr.ctx.Done():
				return
			}
		}
	}()

	// A routine that sends the next bucket to scan for each erasure set listener
	go func() {
		tick := 10 * time.Second

		t := time.NewTimer(tick)
		defer t.Stop()

		for {
			select {
			case <-t.C:
				mgr.mu.RLock()
				for id, ch := range mgr.bucketsCh {
					if len(ch) == 0 {
						b := mgr.unsafeGetNextBucket(id)
						if b != "" {
							select {
							case ch <- b:
							default:
							}
						}
					}
				}
				mgr.mu.RUnlock()

				t.Reset(tick)
			case <-mgr.ctx.Done():
				return
			}
		}
	}()
}

// Return a channel of buckets names to scan a given erasure set identifier
func (mgr *bucketsScanMgr) getBucketCh(id int) chan string {
	mgr.mu.Lock()
	defer mgr.mu.Unlock()

	mgr.internal[id] = make(map[string]bucketScanStat)
	mgr.bucketsCh[id] = make(chan string, 1)

	return mgr.bucketsCh[id]
}

func less(st1, st2 bucketScanStat) bool {
	if st1.ongoing != st2.ongoing {
		return st1.ongoing == false
	}
	if st1.cycle != st2.cycle {
		return st1.cycle < st2.cycle
	}
	if st1.lifecycle != st2.lifecycle {
		return st1.lifecycle == true
	}
	return false
}

// Return the next bucket name to scan of a given erasure set identifier
// If all buckets are in a scanning state, return empty result
func (mgr *bucketsScanMgr) unsafeGetNextBucket(id int) string {
	var (
		nextBucketStat = bucketScanStat{}
		nextBucketName = ""
	)

	for bucket, stat := range mgr.internal[id] {
		if stat.ongoing {
			continue
		}
		if nextBucketName == "" {
			nextBucketName = bucket
			nextBucketStat = stat
			continue
		}
		if nextBucketName == "" || less(stat, nextBucketStat) {
			nextBucketStat = stat
			nextBucketName = bucket
		}
	}

	return nextBucketName
}

// Mark a bucket as done in a specific erasure set - returns true if successful,
// false if the bucket is already in a scanning phase
func (mgr *bucketsScanMgr) markBucketScanStarted(id int, bucket string, cycle uint32) bool {
	mgr.mu.Lock()
	defer mgr.mu.Unlock()

	m, _ := mgr.internal[id][bucket]
	if m.ongoing {
		return false
	}

	m.ongoing = true
	m.cycle = cycle
	mgr.internal[id][bucket] = m
	return true
}

// Mark a bucket as done in a specific erasure set
func (mgr *bucketsScanMgr) markBucketScanDone(id int, bucket string) {
	mgr.mu.Lock()
	defer mgr.mu.Unlock()

	m, _ := mgr.internal[id][bucket]
	m.ongoing = false
	mgr.internal[id][bucket] = m
}
