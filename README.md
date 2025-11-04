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
go mod tidy
go build -o rate-limiter ./cmd/server
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

### Fixed Window Counter

- **Description:** Counts requests in fixed time windows
- **How it works:** Resets the counter at the start of each time window
- **Pros:**
    - Simple implementation
    - Low memory usage (only stores count + timestamp)
    - Fast O(1) lookup
- **Cons:**
    - Can allow bursts at window boundaries
    - Example: If window = 1 minute, user can make 10 requests at 00:59 and another 10 at 01:01 (20 requests in 2 seconds)



# Request Processing Flow

```
Client sends HTTP request
   ↓
Server receives request on :8080
   ↓
RateLimitMiddleware intercepts
   ↓
┌─────────────────────────────────────────┐
│ 1. Extract API Key from X-API-Key header│ ← Layer 1: HTTP Gateway
└─────────────────────────────────────────┘
   ↓
┌─────────────────────────────────────────┐
│ 2. Call limiter.Allow(apiKey)           │
└─────────────────────────────────────────┘
   ↓
┌─────────────────────────────────────────┐
│ 3. RateLimiter.Allow() sends request    │
│    to process() goroutine via channel   │
└─────────────────────────────────────────┘
   ↓
┌─────────────────────────────────────────┐
│ 4. process() checks premiumClients map  │
│    • If key = "premium-api-key"         │ ← Layer 2: Router/Orchestrator
│      → Use premium strategy (100/min)   │
│    • Otherwise                          │
│      → Use default strategy (10/min)    │
└─────────────────────────────────────────┘
   ↓
┌─────────────────────────────────────────┐
│ 5. Strategy.Allow() checks buckets map  │
│    • Get or create bucket for this key  │  ← Layer 3: Business Logic
│    • Check if window expired            │
│    • Increment count                    │
│    • Return true/false                  │
└─────────────────────────────────────────┘
   ↓
┌─────────────────────────────────────────┐
│ 6. Send result back via response channel│
└─────────────────────────────────────────┘
   ↓
┌─────────────────────────────────────────┐
│ 7. Middleware receives result           │
└─────────────────────────────────────────┘
   ↓
   ├─ If ALLOWED (true)
   │    ↓
   │  Continue to handler (/api/hello or /api/status)
   │    ↓
   │  Handler executes
   │    ↓
   │  Response sent with 200 OK
   │
   └─ If DENIED (false)
        ↓
      Return 429 Too Many Requests
        ↓
      Send JSON error response
      
      
Middleware → HTTP concerns
RateLimiter → Routing concerns  
Strategy → Algorithm concerns

HTTP Request → Middleware → RateLimiter → Strategy
                    ↓           ↓                    ↓
                Extract    Route to strategy    Implement rate limiting algorithm
                identifier
                    ↓           ↓                                        ↓
                 Panggil      Manage premium vs default clients     Track request counts per client
                 Allow(),
                 send http
                 
      
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


## Future Improvements

With more time, I would implement:

1. **Sliding Window Log Algorithm**
    - More accurate rate limiting
    - No burst at window boundaries
    - Trade-off: Higher memory usage

2. **Token Bucket Algorithm**
    - Allow controlled bursts
    - Better for bursty traffic patterns


# Test Thread Safe

````
go test -race -v ./internal/rateLimiter/...

Codebase ini thread-safe karena:
Shared state (maps) hanya diakses 1 goroutine (process())
Communication: All via channels (requests, config, resets)
Sequential processing (requests diproses 1-by-1 di queue
````

# Docker Run

````
docker build -t rate-limiter -f .dockerfile .
docker run -p 8080:8080 rate-limiter

docker stop my-rate-limiter
docker rm my-rate-limiter
````
