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
)

// A Group is a collection of goroutines working on subtasks that are part of
// the same overall task.
//
// A zero Group is valid and does not cancel on error.
type Group struct {
	errs   []error
	doneCh chan struct{}
	total  int64
}

var ErrTaskIgnored = errors.New("task ignored")

// WithNErrs returns a new Group with length of errs slice upto nerrs,
// upon Wait() errors are returned collected from all tasks.
func WithNErrs(nerrs int) *Group {
	errs := make([]error, nerrs)
	for i := range errs {
		errs[i] = ErrTaskIgnored
	}
	return &Group{errs: errs, doneCh: make(chan struct{})}
}

// Wait blocks until all function calls from the Go method have returned, then
// returns the slice of errors from all function calls.
func (g *Group) Wait() []error {
	if atomic.LoadInt64(&g.total) == 0 {
		return g.errs
	}

	var done int64
	for range g.doneCh {
		done++
		if done == atomic.LoadInt64(&g.total) {
			break
		}
	}
	return g.errs
}

// Go calls the given function in a new goroutine.
//
// The first call to return a non-nil error will be
// collected in errs slice and returned by Wait().
func (g *Group) Go(f func() error, index int) {
	atomic.AddInt64(&g.total, 1)
	go func() {
		g.errs[index] = f()
		g.doneCh <- struct{}{}
	}()
}
