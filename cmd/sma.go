/*
 * MinIO Cloud Storage, (C) 2020 MinIO, Inc.
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

//** Inspired from https://github.com/mxmCherry/movavg

package cmd

import (
	"sync"
)

type movingAverage struct {
	mu sync.Mutex

	window int // Simple Moving Average window (number of latest values to calculate average value)

	vals []float64 // stored values, circular list
	n    int       // number of actually stored values (vals slice usage)
	i    int       // last-set value index; first-set (oldest) value is (i+1)%window when n == window

	avg float64 // current Simple Moving Average value
}

// NewMovingAverage constructs new Simple Moving Average calculator.
// Window arg must be >= 1.
func newMovingAverage(window int) *movingAverage {
	return &movingAverage{
		window: window,
		vals:   make([]float64, window),
	}
}

// Add recalculates Simple Moving Average value and returns it.
func (a *movingAverage) Add(v float64) float64 {
	a.mu.Lock()
	defer a.mu.Unlock()

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
	return a.avg
}

// Avg returns current Simple Moving Average value.
func (a *movingAverage) Avg() float64 {
	a.mu.Lock()
	defer a.mu.Unlock()

	return a.avg
}
