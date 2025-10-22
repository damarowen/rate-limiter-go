package rateLimiter

type Strategy interface {
	Allow(key string) bool
	Reset(key string)
}

type RateLimiter struct {
	defaultStrategy Strategy
	premiumClients  map[string]Strategy
	requests        chan request
	config          chan configRequest
	resets          chan string // Add this
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
		config:          make(chan configRequest),
		resets:          make(chan string),
	}
	go rl.process()
	return rl
}

// thread-safe , Multiple goroutines can safely send
func (rl *RateLimiter) SetPremiumClient(key string, strategy Strategy) {
	rl.config <- configRequest{
		key:      key,
		strategy: strategy,
	}
}

func (rl *RateLimiter) Allow(key string) bool {
	response := make(chan bool)
	//Sends requests from Allow() to process()
	rl.requests <- request{
		key:      key,
		response: response,
	}
	return <-response // Blocks until response arr
}

func (rl *RateLimiter) Reset(key string) {
	rl.resets <- key // Send to process goroutine
}

// listening requests and config changes
func (rl *RateLimiter) process() {
	for {
		select {
		// Only ONE goroutine receives and processes
		//Handles dynamic configuration of premium clients
		//allows adding premium users with custom rate limits while the server is running.
		case cfg := <-rl.config:
			// Handle premium client configuration

			rl.premiumClients[cfg.key] = cfg.strategy // ← Write

		case req := <-rl.requests:
			// Handle rate limit check
			if strategy, isPremium := rl.premiumClients[req.key]; isPremium {
				req.response <- strategy.Allow(req.key) // ← Read
			} else {
				//Sends results from process() back to Allow()
				req.response <- rl.defaultStrategy.Allow(req.key) // ← Read
			}
		case key := <-rl.resets:
			if strategy, isPremium := rl.premiumClients[key]; isPremium {
				strategy.Reset(key)
			} else {
				rl.defaultStrategy.Reset(key)
			}
		}
	}
}
