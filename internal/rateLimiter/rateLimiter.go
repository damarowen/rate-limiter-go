package rateLimiter

type Strategy interface {
	Allow(key string) bool
	Reset(key string)
}

type RateLimiter struct {
	defaultStrategy Strategy
	premiumClients  map[string]Strategy
	requests        chan request
	config          chan configRequest // Add this
}

type request struct {
	key      string
	response chan bool
}

type configRequest struct { // Add this
	key      string
	strategy Strategy
}

func NewRateLimiter(defaultStrategy Strategy) *RateLimiter {
	rl := &RateLimiter{
		defaultStrategy: defaultStrategy,
		premiumClients:  make(map[string]Strategy),
		requests:        make(chan request),
		config:          make(chan configRequest), // Add this
	}
	go rl.process()
	return rl
}

// FIXED: Now thread-safe
func (rl *RateLimiter) SetPremiumClient(key string, strategy Strategy) {
	rl.config <- configRequest{
		key:      key,
		strategy: strategy,
	}
}

func (rl *RateLimiter) Allow(key string) bool {
	response := make(chan bool)
	rl.requests <- request{
		key:      key,
		response: response,
	}
	return <-response
}

// FIXED: All map access in single goroutine
func (rl *RateLimiter) process() {
	for {
		select {
		case cfg := <-rl.config:
			// Handle premium client configuration
			rl.premiumClients[cfg.key] = cfg.strategy

		case req := <-rl.requests:
			// Handle rate limit check
			if strategy, isPremium := rl.premiumClients[req.key]; isPremium {
				req.response <- strategy.Allow(req.key)
			} else {
				req.response <- rl.defaultStrategy.Allow(req.key)
			}
		}
	}
}

func (rl *RateLimiter) Reset(key string) {
	response := make(chan bool)
	rl.requests <- request{
		key:      "RESET:" + key,
		response: response,
	}
	<-response

	// Delegate to strategy
	if strategy, isPremium := rl.premiumClients[key]; isPremium {
		strategy.Reset(key)
	} else {
		rl.defaultStrategy.Reset(key)
	}
}
