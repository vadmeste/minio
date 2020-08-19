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
	"errors"
	"time"
)

// A Group is a collection of goroutines working on subtasks that are part of
// the same overall task.
//
// A zero Group is valid and does not cancel on error.
type Group struct {
	errs []error

	total      int
	updateCh   chan time.Duration
	failFactor int

	quitting bool
}

type GroupOpts struct {
	Min   int
	NErrs int
}

var ErrTaskAborted = errors.New("task aborted")

// WithNErrs returns a new Group with length of errs slice upto nerrs,
// upon Wait() errors are returned collected from all tasks.
func WithNErrs(nerrs int) *Group {
	return New(nerrs, 0)

}

func New(nerrs, failFactor int) *Group {
	g := &Group{
		errs:       make([]error, nerrs),
		failFactor: failFactor,
		updateCh:   make(chan time.Duration),
	}
	return g
}

// Wait blocks until all function calls from the Go method have returned, then
// returns the slice of errors from all function calls.
func (g *Group) Wait() []error {
	maxDuration := time.Duration(0)
	totalGoRoutines := 0

	var abortTimer <-chan time.Time

	// Wait for at least N/2 functions to return
	// to calculate the median execution time and
	// abort everything if trigger fail factor is
	// reached
	for {
		if g.failFactor > 0 {
			waitBeforeAbort := 5 * time.Minute
			if totalGoRoutines >= g.total/2+1 {
				waitBeforeAbort = maxDuration * time.Duration(g.failFactor)
			}
			abortTimer = time.NewTimer(waitBeforeAbort).C
		}

		select {
		case d := <-g.updateCh:
			totalGoRoutines++
			if totalGoRoutines == g.total {
				goto quit
			}
			if d > maxDuration {
				maxDuration = d
			}
		case <-abortTimer:
			goto quit
		}
	}

quit:
	g.quitting = true
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
		start := time.Now()
		err := f()
		callDuration := time.Now().Sub(start)
		if err != nil {
			g.errs[index] = err
		}
		g.updateCh <- callDuration
	}()
}
