# Rate Limiter

A simple rate limiter implementation in Go without external dependencies.

## Features

### 🎯 Core Rate Limiting
- ✅ In-memory storage (Go maps) - no database required
- ✅ Thread-safe with channel-based synchronization
- ✅ Concurrent request handling without race conditions
- ✅ Zero external dependencies - pure Go standard library

### 📊 Multiple Rate Limiting Strategies
- ✅ **Fixed Window Counter** - Simple, memory efficient
- ✅ **Sliding Window Log** - Accurate, smooth limiting
- ✅ **Token Bucket** - Handles traffic bursts gracefully

### 🔑 Flexible Client Identification
- ✅ IP-based rate limiting
- ✅ API key-based rate limiting
- ✅ Custom header extraction
- ✅ Per-client custom limits

### 💎 Premium Tier Support
- ✅ Configurable limits per API key
- ✅ Premium users: 100 req/min
- ✅ Basic users: 10 req/min (default)
- ✅ Easy tier configuration

### 🔌 Easy Integration
- ✅ HTTP middleware for `net/http`
- ✅ Plug-and-play with existing servers
- ✅ Clean API with minimal code changes

### ✅ Production-Ready
- ✅ Comprehensive unit tests
- ✅ Race condition testing
- ✅ Clean architecture (Strategy + Middleware patterns)
- ✅ Lightweight and fast
- ✅ Docker


## Prerequisites

- Go 1.16 or higher

## Installation

```bash
git clone <repository-url>
cd rate-limiter
go mod download
go build -o rate-limiter
./rate-limiter
```

# Project Structure
````
rate-limiter/
├── cmd/
│   └── server/
│       └── main.go              # Application entry point
├── internal/
│   ├── middleware/
│   │   └── middleware.go        # HTTP middleware for rate limiting
│   └── rateLimiter/
│       ├── rateLimiter.go       # Core rate limiter implementation
│       ├── strategy.go          # Rate limiting strategies
│       └── rateLimiter_test.go  # Unit tests
├── go.mod                       # Go module file
└── README.md                    # Project documentation
````

# Testing
1. go test ./internal/rateLimiter/...
2. go test -v -cover ./internal/rateLimiter/...

# Architecture
- Strategy Pattern: Implemented to allow pluggable rate limiting algorithms. This makes it easy to switch between different strategies or add new ones without modifying existing code.
- Middleware Approach: Rate limiting is implemented as HTTP middleware, allowing easy integration into any HTTP server with minimal code changes.
- In-Memory Storage: Uses Go maps for simplicity and speed. Suitable for single-instance deployments. For distributed systems, this could be extended to use Redis or similar.
- Clean Architecture: Separation of concerns with:
cmd/ for application entry points
internal/ for implementation details (not importable by external packages)
middleware/ for HTTP integration layer
rateLimiter/ for core business logic

# Configuration
````
// Example configuration
limiter := rateLimiter.NewRateLimiter(
    100,              // max requests
    time.Minute,      // time window
    yourStrategy,     // rate limiting strategy
)
````

# API Usage Example
```
Test Basic Tier:
GET http://localhost:8080/api/hello
Header: X-API-Key: basic-user-123

Response (1st-10th request):
{
  "message": "Hello! You are within rate limits.",
  "time": "2025-01-15T10:30:00Z",
  "user_tier": "basic",
  "rate_limit": "10 requests/minute"
}

Response (11th request - Rate Limited):
{
  "error": "rate limit exceeded",
  "tier": "basic",
  "limit": "10 requests/minute"
}

Test Premium Tier:
GET http://localhost:8080/api/hello
Header: X-API-Key: premium-api-key

Response (1st-100th request):
{
  "message": "Hello! You are within rate limits.",
  "time": "2025-01-15T10:30:00Z",
  "user_tier": "premium",
  "rate_limit": "100 requests/minute"
}

Response (101st request - Rate Limited):
{
  "error": "rate limit exceeded",
  "tier": "premium",
  "limit": "100 requests/minute"
}
```

## Per-Client Configuration ✨

### Premium Tier Example

```
rl := rateLimiter.NewRateLimiter(defaultStrategy)

// Set premium tier: 100 requests/minute
rl.SetPremiumClient("premium-api-key", 100, time.Minute)

// Basic users get default limit (10 req/min)

# Premium user (100 req/min)
curl -H "X-API-Key: premium-api-key" http://localhost:8080/api

# Basic user (10 req/min default)
curl -H "X-API-Key: any-other-key" http://localhost:8080/api
````

## Key Benefits of This Approach

✅ **No mutex needed** - uses existing channel pattern  
✅ **Thread-safe** - all operations serialized through channel  
✅ **Simple** - just adds `premiumClients` map lookup  
✅ **Consistent** - follows your existing architecture  
✅ **Meets Requirement #3** - premium clients get 100 req/min

## How It Works

```
// Request flow:
// 1. Request comes in with API key
// 2. Allow() called with API key as key
// 3. Channel operation checks if key is in premiumClients map
// 4. If premium → use premium strategy (100 req/min)
// 5. If not premium → use default strategy (10 req/min)
````

# Rate Limiting Algorithms
1. Fixed Window Counter
   Description: Counts requests in fixed time windows
   How it works: Resets the counter at the start of each time window
   Pros: Simple, memory efficient
   Cons: Can allow bursts at window boundaries
2. Sliding Window Log
   Description: Maintains a log of request timestamps
   How it works: Counts requests within a sliding time window
   Pros: Accurate, smooth rate limiting
   Cons: Higher memory usage for storing timestamps
3. Token Bucket
   Description: Tokens are added at a fixed rate, requests consume tokens
   How it works: Allows bursts up to bucket capacity
   Pros: Handles bursts gracefully
   Cons: More complex implementation

### Burst Traffic Handling Example

#### Token Bucket Strategy (Recommended for Bursts)
```
// Allow burst of 100 requests, refill at 50/second
strategy := NewTokenBucketStrategy(100, 50)

// Client can immediately send 100 requests (burst)
for i := 0; i < 100; i++ {
    strategy.Allow("client1") // ✅ All allowed
}

// Further requests limited to 50/second refill rate
strategy.Allow("client1") // ❌ Blocked until tokens refill

// After 2 seconds: 100 tokens refilled
time.Sleep(2 * time.Second)
// Can burst again

Benefits:
Handles legitimate traffic spikes
Prevents sustained abuse
Smooth token replenishment
Configurable burst capac
```


# Assumptions
1. Single Instance: Current implementation assumes a single server instance. For distributed systems, you'll need a shared cache (e.g., Redis).
2. Client Identification: Clients are identified by IP address or custom headers. Modify the middleware to change identification logic.
3. Memory Limits: In-memory storage is suitable for moderate traffic. High-traffic scenarios should use persistent storage.
4. Time Synchronization: Assumes server time is accurate. Clock skew can affect rate limiting accuracy.

# Limitations
1. No Persistence: Rate limit data is lost on server restart
2. Single Instance Only: Not designed for distributed deployments without modification
3. Memory Growth: Long-running instances may need periodic cleanup of old entries
4. No Built-in Cleanup: Old client records are not automatically removed (can be added)
5. Fixed Time Windows: Some strategies use fixed time windows which can have edge cases