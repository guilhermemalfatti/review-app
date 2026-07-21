package middleware

import (
	"net"
	"net/http"
	"sync"
	"time"
)

type IPRateLimiter struct {
	mu      sync.Mutex
	hits    map[string][]time.Time
	limit   int
	window  time.Duration
	writeErr ErrorWriter
}

func NewIPRateLimiter(limit int, window time.Duration, writeErr ErrorWriter) *IPRateLimiter {
	return &IPRateLimiter{
		hits:     make(map[string][]time.Time),
		limit:    limit,
		window:   window,
		writeErr: writeErr,
	}
}

func (l *IPRateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := clientIP(r)
		now := time.Now()
		cutoff := now.Add(-l.window)

		l.mu.Lock()
		recent := l.hits[ip][:0]
		for _, t := range l.hits[ip] {
			if t.After(cutoff) {
				recent = append(recent, t)
			}
		}
		if len(recent) >= l.limit {
			l.hits[ip] = recent
			l.mu.Unlock()
			l.writeErr(w, http.StatusTooManyRequests, "too many requests")
			return
		}
		l.hits[ip] = append(recent, now)
		l.mu.Unlock()

		next.ServeHTTP(w, r)
	})
}

func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
