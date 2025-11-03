package rateLimiter

import (
	"fmt"
	"log"
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
	go fws.cleanup() //to prevent memory leak

	return fws
}

// run processes all operations sequentially
func (fws *FixedWindowStrategy) run() {
	//Setiap kali Allow() atau Reset() dipanggil, mereka tidak langsung mengubah map.
	//Mereka "mengirimkan pekerjaan" (dalam bentuk anonymous function) ke channel ops.
	// listening
	for op := range fws.ops {
		op()
	}
}

// Allow checks if a request is allowed
// use Fixed Window Counter Algorithm
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
		//Waktu Sekarang - Waktu Mulai Jendela
		//eg (10:30:01 - 10:00:00) >= 30 minutes
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

		log.Printf("Request for key %s exceeded rate limit", key)
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
	//Membuat timer yang akan "berdetak" setiap fws.window (misal: setiap 1 menit).
	ticker := time.NewTicker(fws.window)
	defer ticker.Stop()

	for range ticker.C {
		fws.ops <- func() {
			now := time.Now()
			for key, bucket := range fws.buckets {
				//Deletes buckets inactive for 2 × window (e.g., 2 minutes)
				// di kali 2 supaya tidak agressive dan lebih aman
				if now.Sub(bucket.windowStart) >= fws.window*2 {
					delete(fws.buckets, key)
				}
			}
		}
	}
}
