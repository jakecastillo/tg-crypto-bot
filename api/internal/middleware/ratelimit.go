package middleware

import (
    "net"
    "net/http"
    "sync"
    "time"
)

// RateLimiter enforces per-IP request rate limits.
type RateLimiter struct {
    limit   int
    window  time.Duration
    visitors map[string]*visitor
    mu      sync.Mutex
}

type visitor struct {
    tokens int
    last   time.Time
}

// NewRateLimiter constructs a new rate limiter with the provided limit per second.
func NewRateLimiter(limit int) *RateLimiter {
    if limit <= 0 {
        limit = 5
    }
    return &RateLimiter{
        limit:   limit,
        window:  time.Second,
        visitors: make(map[string]*visitor),
    }
}

// Middleware enforces the configured rate limits.
func (r *RateLimiter) Middleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
        key := clientIP(req)
        if !r.allow(key) {
            http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
            return
        }
        next.ServeHTTP(w, req)
    })
}

func (r *RateLimiter) allow(key string) bool {
    now := time.Now()
    r.mu.Lock()
    defer r.mu.Unlock()

    v, ok := r.visitors[key]
    if !ok {
        r.visitors[key] = &visitor{tokens: r.limit - 1, last: now}
        return true
    }

    elapsed := now.Sub(v.last)
    tokens := v.tokens + int(elapsed/r.window)*r.limit
    if tokens > r.limit {
        tokens = r.limit
    }
    if tokens <= 0 {
        v.tokens = tokens
        v.last = now
        return false
    }
    v.tokens = tokens - 1
    v.last = now
    return true
}

func clientIP(r *http.Request) string {
    host, _, err := net.SplitHostPort(r.RemoteAddr)
    if err != nil {
        return r.RemoteAddr
    }
    return host
}
