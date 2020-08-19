/*
 * MinIO Cloud Storage, (C) 2017-2020 MinIO, Inc.
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
	"sync"
)

// A Group is a collection of goroutines working on subtasks that are part of
// the same overall task.
//
// A zero Group is valid and does not cancel on error.
type Group struct {
	errs []error

	min     int
	total   int
	done    int
	success int

	quitting bool
	cond     *sync.Cond
	mu       sync.Mutex
}

type GroupOpts struct {
	min   int
	nerrs int
}

func WithOpts(opts GroupOpts) *Group {
	g := &Group{
		errs: make([]error, opts.nerrs),
		min:  opts.min,
		cond: sync.NewCond(&sync.Mutex{}),
	}
	g.cond.L.Lock()
	return g
}

// WithNErrs returns a new Group with length of errs slice upto nerrs,
// upon Wait() errors are returned collected from all tasks.
func WithNErrs(nerrs int) *Group {
	return WithOpts(GroupOpts{nerrs: nerrs, min: nerrs})
}

// Wait blocks until all function calls from the Go method have returned, then
// returns the slice of errors from all function calls.
func (g *Group) Wait() []error {
	for {
		if g.done >= g.total {
			break
		}
		if g.success >= g.min {
			break
		}
		g.cond.Wait()
	}

	g.quitting = true
	g.cond.L.Unlock()
	return g.errs
}

// Go calls the given function in a new goroutine.
//
// The first call to return a non-nil error will be
// collected in errs slice and returned by Wait().
func (g *Group) Go(f func() error, index int) {
	if g.quitting {
		return
	}

	g.total++
	go func() {
		err := f()
		// Changes must be made when cond.L is locked.
		g.cond.L.Lock()
		g.done++
		if err == nil {
			g.success++
		} else {
			g.errs[index] = err
		}

		// Notify when cond.L lock is acquired.
		g.cond.Broadcast()
		g.cond.L.Unlock()

	}()
}
