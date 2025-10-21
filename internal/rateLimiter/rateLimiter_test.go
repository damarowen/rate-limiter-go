package rateLimiter

import (
	"sync"
	"testing"
	"time"
)

func TestFixedWindowStrategy_Allow(t *testing.T) {
	strategy := NewFixedWindowStrategy(5, time.Second)

	// Test within limit
	for i := 0; i < 5; i++ {
		if !strategy.Allow("client1") {
			t.Errorf("Request %d should be allowed", i+1)
		}
	}

	// Test exceeding limit
	if strategy.Allow("client1") {
		t.Error("Request 6 should be blocked")
	}
}

func TestFixedWindowStrategy_WindowReset(t *testing.T) {
	strategy := NewFixedWindowStrategy(2, 100*time.Millisecond)

	strategy.Allow("client1")
	strategy.Allow("client1")

	if strategy.Allow("client1") {
		t.Error("Should be rate limited")
	}

	time.Sleep(150 * time.Millisecond)

	if !strategy.Allow("client1") {
		t.Error("Should be allowed after window reset")
	}
}

func TestFixedWindowStrategy_MultipleClients(t *testing.T) {
	strategy := NewFixedWindowStrategy(3, time.Second)

	for i := 0; i < 3; i++ {
		strategy.Allow("client1")
		strategy.Allow("client2")
	}

	if strategy.Allow("client1") {
		t.Error("Client1 should be rate limited")
	}

	if strategy.Allow("client2") {
		t.Error("Client2 should be rate limited")
	}
}

func TestRateLimiter_Concurrency(t *testing.T) {
	strategy := NewFixedWindowStrategy(100, time.Second)
	rl := NewRateLimiter(strategy)

	var wg sync.WaitGroup
	allowed := 0
	var mu sync.Mutex

	for i := 0; i < 150; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if rl.Allow("client1") {
				mu.Lock()
				allowed++
				mu.Unlock()
			}
		}()
	}

	wg.Wait()

	if allowed != 100 {
		t.Errorf("Expected 100 allowed requests, got %d", allowed)
	}
}

func TestRateLimiter_Reset(t *testing.T) {
	strategy := NewFixedWindowStrategy(2, time.Second)
	rl := NewRateLimiter(strategy)

	rl.Allow("client1")
	rl.Allow("client1")

	if rl.Allow("client1") {
		t.Error("Should be rate limited")
	}

	rl.Reset("client1")
	time.Sleep(10 * time.Millisecond)

	if !rl.Allow("client1") {
		t.Error("Should be allowed after reset")
	}
}
