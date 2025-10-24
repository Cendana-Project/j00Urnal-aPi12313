package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/api-monolith-template/internal/model/response"
	"github.com/api-monolith-template/internal/util"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type RateLimiter struct {
	requests map[string][]time.Time
	mutex    sync.RWMutex
	limit    int
	window   time.Duration
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}
}

func (rl *RateLimiter) RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.Next()
			return
		}

		userIDStr := userID.(string)
		now := time.Now()

		rl.mutex.Lock()
		defer rl.mutex.Unlock()

		if requests, exists := rl.requests[userIDStr]; exists {
			var validRequests []time.Time
			for _, reqTime := range requests {
				if now.Sub(reqTime) <= rl.window {
					validRequests = append(validRequests, reqTime)
				}
			}
			rl.requests[userIDStr] = validRequests
		}

		if len(rl.requests[userIDStr]) >= rl.limit {
			logrus.WithFields(logrus.Fields{
				"userID": userIDStr,
				"limit":  rl.limit,
				"window": rl.window,
			}).Warn("Rate limit exceeded for user")

			resp := response.BaseResponse{
				StatusCode: http.StatusTooManyRequests,
				Message:    "RATE_LIMIT_EXCEEDED",
				MessageDetail: response.MessageDetail{
					TitleEng: "Rate limit exceeded",
					TitleIdn: "Silakan tunggu sebelum mencoba lagi",
				},
			}
			util.HandleResponse(c, &resp, nil)
			c.Abort()
			return
		}

		rl.requests[userIDStr] = append(rl.requests[userIDStr], now)

		c.Next()
	}
}

func UploadRateLimit() gin.HandlerFunc {
	limiter := NewRateLimiter(10, time.Minute)
	return limiter.RateLimit()
}
