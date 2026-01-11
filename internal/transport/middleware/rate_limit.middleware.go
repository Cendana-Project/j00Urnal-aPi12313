package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/api-monolith-template/internal/model/response"
	"github.com/api-monolith-template/internal/util"
	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// IPRateLimiter menyimpan limiter per IP
type IPRateLimiter struct {
	ips map[string]*rate.Limiter
	mu  *sync.RWMutex
	r   rate.Limit
	b   int
}

// NewIPRateLimiter membuat instance limiter baru (r = rate per second, b = burst size)
func NewIPRateLimiter(r rate.Limit, b int) *IPRateLimiter {
	return &IPRateLimiter{
		ips: make(map[string]*rate.Limiter),
		mu:  &sync.RWMutex{},
		r:   r,
		b:   b,
	}
}

// GetLimiter mengambil atau membuat limiter untuk IP tertentu
func (i *IPRateLimiter) GetLimiter(ip string) *rate.Limiter {
	i.mu.Lock()
	defer i.mu.Unlock()

	limiter, exists := i.ips[ip]
	if !exists {
		limiter = rate.NewLimiter(i.r, i.b)
		i.ips[ip] = limiter
	}

	return limiter
}

// RateLimitMiddleware membatasi request per IP
func RateLimitMiddleware() gin.HandlerFunc {
	// Konfigurasi: 5 request per detik, burst 10 (cukup longgar untuk API jurnal)
	// Bisa dipindahkan ke config jika perlu.
	limiter := NewIPRateLimiter(5, 10)

	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		if !limiter.GetLimiter(clientIP).Allow() {
			resp := response.BaseResponse{
				StatusCode: http.StatusTooManyRequests,
				Message:    "Too Many Requests",
				MessageDetail: response.MessageDetail{
					TitleEng: "Too Many Requests",
					DescEng:  "Please try again later.",
					TitleIdn: "Terlalu Banyak Permintaan",
					DescIdn:  "Silakan coba lagi nanti.",
				},
				Timestamp: time.Now().UTC(),
				TraceID:   util.GetTraceID(c),
			}
			c.AbortWithStatusJSON(http.StatusTooManyRequests, resp)
			return
		}

		c.Next()
	}
}
