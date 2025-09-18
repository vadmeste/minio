package cmd

import (
	"context"
	"sync"
	"time"
)

// delayedCancelContext wraps another context and delays its cancellation
type delayedCancelContext struct {
	parent context.Context
	done   chan struct{}
	err    error
	mu     sync.RWMutex
	values map[interface{}]interface{}
	delay  time.Duration
	once   sync.Once
}

// WithDelayedCancel creates a new context that will be canceled one minute
// after the parent context is canceled
func WithDelayedCancel(parent context.Context, delay time.Duration) (context.Context, context.CancelFunc) {
	ctx := &delayedCancelContext{
		parent: parent,
		done:   make(chan struct{}),
		values: make(map[interface{}]interface{}),
		delay:  delay,
	}

	// Set up manual cancel function
	manualCancel := func() {
		ctx.once.Do(func() {
			ctx.mu.Lock()
			ctx.err = context.Canceled
			close(ctx.done)
			ctx.mu.Unlock()
		})
	}

	// Start goroutine to monitor parent context
	go func() {
		select {
		case <-parent.Done():
			timer := time.NewTimer(delay)
			defer timer.Stop()

			select {
			case <-timer.C:
				ctx.once.Do(func() {
					ctx.mu.Lock()
					ctx.err = parent.Err() // Inherit the parent's error
					close(ctx.done)
					ctx.mu.Unlock()
				})
			case <-ctx.done:
				// Context was manually canceled before the delay expired
				return
			}
		case <-ctx.done:
			// Context was manually canceled
			return
		}
	}()

	return ctx, manualCancel
}

// Implement the Context interface
func (c *delayedCancelContext) Deadline() (deadline time.Time, ok bool) {
	// If parent has deadline, add one minute to account for the delay
	if parentDeadline, hasDeadline := c.parent.Deadline(); hasDeadline {
		return parentDeadline.Add(c.delay), true
	}
	return time.Time{}, false
}

func (c *delayedCancelContext) Done() <-chan struct{} {
	return c.done
}

func (c *delayedCancelContext) Err() error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.err
}

func (c *delayedCancelContext) Value(key interface{}) interface{} {
	c.mu.RLock()
	if val, exists := c.values[key]; exists {
		c.mu.RUnlock()
		return val
	}
	c.mu.RUnlock()

	// Delegate to parent for inherited values
	return c.parent.Value(key)
}

// SetValue allows setting values on the delayed context (optional utility)
func (c *delayedCancelContext) SetValue(key, value interface{}) {
	c.mu.Lock()
	c.values[key] = value
	c.mu.Unlock()
}
