package cmd

import (
	"sync"
	"time"
)

type movingAverage struct {
	mu sync.Mutex

	window int // Simple Moving Average window (number of latest values to calculate average value)

	vals []float64 // stored values, circular list
	n    int       // number of actually stored values (vals slice usage)
	i    int       // last-set value index; first-set (oldest) value is (i+1)%window when n == window

	avg float64 // current Simple Moving Average value

	frequency  time.Duration
	lastUpdate time.Time
}

// NewMovingAverage constructs new Simple Moving Average calculator.
// Window arg must be >= 1.
func newMovingAverage(window int, frequency time.Duration) movingAverage {
	if window <= 0 {
		panic("movavg.NewSMA: window should be > 0")
	}
	if frequency == 0 {
		frequency = time.Second
	}
	return movingAverage{
		window:    window,
		frequency: frequency,
		vals:      make([]float64, window),
	}
}

// Add recalculates Simple Moving Average value and returns it.
func (a *movingAverage) Add(v float64) {
	a.mu.Lock()
	defer a.mu.Unlock()

	now := time.Now()
	if now.Sub(a.lastUpdate) < a.frequency {
		return
	}
	a.lastUpdate = now

	if a.n == a.window {
		// filled window - most frequent case:
		// https://en.wikipedia.org/wiki/Moving_average#Simple_moving_average
		a.i = (a.i + 1) % a.window
		a.avg += (v - a.vals[a.i]) / float64(a.n)
		a.vals[a.i] = v
	} else if a.n != 0 {
		// partially-filled window - second most frequent case:
		// https://en.wikipedia.org/wiki/Moving_average#Cumulative_moving_average
		a.i = (a.i + 1) % a.window
		a.avg = (v + float64(a.n)*a.avg) / float64(a.n+1)
		a.vals[a.i] = v
		a.n++
	} else {
		// empty window - least frequent case (occurs only once, on first value added):
		// simply assign given value as current average:
		a.avg = v
		a.vals[0] = v
		a.n = 1
	}

}

// Avg returns current Simple Moving Average value.
func (a *movingAverage) Avg() float64 {
	a.mu.Lock()
	defer a.mu.Unlock()

	return a.avg
}
