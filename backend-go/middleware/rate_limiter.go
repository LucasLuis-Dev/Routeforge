package middleware

import (
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type ipVisitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type IPRateLimiter struct {
	visitors map[string]*ipVisitor
	mu       sync.Mutex
	r        rate.Limit
	b        int
}

func NewIPRateLimiter(r rate.Limit, b int) *IPRateLimiter {
	limiter := &IPRateLimiter{
		visitors: make(map[string]*ipVisitor),
		r:        r,
		b:        b,
	}

	// Goroutine para expurgar visitantes inativos a cada 5 minutos
	go func() {
		for {
			time.Sleep(5 * time.Minute)
			limiter.mu.Lock()
			for ip, visitor := range limiter.visitors {
				if time.Since(visitor.lastSeen) > 5*time.Minute {
					delete(limiter.visitors, ip)
				}
			}
			limiter.mu.Unlock()
		}
	}()

	return limiter
}

func (i *IPRateLimiter) getVisitor(ip string) *rate.Limiter {
	i.mu.Lock()
	defer i.mu.Unlock()

	visitor, exists := i.visitors[ip]
	if !exists {
		limiter := rate.NewLimiter(i.r, i.b)
		i.visitors[ip] = &ipVisitor{limiter: limiter, lastSeen: time.Now()}
		return limiter
	}

	visitor.lastSeen = time.Now()
	return visitor.limiter
}

func (i *IPRateLimiter) Limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			ip = r.RemoteAddr
		}

		forwardedFor := r.Header.Get("X-Forwarded-For")
		if forwardedFor != "" {
			ip = forwardedFor
		}

		limiter := i.getVisitor(ip)
		if !limiter.Allow() {
			w.Header().Set("Retry-After", "60")
			respondError(w, http.StatusTooManyRequests, "limite de requisições excedido. Tente novamente em 1 minuto.")
			return
		}

		next.ServeHTTP(w, r)
	})
}
