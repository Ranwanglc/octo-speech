package admin

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestRateLimiter_AllowsUnderLimit(t *testing.T) {
	rl := &LoginRateLimiter{
		attempts:       make(map[string][]time.Time),
		maxAttempts:    5,
		window:         time.Minute,
		trustedProxies: make(map[string]bool),
	}

	for i := 0; i < 5; i++ {
		if !rl.isAllowed("192.168.1.1") {
			t.Fatalf("expected allowed on attempt %d", i+1)
		}
	}
}

func TestRateLimiter_BlocksOverLimit(t *testing.T) {
	rl := &LoginRateLimiter{
		attempts:       make(map[string][]time.Time),
		maxAttempts:    5,
		window:         time.Minute,
		trustedProxies: make(map[string]bool),
	}

	for i := 0; i < 5; i++ {
		rl.isAllowed("192.168.1.1")
	}

	if rl.isAllowed("192.168.1.1") {
		t.Fatal("expected blocked after exceeding limit")
	}
}

func TestRateLimiter_DifferentIPsIndependent(t *testing.T) {
	rl := &LoginRateLimiter{
		attempts:       make(map[string][]time.Time),
		maxAttempts:    2,
		window:         time.Minute,
		trustedProxies: make(map[string]bool),
	}

	rl.isAllowed("10.0.0.1")
	rl.isAllowed("10.0.0.1")

	if rl.isAllowed("10.0.0.1") {
		t.Fatal("expected 10.0.0.1 blocked")
	}

	if !rl.isAllowed("10.0.0.2") {
		t.Fatal("expected 10.0.0.2 allowed")
	}
}

func TestRateLimiter_WindowExpiry(t *testing.T) {
	rl := &LoginRateLimiter{
		attempts:       make(map[string][]time.Time),
		maxAttempts:    2,
		window:         50 * time.Millisecond,
		trustedProxies: make(map[string]bool),
	}

	rl.isAllowed("192.168.1.1")
	rl.isAllowed("192.168.1.1")

	if rl.isAllowed("192.168.1.1") {
		t.Fatal("expected blocked")
	}

	time.Sleep(60 * time.Millisecond)

	if !rl.isAllowed("192.168.1.1") {
		t.Fatal("expected allowed after window expiry")
	}
}

func TestRateLimiter_TrustedProxy(t *testing.T) {
	rl := &LoginRateLimiter{
		attempts:       make(map[string][]time.Time),
		maxAttempts:    5,
		window:         time.Minute,
		trustedProxies: map[string]bool{"127.0.0.1": true},
	}

	r := gin.New()
	r.POST("/test", rl.Middleware(), func(c *gin.Context) {
		c.String(200, "ok")
	})

	req := httptest.NewRequest("POST", "/test", nil)
	req.RemoteAddr = "127.0.0.1:1234"
	req.Header.Set("X-Real-IP", "203.0.113.50")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestRateLimiter_Middleware429(t *testing.T) {
	rl := &LoginRateLimiter{
		attempts:       make(map[string][]time.Time),
		maxAttempts:    1,
		window:         time.Minute,
		trustedProxies: make(map[string]bool),
	}

	r := gin.New()
	r.POST("/test", rl.Middleware(), func(c *gin.Context) {
		c.String(200, "ok")
	})

	req := httptest.NewRequest("POST", "/test", nil)
	req.RemoteAddr = "192.168.1.1:1234"
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatalf("first request expected 200, got %d", w.Code)
	}

	req = httptest.NewRequest("POST", "/test", nil)
	req.RemoteAddr = "192.168.1.1:1234"
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusTooManyRequests {
		t.Errorf("expected 429, got %d", w.Code)
	}
}
