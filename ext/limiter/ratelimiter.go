package limiter

/*
### Licence

```
Copyright (c) 2015 Black Square Media

Permission is hereby granted, free of charge, to any person obtaining
a copy of this software and associated documentation files (the
"Software"), to deal in the Software without restriction, including
without limitation the rights to use, copy, modify, merge, publish,
distribute, sublicense, and/or sell copies of the Software, and to
permit persons to whom the Software is furnished to do so, subject to
the following conditions:

The above copyright notice and this permission notice shall be
included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE
LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION
OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION
WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
```
*/

import (
	"sync/atomic"
	"time"
)

// RateLimiter instances are thread-safe.
type RateLimiter struct {
	rate, allowance, max, unit, lastCheck uint64
}

// New creates a new rate limiter instance
func NewRateLimiter(rate int, per time.Duration) *RateLimiter {
	nano := uint64(per)
	if nano < 1 {
		nano = uint64(time.Second)
	}
	if rate < 1 {
		rate = 1
	}

	return &RateLimiter{
		rate:      uint64(rate),        // store the rate
		allowance: uint64(rate) * nano, // set our allowance to max in the beginning
		max:       uint64(rate) * nano, // remember our maximum allowance
		unit:      nano,                // remember our unit size

		lastCheck: unixNano(),
	}
}

// UpdateRate allows to update the allowed rate
func (rl *RateLimiter) UpdateRate(rate int) {
	atomic.StoreUint64(&rl.rate, uint64(rate))
	atomic.StoreUint64(&rl.max, uint64(rate)*rl.unit)
}

// Limit returns true if rate was exceeded
func (rl *RateLimiter) Limit() bool {
	// Calculate the number of ns that have passed since our last call
	now := unixNano()
	passed := now - atomic.SwapUint64(&rl.lastCheck, now)

	// Add them to our allowance
	rate := atomic.LoadUint64(&rl.rate)
	current := atomic.AddUint64(&rl.allowance, passed*rate)

	// Ensure our allowance is not over maximum
	if max := atomic.LoadUint64(&rl.max); current > max {
		atomic.AddUint64(&rl.allowance, max-current)
		current = max
	}

	// If our allowance is less than one unit, rate-limit!
	if current < rl.unit {
		return true
	}

	// Not limited, subtract a unit
	atomic.AddUint64(&rl.allowance, -rl.unit)
	return false
}

// Undo reverts the last Limit() call, returning consumed allowance
func (rl *RateLimiter) Undo() {
	current := atomic.AddUint64(&rl.allowance, rl.unit)

	// Ensure our allowance is not over maximum
	if max := atomic.LoadUint64(&rl.max); current > max {
		atomic.AddUint64(&rl.allowance, max-current)
	}
}

// now as unix nanoseconds
func unixNano() uint64 {
	return uint64(time.Now().UnixNano())
}
