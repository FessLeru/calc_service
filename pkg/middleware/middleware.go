// pkg/middleware/middleware.go
package middleware

import (
	"log"
	"net/http"
	"sync"
	"time"
)

// LoggingMiddleware добавляет логирование запросов
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("method=%s path=%s duration=%v", r.Method, r.URL.Path, time.Since(start))
	})
}

// RecoveryMiddleware перехватывает паники
func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("panic recovered: %v", err)
				http.Error(w, "internal server error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// RateLimitMiddleware ограничивает количество запросов
func RateLimitMiddleware(next http.Handler) http.Handler {
	limits := make(map[string]int)
	mu := sync.Mutex{}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()

		ip := r.RemoteAddr
		limits[ip]++

		if limits[ip] > 100 {
			http.Error(w, "too many requests", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}
