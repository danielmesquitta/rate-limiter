# Rate Limiter

## Overview

This document explains the implementation of a rate limiter using the Token Bucket algorithm, along with its trade-offs and design decisions.

## Algorithm: Token Bucket

### Concept Overview

| Concept     | What it means                                  |
| ----------- | ---------------------------------------------- |
| Bucket size | N tokens (same as allowed requests per second) |
| Refill rate | N tokens added every second                    |
| Request     | Consumes 1 token. If no tokens → deny          |

### Flow

1. First request → bucket starts full
2. On each call:
   - Calculate new tokens: `tokensToAdd = elapsedTime × rate` (cap at bucket size)
   - If tokens ≥ 1 → allow & subtract 1 token; else deny

## Why we chose Token Bucket

- **Handles bursts** – users can spend saved tokens quickly (up to bucket size)
- **Smooth average** – long‑term rate never exceeds N/s
- **Lightweight** – store only three values per client (tokens, lastRefill, rate)
- **Proven** – used by AWS, Stripe, etc.

## Other options (and their deal‑breakers)

### 1. Fixed Window Counter

**Pros:**

- Simplest to implement
- Low memory usage (just one counter per client)

**Cons:**

- Allows double the intended rate at window boundaries
  - Example: With a limit of 100 requests per second
    - At 0.999 seconds: Can make 100 requests
    - At 1.001 seconds: Can make another 100 requests
    - Result: 200 requests in just 0.002 seconds
- Can overwhelm the service due to this burst behavior

### 2. Sliding Window Log

**Pros:**

- Very accurate rate limiting
  - Example: With a limit of 100 requests per minute
    - Keeps track of exact timestamps of all requests
    - Can precisely enforce the limit over any 60-second window
    - No burst issues at window boundaries
- Strict rate limit enforcement
  - Perfect for applications requiring precise control
  - No edge cases or timing exploits possible

**Cons:**

- Memory-intensive
  - Stores timestamp for every request
  - Example: At 100 requests/minute, stores 100 timestamps per client
  - Memory usage grows linearly with request volume

#### 3. Sliding Window Counter

**Pros:**

- Good compromise between Fixed Window and Sliding Window Log
  - Example: With a limit of 100 requests per minute
    - Divides time into smaller sub-windows (e.g., 10-second intervals)
    - Each sub-window tracks its own count
    - Weighted average of current and previous windows
- Memory efficient compared to Sliding Window Log

#### 4. Token Bucket (Chosen Implementation)

**Pros:**

- Balanced performance and resource usage
  - Example: With a limit of 100 requests per minute
    - Only stores 3 values per client:
      1. Current token count
      2. Last refill timestamp
      3. Token refill rate
    - No need to store request history
- Good traffic smoothing
  - Tokens refill at constant rate
  - Example: 100 tokens/minute = 1.67 tokens/second
    - Client can use tokens as they arrive
    - Or save up for bursts (up to bucket size)
