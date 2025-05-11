package ratelimiter

import (
	"sync"
	"testing"
	"time"
)

func TestRateLimiter_BasicLimit(t *testing.T) {
	t.Parallel()
	n := 3
	limiter := NewRateLimiter(n)
	clientName := "testClient1"

	for i := 0; i < n; i++ {
		if !limiter.RateLimit(clientName) {
			t.Errorf(
				"Request %d for %s should have been allowed, but was denied",
				i+1,
				clientName,
			)
		}
	}

	if limiter.RateLimit(clientName) {
		t.Errorf(
			"Request %d for %s should have been denied, but was allowed",
			n+1,
			clientName,
		)
	}
}

func TestRateLimiter_TokenRefill(t *testing.T) {
	t.Parallel()
	n := 1
	limiter := NewRateLimiter(n)
	clientName := "testClient2"

	// Use up the token
	if !limiter.RateLimit(clientName) {
		t.Fatalf("%s: First request should be allowed", clientName)
	}

	// This one should be denied
	if limiter.RateLimit(clientName) {
		t.Fatalf(
			"%s: Second request immediately after should be denied",
			clientName,
		)
	}

	// Wait for token to refill (rate is 1 per second)
	time.Sleep(
		1050 * time.Millisecond,
	) // Sleep a bit more than 1s to ensure refill

	if !limiter.RateLimit(clientName) {
		t.Fatalf(
			"%s: Request after 1s wait should be allowed due to token refill",
			clientName,
		)
	}
}

func TestRateLimiter_BurstAllowance(t *testing.T) {
	t.Parallel()
	n := 5 // Bucket size and rate are 5
	limiter := NewRateLimiter(n)
	clientName := "testClient3"

	// Client is new, should have full bucket
	for i := 0; i < n; i++ {
		if !limiter.RateLimit(clientName) {
			t.Errorf(
				"Burst request %d for %s should have been allowed, but was denied",
				i+1,
				clientName,
			)
		}
	}
	if limiter.RateLimit(clientName) {
		t.Errorf(
			"Burst request %d for %s (exceeding bucket) should have been denied, but was allowed",
			n+1,
			clientName,
		)
	}

	// Wait for some tokens to refill, but not all
	time.Sleep(
		time.Duration(float64(time.Second) * 2.0 / float64(n)),
	) // Wait for 2 tokens to refill (2/5th of a second)

	// Should be able to make 2 more requests
	for i := 0; i < 2; i++ {
		if !limiter.RateLimit(clientName) {
			t.Errorf(
				"Request %d post-partial-refill for %s should have been allowed",
				i+1,
				clientName,
			)
			return
		}
	}
	if limiter.RateLimit(clientName) {
		t.Errorf(
			"Request %d post-partial-refill for %s (exceeding refilled tokens) should have been denied",
			3,
			clientName,
		)
	}
}

func TestRateLimiter_MultipleClients(t *testing.T) {
	t.Parallel()
	n := 2
	limiter := NewRateLimiter(n)
	clientA := "clientA"
	clientB := "clientB"

	// Client A uses up its quota
	for i := 0; i < n; i++ {
		if !limiter.RateLimit(clientA) {
			t.Errorf("Request %d for %s should have been allowed", i+1, clientA)
		}
	}
	if limiter.RateLimit(clientA) {
		t.Errorf("Request %d for %s should have been denied", n+1, clientA)
	}

	// Client B should still have its full quota
	for i := 0; i < n; i++ {
		if !limiter.RateLimit(clientB) {
			t.Errorf(
				"Request %d for %s should have been allowed, but was denied (independent of %s)",
				i+1,
				clientB,
				clientA,
			)
		}
	}
	if limiter.RateLimit(clientB) {
		t.Errorf(
			"Request %d for %s should have been denied (independent of %s)",
			n+1,
			clientB,
			clientA,
		)
	}
}

func TestRateLimiter_ConcurrentAccess(t *testing.T) {
	t.Parallel()
	n := 100 // Allow a higher number for concurrency test
	limiter := NewRateLimiter(n)
	clientName := "concurrentClient"
	numGoroutines := 200 // More goroutines than allowed requests
	allowedCount := 0
	var mu sync.Mutex // Mutex for shared allowedCount
	var wg sync.WaitGroup

	// All goroutines will try to make a request at roughly the same time
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if limiter.RateLimit(clientName) {
				mu.Lock()
				allowedCount++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()

	if allowedCount > n {
		t.Errorf(
			"Allowed %d requests concurrently, but bucket size was %d. Expected at most %d.",
			allowedCount,
			n,
			n,
		)
	}
}

func TestRateLimiter_ZeroOrNegativeN(t *testing.T) {
	t.Parallel()
	limiter := NewRateLimiter(0) // Should default to 1
	if !limiter.RateLimit("clientX") {
		t.Error("RateLimiter(0): First request should be allowed (default N=1)")
	}
	if limiter.RateLimit("clientX") {
		t.Error("RateLimiter(0): Second request should be denied (default N=1)")
	}

	limiter = NewRateLimiter(-5) // Should default to 1
	if !limiter.RateLimit("clientY") {
		t.Error(
			"RateLimiter(-5): First request should be allowed (default N=1)",
		)
	}
	if limiter.RateLimit("clientY") {
		t.Error(
			"RateLimiter(-5): Second request should be denied (default N=1)",
		)
	}
}
