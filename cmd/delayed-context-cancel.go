// MinIO, Inc. CONFIDENTIAL
//
// [2014] - [2025] MinIO, Inc. All Rights Reserved.
//
// NOTICE:  All information contained herein is, and remains the property
// of MinIO, Inc and its suppliers, if any.  The intellectual and technical
// concepts contained herein are proprietary to MinIO, Inc and its suppliers
// and may be covered by U.S. and Foreign Patents, patents in process, and are
// protected by trade secret or copyright law. Dissemination of this information
// or reproduction of this material is strictly forbidden unless prior written
// permission is obtained from MinIO, Inc.

package cmd

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

// delayedCancelContext delays cancellation after parent context is canceled
type delayedCancelContext struct {
	parent context.Context
	done   chan struct{}
	err    atomic.Pointer[error]
	delay  time.Duration
}

// WithDelayedCancel returns a context that cancels after a delay when parent cancels
func WithDelayedCancel(parent context.Context, delay time.Duration) (context.Context, context.CancelFunc) {
	ctx := &delayedCancelContext{
		parent: parent,
		done:   make(chan struct{}),
		delay:  delay,
	}

	var once sync.Once
	cancel := func() {
		once.Do(func() {
			ctx.err.Store(&context.Canceled)
			close(ctx.done)
		})
	}

	go func() {
		select {
		case <-parent.Done():
		case <-ctx.done:
			return
		}

		timer := time.NewTimer(delay)
		defer timer.Stop()

		select {
		case <-timer.C:
			once.Do(func() {
				parentErr := parent.Err()
				ctx.err.Store(&parentErr)
				close(ctx.done)
			})
		case <-ctx.done:
			// Manual cancel before delay
		}
	}()

	return ctx, cancel
}

func (c *delayedCancelContext) Deadline() (time.Time, bool) {
	if d, ok := c.parent.Deadline(); ok {
		// Add delay to parent deadline to reflect actual cancellation time
		return d.Add(c.delay), true
	}
	return time.Time{}, false
}

func (c *delayedCancelContext) Done() <-chan struct{} {
	return c.done
}

func (c *delayedCancelContext) Err() error {
	if err := c.err.Load(); err != nil {
		return *err
	}
	return nil
}

func (c *delayedCancelContext) Value(key interface{}) interface{} {
	return c.parent.Value(key)
}
