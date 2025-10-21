package rateLimiter

// Strategy defines the rate limiting algorithm interface
type Strategy interface {
	Allow(key string) bool
	Reset(key string)
}

// RateLimiter manages rate limiting for multiple clients
type RateLimiter struct {
	defaultStrategy Strategy
	premiumClients  map[string]Strategy
	requests        chan request
}

type request struct {
	key      string
	response chan bool
}

// NewRateLimiter creates a new rate limiter with the given default strategy
func NewRateLimiter(defaultStrategy Strategy) *RateLimiter {
	rl := &RateLimiter{
		defaultStrategy: defaultStrategy,
		premiumClients:  make(map[string]Strategy),
		requests:        make(chan request),
	}
	go rl.process()
	return rl
}

// SetPremiumClient configures a premium client with a custom strategy
func (rl *RateLimiter) SetPremiumClient(key string, strategy Strategy) {
	// Send configuration through channel to maintain thread-safety
	go func() {
		req := request{
			key:      "CONFIG:" + key, // Special prefix for config operations
			response: make(chan bool),
		}
		rl.requests <- req
		<-req.response

		// Store in map (safe because process() goroutine handles it)
		rl.premiumClients[key] = strategy
	}()
}

// Allow checks if a request from the given key is allowed
func (rl *RateLimiter) Allow(key string) bool {
	response := make(chan bool)
	rl.requests <- request{
		key:      key,
		response: response,
	}
	return <-response
}

// process handles all rate limiting logic in a single goroutine
func (rl *RateLimiter) process() {
	for req := range rl.requests {
		// Check if this is a premium client
		if strategy, isPremium := rl.premiumClients[req.key]; isPremium {
			req.response <- strategy.Allow(req.key)
		} else {
			req.response <- rl.defaultStrategy.Allow(req.key)
		}
	}
}

// Reset resets the rate limit for a specific key
func (rl *RateLimiter) Reset(key string) {
	if strategy, isPremium := rl.premiumClients[key]; isPremium {
		strategy.Reset(key)
	} else {
		rl.defaultStrategy.Reset(key)
	}
}
