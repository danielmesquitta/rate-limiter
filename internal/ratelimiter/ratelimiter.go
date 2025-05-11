package ratelimiter

import (
	"sync"
	"time"
)

// clientData holds the state of a single client's token bucket.
type clientData struct {
	tokens         float64    // Current number of tokens available. Using float for precise refill.
	lastRefillTime time.Time  // Timestamp of the last token refill.
	mu             sync.Mutex // Mutex to protect this client's data during concurrent access.
}

// RateLimiter manages rate limiting for multiple clients using the Token Bucket algorithm.
type RateLimiter struct {
	clients    map[string]*clientData // Map to store each client's bucket data.
	rate       float64                // Tokens added per second (N).
	bucketSize float64                // Maximum tokens in a bucket (N).
	mu         sync.Mutex             // Mutex to protect the 'clients' map.
}

// NewRateLimiter creates a new RateLimiter.
// n: the number of requests allowed per second per client.
func NewRateLimiter(n int) *RateLimiter {
	if n <= 0 {
		n = 1 // Default to at least 1 request per second if n is invalid
	}
	return &RateLimiter{
		clients:    make(map[string]*clientData),
		rate:       float64(n),
		bucketSize: float64(n),
	}
}

// RateLimit checks if a request from a given client should be allowed or denied.
// It returns true if the request is allowed, false otherwise.
func (rl *RateLimiter) RateLimit(clientName string) bool {
	rl.mu.Lock()
	client, exists := rl.clients[clientName]
	if !exists {
		client = &clientData{
			tokens:         rl.bucketSize, // New client starts with a full bucket.
			lastRefillTime: time.Now(),
		}
		rl.clients[clientName] = client
	}
	rl.mu.Unlock()

	client.mu.Lock()
	defer client.mu.Unlock()

	// Refill tokens based on elapsed time.
	now := time.Now()
	elapsed := now.Sub(client.lastRefillTime).Seconds()
	tokensToAdd := elapsed * rl.rate
	client.tokens += tokensToAdd
	client.lastRefillTime = now

	// Cap tokens at bucketSize.
	if client.tokens > rl.bucketSize {
		client.tokens = rl.bucketSize
	}

	// Check if enough tokens are available.
	if client.tokens >= 1 {
		client.tokens--
		return true // Allowed
	}

	return false // Denied
}
