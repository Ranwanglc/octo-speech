package admin

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type LoginRateLimiter struct {
	mu             sync.Mutex
	attempts       map[string][]time.Time
	maxAttempts    int
	window         time.Duration
	trustedProxies map[string]bool
	stop           chan struct{}
}

func NewLoginRateLimiter(maxAttempts int, window time.Duration, trustedProxies []string) *LoginRateLimiter {
	tp := make(map[string]bool, len(trustedProxies))
	for _, p := range trustedProxies {
		tp[p] = true
	}
	rl := &LoginRateLimiter{
		attempts:       make(map[string][]time.Time),
		maxAttempts:    maxAttempts,
		window:         window,
		trustedProxies: tp,
		stop:           make(chan struct{}),
	}
	go rl.cleanup()
	return rl
}

func (rl *LoginRateLimiter) clientIP(c *gin.Context) string {
	remoteIP, _, _ := net.SplitHostPort(c.Request.RemoteAddr)

	if rl.trustedProxies[remoteIP] {
		if ip := c.GetHeader("X-Real-IP"); ip != "" {
			return ip
		}
		if forwarded := c.GetHeader("X-Forwarded-For"); forwarded != "" {
			if idx := strings.Index(forwarded, ","); idx != -1 {
				return strings.TrimSpace(forwarded[:idx])
			}
			return strings.TrimSpace(forwarded)
		}
	}

	return remoteIP
}

func (rl *LoginRateLimiter) isAllowed(ip string) bool {
	now := time.Now()
	rl.mu.Lock()
	defer rl.mu.Unlock()

	windowStart := now.Add(-rl.window)
	attempts := rl.attempts[ip]
	valid := attempts[:0]
	for _, t := range attempts {
		if t.After(windowStart) {
			valid = append(valid, t)
		}
	}
	rl.attempts[ip] = valid

	if len(valid) >= rl.maxAttempts {
		return false
	}

	rl.attempts[ip] = append(valid, now)
	return true
}

func (rl *LoginRateLimiter) Stop() {
	close(rl.stop)
}

func (rl *LoginRateLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			now := time.Now()
			windowStart := now.Add(-rl.window)
			rl.mu.Lock()
			for ip, attempts := range rl.attempts {
				valid := attempts[:0]
				for _, t := range attempts {
					if t.After(windowStart) {
						valid = append(valid, t)
					}
				}
				if len(valid) == 0 {
					delete(rl.attempts, ip)
				} else {
					rl.attempts[ip] = valid
				}
			}
			rl.mu.Unlock()
		case <-rl.stop:
			return
		}
	}
}

func (rl *LoginRateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := rl.clientIP(c)
		if !rl.isAllowed(ip) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"status": 429,
				"msg":    "too many login attempts, please try again later",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}
