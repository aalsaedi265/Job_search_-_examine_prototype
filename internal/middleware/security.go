package middleware

import (
	"net/http"
	"sync"
	"time"
)

// SecurityHeaders adds comprehensive security headers to prevent XSS, clickjacking, and other attacks
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Content Security Policy - prevents XSS by restricting resource sources
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline' https://fonts.googleapis.com; style-src 'self' 'unsafe-inline' https://fonts.googleapis.com; font-src 'self' https://fonts.gstatic.com; img-src 'self' data:; connect-src 'self'")

		// X-Frame-Options - prevents clickjacking attacks
		w.Header().Set("X-Frame-Options", "DENY")

		// X-Content-Type-Options - prevents MIME type sniffing
		w.Header().Set("X-Content-Type-Options", "nosniff")

		// X-XSS-Protection - enables browser's XSS filter
		w.Header().Set("X-XSS-Protection", "1; mode=block")

		// Strict-Transport-Security - forces HTTPS connections
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")

		// Referrer-Policy - controls referrer information leakage
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

		// Permissions-Policy - restricts browser features
		w.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

		next.ServeHTTP(w, r)
	})
}

// RateLimiter implements token bucket algorithm to prevent DDoS and brute force attacks
type RateLimiter struct {
	visitors map[string]*visitor
	mu       sync.RWMutex
	rate     int           // requests per window
	window   time.Duration // time window
}

type visitor struct {
	tokens     int
	lastSeen   time.Time
	violations int // Track repeated violations for aggressive attackers
}

func NewRateLimiter(requestsPerMinute int) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*visitor),
		rate:     requestsPerMinute,
		window:   time.Minute,
	}

	// Cleanup stale entries every 5 minutes to prevent memory leaks
	go rl.cleanupVisitors()

	return rl
}

func (rl *RateLimiter) cleanupVisitors() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		for ip, v := range rl.visitors {
			if time.Since(v.lastSeen) > 10*time.Minute {
				delete(rl.visitors, ip)
			}
		}
		rl.mu.Unlock()
	}
}

func (rl *RateLimiter) getVisitor(ip string) *visitor {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	v, exists := rl.visitors[ip]
	if !exists {
		v = &visitor{
			tokens:     rl.rate,
			lastSeen:   time.Now(),
			violations: 0,
		}
		rl.visitors[ip] = v
	}

	// Refill tokens based on time passed
	elapsed := time.Since(v.lastSeen)
	if elapsed > rl.window {
		v.tokens = rl.rate
		v.lastSeen = time.Now()
	}

	return v
}

// Middleware applies rate limiting to prevent DDoS attacks
func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := getIP(r)
		v := rl.getVisitor(ip)

		// Block IPs with excessive violations more aggressively
		if v.violations > 10 {
			http.Error(w, "Too many violations. Temporarily blocked.", http.StatusTooManyRequests)
			return
		}

		if v.tokens <= 0 {
			v.violations++
			w.Header().Set("Retry-After", "60")
			http.Error(w, "Rate limit exceeded. Please try again later.", http.StatusTooManyRequests)
			return
		}

		v.tokens--
		v.lastSeen = time.Now()

		next.ServeHTTP(w, r)
	})
}

// getIP extracts the real IP address from request, handling proxies
func getIP(r *http.Request) string {
	// Check X-Forwarded-For header (but validate to prevent spoofing)
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		// Take the first IP in the chain
		return forwarded
	}

	// Check X-Real-IP header
	if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
		return realIP
	}

	// Fall back to RemoteAddr
	return r.RemoteAddr
}

// MaxBytesMiddleware limits request body size to prevent memory exhaustion attacks
func MaxBytesMiddleware(maxBytes int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
			next.ServeHTTP(w, r)
		})
	}
}
