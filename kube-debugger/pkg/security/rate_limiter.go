package security

import (
	"fmt"
	"sync"
	"time"
)

// RateLimiter implements token bucket rate limiting
type RateLimiter struct {
	mu              sync.Mutex
	requestsPerSec  float64
	tokensAvailable float64
	maxTokens       float64
	lastRefillTime  time.Time
	allowBurst      bool
}

// NewRateLimiter creates a new rate limiter with specified requests per second
func NewRateLimiter(requestsPerSec float64, allowBurst bool) *RateLimiter {
	return &RateLimiter{
		requestsPerSec:  requestsPerSec,
		tokensAvailable: requestsPerSec,
		maxTokens:       requestsPerSec,
		lastRefillTime:  time.Now(),
		allowBurst:      allowBurst,
	}
}

// Allow checks if the next request should be allowed
func (rl *RateLimiter) Allow() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.refillTokens()

	if rl.tokensAvailable >= 1.0 {
		rl.tokensAvailable--
		return true
	}

	return false
}

// AllowN checks if N requests should be allowed
func (rl *RateLimiter) AllowN(n int) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.refillTokens()

	if rl.tokensAvailable >= float64(n) {
		rl.tokensAvailable -= float64(n)
		return true
	}

	return false
}

// Wait blocks until a request is allowed
func (rl *RateLimiter) Wait() {
	for !rl.Allow() {
		time.Sleep(10 * time.Millisecond)
	}
}

// WaitN blocks until N requests are allowed
func (rl *RateLimiter) WaitN(n int) {
	for !rl.AllowN(n) {
		time.Sleep(10 * time.Millisecond)
	}
}

// refillTokens adds tokens based on elapsed time
func (rl *RateLimiter) refillTokens() {
	now := time.Now()
	elapsed := now.Sub(rl.lastRefillTime).Seconds()
	tokensToAdd := elapsed * rl.requestsPerSec

	rl.tokensAvailable += tokensToAdd
	if rl.tokensAvailable > rl.maxTokens {
		rl.tokensAvailable = rl.maxTokens
	}

	rl.lastRefillTime = now
}

// OperationLimiter limits operations with per-resource tracking
type OperationLimiter struct {
	mu       sync.Mutex
	limiters map[string]*RateLimiter
	opsPerSec float64
}

// NewOperationLimiter creates a new operation limiter
func NewOperationLimiter(opsPerSec float64) *OperationLimiter {
	return &OperationLimiter{
		limiters:  make(map[string]*RateLimiter),
		opsPerSec: opsPerSec,
	}
}

// Allow checks if an operation on a resource is allowed
func (ol *OperationLimiter) Allow(resource string) bool {
	ol.mu.Lock()
	limiter, exists := ol.limiters[resource]
	if !exists {
		limiter = NewRateLimiter(ol.opsPerSec, true)
		ol.limiters[resource] = limiter
	}
	ol.mu.Unlock()

	return limiter.Allow()
}

// Wait blocks until an operation is allowed
func (ol *OperationLimiter) Wait(resource string) {
	ol.mu.Lock()
	limiter, exists := ol.limiters[resource]
	if !exists {
		limiter = NewRateLimiter(ol.opsPerSec, true)
		ol.limiters[resource] = limiter
	}
	ol.mu.Unlock()

	limiter.Wait()
}

// ResetResource resets the limiter for a specific resource
func (ol *OperationLimiter) ResetResource(resource string) {
	ol.mu.Lock()
	delete(ol.limiters, resource)
	ol.mu.Unlock()
}

// RequestCounter tracks request counts for rate limit enforcement
type RequestCounter struct {
	mu            sync.Mutex
	counts        map[string]int
	limits        map[string]int
	resetTime     time.Time
	resetInterval time.Duration
}

// NewRequestCounter creates a new request counter
func NewRequestCounter(resetInterval time.Duration) *RequestCounter {
	return &RequestCounter{
		counts:        make(map[string]int),
		limits:        make(map[string]int),
		resetTime:     time.Now(),
		resetInterval: resetInterval,
	}
}

// SetLimit sets the request limit for an identifier
func (rc *RequestCounter) SetLimit(identifier string, limit int) {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	rc.limits[identifier] = limit
}

// Increment increments the counter and returns error if limit exceeded
func (rc *RequestCounter) Increment(identifier string) error {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	// Check if reset interval has passed
	if time.Since(rc.resetTime) > rc.resetInterval {
		rc.counts = make(map[string]int)
		rc.resetTime = time.Now()
	}

	limit, exists := rc.limits[identifier]
	if !exists {
		limit = 100 // Default limit
	}

	rc.counts[identifier]++

	if rc.counts[identifier] > limit {
		return fmt.Errorf("rate limit exceeded for %s (limit: %d)", identifier, limit)
	}

	return nil
}

// GetCount returns current count for identifier
func (rc *RequestCounter) GetCount(identifier string) int {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	return rc.counts[identifier]
}

// Reset resets all counters
func (rc *RequestCounter) Reset() {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	rc.counts = make(map[string]int)
	rc.resetTime = time.Now()
}

// Global rate limiters
var (
	globalKubeAPILimiter   *RateLimiter
	globalOperationLimiter *OperationLimiter
	apiLimiterOnce         sync.Once
	opLimiterOnce          sync.Once
)

// GetGlobalRateLimiter returns the global Kubernetes API rate limiter
func GetGlobalRateLimiter() *RateLimiter {
	apiLimiterOnce.Do(func() {
		// 100 requests per second by default
		globalKubeAPILimiter = NewRateLimiter(100, true)
	})
	return globalKubeAPILimiter
}

// GetGlobalOperationLimiter returns the global operation limiter
func GetGlobalOperationLimiter() *OperationLimiter {
	opLimiterOnce.Do(func() {
		// 50 operations per second by default
		globalOperationLimiter = NewOperationLimiter(50)
	})
	return globalOperationLimiter
}
