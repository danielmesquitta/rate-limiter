package main

import (
	"fmt"
	"time"

	ratelimiter "github.com/danielmesquitta/rate-limiter/internal/ratelimiter"
)

func main() {
	// Example Usage:
	// N = 3 requests per second per client.
	limiter := ratelimiter.NewRateLimiter(3)

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
