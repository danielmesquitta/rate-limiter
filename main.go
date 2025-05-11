package main

import (
	"fmt"
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

func main() {
	// Example Usage:
	// N = 3 requests per second per client.
	limiter := NewRateLimiter(3)

	// Simulate requests from "clientA"
	fmt.Println("Simulating requests for clientA (N=3):")
	for i := 0; i < 5; i++ {
		if limiter.RateLimit("clientA") {
			fmt.Printf("Request %d for clientA: Allowed\n", i+1)
		} else {
			fmt.Printf("Request %d for clientA: Denied\n", i+1)
		}
		time.Sleep(100 * time.Millisecond) // Small delay between quick requests
	}

	fmt.Println("\nWaiting for tokens to refill for clientA...")
	time.Sleep(1 * time.Second) // Wait for 1 second

	if limiter.RateLimit("clientA") {
		fmt.Println("Request after 1s wait for clientA: Allowed")
	} else {
		fmt.Println("Request after 1s wait for clientA: Denied")
	}

	// Simulate requests from "clientB"
	fmt.Println("\nSimulating requests for clientB (N=3):")
	for i := 0; i < 4; i++ {
		if limiter.RateLimit("clientB") {
			fmt.Printf("Request %d for clientB: Allowed\n", i+1)
		} else {
			fmt.Printf("Request %d for clientB: Denied\n", i+1)
		}
	}
}
