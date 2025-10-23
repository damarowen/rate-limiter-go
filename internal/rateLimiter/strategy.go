package rateLimiter

import (
	"fmt"
	"time"
)

// ClientBucket tracks requests for a single client
type ClientBucket struct {
	count       int
	windowStart time.Time
}

// FixedWindowStrategy implements fixed window rate limiting
type FixedWindowStrategy struct {
	limit   int
	window  time.Duration
	buckets map[string]*ClientBucket
	ops     chan func()
}

// NewFixedWindowStrategy creates a new fixed window strategy
func NewFixedWindowStrategy(limit int, window time.Duration) *FixedWindowStrategy {
	fws := &FixedWindowStrategy{
		limit:   limit,                          // Max requests (e.g., 10)
		window:  window,                         // Time window (e.g., 1 minute)
		buckets: make(map[string]*ClientBucket), // Empty map for tracking clients, key string value *ClientBucket
		ops:     make(chan func(), 100),         // Buffered channel with 100 capacity, if  reached > 100 it will block
	}

	go fws.run()
	go fws.cleanup()

	return fws
}

// run processes all operations sequentially
func (fws *FixedWindowStrategy) run() {
	for op := range fws.ops {
		op()
	}
}

// Allow checks if a request is allowed
func (fws *FixedWindowStrategy) Allow(key string) bool {
	result := make(chan bool, 1)

	fws.ops <- func() {
		// Add this to see all buckets
		fmt.Println("=== Current Buckets ===")
		for k, b := range fws.buckets {
			fmt.Printf("Key: %s | Count: %d | Window: %s\n",
				k, b.count, b.windowStart.Format("15:04:05"))
		}
		fmt.Println("=======================")

		now := time.Now()
		bucket, exists := fws.buckets[key]

		if !exists {
			//initiate jika tidak ada
			fws.buckets[key] = &ClientBucket{
				count:       1,
				windowStart: now,
			}
			result <- true
			return
		}

		//Jika klien ada, periksa apakah jendela waktunya sudah lewat
		if now.Sub(bucket.windowStart) >= fws.window {
			//reset
			bucket.count = 1
			bucket.windowStart = now
			result <- true
			return
		}

		//Jika jendela belum kadaluarsa:
		if bucket.count < fws.limit {
			bucket.count++
			result <- true
			return
		}

		result <- false
	}

	return <-result
}

// Reset resets the bucket for a specific key
func (fws *FixedWindowStrategy) Reset(key string) {
	fws.ops <- func() {
		delete(fws.buckets, key)
	}
}

// cleanup removes expired buckets periodically
func (fws *FixedWindowStrategy) cleanup() {
	ticker := time.NewTicker(fws.window)
	defer ticker.Stop()

	for range ticker.C {
		fws.ops <- func() {
			now := time.Now()
			for key, bucket := range fws.buckets {
				//Deletes buckets inactive for 2 × window (e.g., 2 minutes)
				if now.Sub(bucket.windowStart) >= fws.window*2 {
					delete(fws.buckets, key)
				}
			}
		}
	}
}
