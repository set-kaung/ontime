package internal

import (
	"net"
	"net/http"
	"sync"

	"golang.org/x/time/rate"
)

type SimpleRateLimiter struct {
	ips map[string]*rate.Limiter
	mu  *sync.RWMutex
	r   rate.Limit
	b   int
}

func NewSimpleRateLimiter(r rate.Limit, b int) *SimpleRateLimiter {
	return &SimpleRateLimiter{
		ips: make(map[string]*rate.Limiter),
		mu:  &sync.RWMutex{},
		r:   r,
		b:   b,
	}
}

func (i *SimpleRateLimiter) AddIP(ip string) *rate.Limiter {
	i.mu.Lock()
	defer i.mu.Unlock()

	limiter := rate.NewLimiter(i.r, i.b)
	i.ips[ip] = limiter

	return limiter
}

func (i *SimpleRateLimiter) GetLimiter(ip string) *rate.Limiter {
	i.mu.Lock()
	limiter, exists := i.ips[ip]

	if !exists {
		i.mu.Unlock()
		return i.AddIP(ip)
	}

	i.mu.Unlock()
	return limiter
}

func getIP(r *http.Request) string {

	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {

		if ip := net.ParseIP(forwarded); ip != nil {
			return ip.String()
		}
	}

	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		if ip := net.ParseIP(realIP); ip != nil {
			return ip.String()
		}
	}

	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}

	return ip
}
