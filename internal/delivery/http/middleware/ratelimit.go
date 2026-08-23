package middleware

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

type clientRecord struct {
	tokens    float64
	lastCheck time.Time
}

type IPRateLimiter struct {
	mu       sync.Mutex
	clients  map[string]*clientRecord
	rate     float64       // tokens per second
	capacity float64       // max burst tokens
	cleanup  time.Duration
}

// NewIPRateLimiter creates a token-bucket rate limiter per client IP.
// maxRequests: number of allowed requests per window
// window: time window (e.g. 1 minute)
func NewIPRateLimiter(maxRequests int, window time.Duration) *IPRateLimiter {
	rate := float64(maxRequests) / window.Seconds()
	limiter := &IPRateLimiter{
		clients:  make(map[string]*clientRecord),
		rate:     rate,
		capacity: float64(maxRequests),
		cleanup:  5 * time.Minute,
	}

	// Periodic cleanup of stale client entries
	go limiter.cleanupLoop()

	return limiter
}

func (l *IPRateLimiter) cleanupLoop() {
	ticker := time.NewTicker(l.cleanup)
	defer ticker.Stop()

	for range ticker.C {
		l.mu.Lock()
		now := time.Now()
		for ip, client := range l.clients {
			if now.Sub(client.lastCheck) > l.cleanup {
				delete(l.clients, ip)
			}
		}
		l.mu.Unlock()
	}
}

func (l *IPRateLimiter) allow(ip string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	client, exists := l.clients[ip]
	if !exists {
		l.clients[ip] = &clientRecord{
			tokens:    l.capacity - 1,
			lastCheck: now,
		}
		return true
	}

	// Replenish tokens based on elapsed time
	elapsed := now.Sub(client.lastCheck).Seconds()
	client.lastCheck = now
	client.tokens += elapsed * l.rate
	if client.tokens > l.capacity {
		client.tokens = l.capacity
	}

	if client.tokens >= 1 {
		client.tokens--
		return true
	}

	return false
}

// GetClientIP extracts the real client IP, respecting proxy headers.
func GetClientIP(r *http.Request) string {
	// 1. Check X-Forwarded-For
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		if len(parts) > 0 {
			ip := strings.TrimSpace(parts[0])
			if ip != "" {
				return ip
			}
		}
	}

	// 2. Check X-Real-IP
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		ip := strings.TrimSpace(xri)
		if ip != "" {
			return ip
		}
	}

	// 3. Fallback to RemoteAddr
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}

	return r.RemoteAddr
}

// Limit returns a standard middleware handler that blocks IP requests exceeding the rate limit.
func (l *IPRateLimiter) Limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := GetClientIP(r)
		if !l.allow(ip) {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Retry-After", "10")
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"error":"too_many_requests","message":"Rate limit exceeded. Please slow down."}`))
			return
		}
		next.ServeHTTP(w, r)
	})
}
