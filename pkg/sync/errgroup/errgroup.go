/*
 * MinIO Cloud Storage, (C) 2017 MinIO, Inc.
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

package errgroup

import (
	"errors"
	"sync/atomic"
	"time"
)

// A Group is a collection of goroutines working on subtasks that are part of
// the same overall task.
//
// A zero Group is valid and does not cancel on error.
type Group struct {
	errs     []error
	doneCh   chan time.Duration
	total    int64
	quitting int64

	failFactor int
}

var ErrTaskIgnored = errors.New("task ignored")

// New returns a new Group upon Wait() errors are returned collected from all tasks.
func New(nerrs, failFactor int) *Group {
	errs := make([]error, nerrs)
	for i := range errs {
		errs[i] = ErrTaskIgnored
	}
	return &Group{
		errs:       errs,
		doneCh:     make(chan time.Duration),
		failFactor: failFactor,
	}
}

func WithNErrs(nerrs int) *Group {
	return New(nerrs, 0)
}

// Wait blocks until all function calls from the Go method have returned, then
// returns the slice of errors from all function calls.
func (g *Group) Wait() []error {
	if atomic.LoadInt64(&g.total) == 0 {
		return g.errs
	}

	var done int64
	var maxTaskDuration time.Duration

	for {
		var abortTimer <-chan time.Time
		if g.failFactor != 0 && done >= atomic.LoadInt64(&g.total)/2+1 {
			abortTimer = time.NewTimer(maxTaskDuration * time.Duration(g.failFactor)).C
		}

		select {
		case d := <-g.doneCh:
			done++
			if done == atomic.LoadInt64(&g.total) {
				goto quit
			}
			if d > maxTaskDuration {
				maxTaskDuration = d
			}
		case <-abortTimer:
			goto quit
		}
	}

quit:
	atomic.StoreInt64(&g.quitting, 1)
	return g.errs
}

// Go calls the given function in a new goroutine.
//
// The first call to return a non-nil error will be
// collected in errs slice and returned by Wait().
func (g *Group) Go(f func() error, index int) {
	if atomic.LoadInt64(&g.quitting) > 0 {
		return
	}
	atomic.AddInt64(&g.total, 1)
	go func() {
		now := time.Now()
		g.errs[index] = f()
		d := time.Now().Sub(now)
		g.doneCh <- d
	}()
}
