package infrastructure

import (
	"fmt"
	"net/http"
	"runtime"
	"time"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"

	"github.com/api-monolith-template/internal/config"
	"github.com/api-monolith-template/internal/constant"
	"github.com/api-monolith-template/internal/model/response"
	transportmw "github.com/api-monolith-template/internal/transport/middleware"
	"github.com/api-monolith-template/internal/util"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func NewGinEngine() *gin.Engine {
	r := gin.New()

	// Set Gin mode based on environment
	if config.Env.Env == constant.ProductionEnvironment {
		gin.SetMode(gin.ReleaseMode)
	}

	// Register custom validators
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		util.RegisterCustomValidators(v)
	}

	// Security headers middleware (must be first)
	r.Use(securityHeaders())

	// Panic recovery middleware
	r.Use(gin.Recovery())

	// TraceID middleware (for request tracing)
	r.Use(transportmw.TraceID())

	// Access logging middleware (structured logging with logrus)
	r.Use(accessLogMiddleware())

	// CORS middleware
	corsConfig := buildCORSConfig()
	r.Use(cors.New(corsConfig))

	// Public health check
	internalGroup := r.Group("/_internal")
	internalGroup.GET("/healthz", func(c *gin.Context) {
		var memStats runtime.MemStats
		runtime.ReadMemStats(&memStats)
		status := "healthy"

		serviceStatuses := make([]response.GetHealthCheckServiceStatusResp, 0)
		for serviceName, healthCheckFn := range MapHealthCheck {
			err := healthCheckFn(c.Request.Context())
			if err != nil {
				status = "unhealthy"
			}
			serviceStatuses = append(serviceStatuses, response.GetHealthCheckServiceStatusResp{
				Name: serviceName,
				IsUp: err == nil,
			})
		}

		healthInfo := response.GetHealthCheckResp{
			Status:       status,
			Environtment: config.Env.Env,
			Version:      fmt.Sprintf("%s@%s", config.ServiceName, config.ServiceVersion),
			GoVersion:    runtime.Version(),
			GoRoutine:    runtime.NumGoroutine(),
			Memory: response.GetHealthCheckMemoryResp{
				Alloc:      memStats.Alloc,
				TotalAlloc: memStats.TotalAlloc,
				Sys:        memStats.Sys,
				HeapAlloc:  memStats.HeapAlloc,
				HeapSys:    memStats.HeapSys,
			},
			ServiceStatuses: serviceStatuses,
		}

		resp := response.BaseResponse{
			StatusCode: http.StatusOK,
			Data:       healthInfo,
		}
		util.HandleResponse(c, &resp, nil)
	})

	return r
}

// securityHeaders adds security headers to all responses
func securityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Prevent MIME type sniffing
		c.Header("X-Content-Type-Options", "nosniff")
		// Prevent clickjacking
		c.Header("X-Frame-Options", "DENY")
		// XSS Protection
		c.Header("X-XSS-Protection", "1; mode=block")
		// Referrer Policy
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		// Content Security Policy (adjust based on your needs)
		if config.Env.Env == constant.ProductionEnvironment {
			c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline';")
		}
		c.Next()
	}
}

// accessLogMiddleware provides structured access logging with logrus
func accessLogMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		traceID := c.GetString("trace_id")
		userID, exists := c.Get("user_id")
		userIDStr := ""
		if exists {
			userIDStr = fmt.Sprintf("%v", userID)
		}

		fields := logrus.Fields{
			"status":     c.Writer.Status(),
			"method":     c.Request.Method,
			"path":       path,
			"query":      query,
			"ip":         c.ClientIP(),
			"user_agent": c.Request.UserAgent(),
			"latency":    latency,
			"trace_id":   traceID,
		}

		if userIDStr != "" {
			fields["user_id"] = userIDStr
		}

		// Log level based on status code
		if c.Writer.Status() >= 500 {
			logrus.WithFields(fields).Error("HTTP request")
		} else if c.Writer.Status() >= 400 {
			logrus.WithFields(fields).Warn("HTTP request")
		} else {
			logrus.WithFields(fields).Info("HTTP request")
		}
	}
}

// buildCORSConfig builds CORS configuration based on environment
func buildCORSConfig() cors.Config {
	corsCfg := cors.Config{
		AllowMethods:           []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders:           []string{"Origin", "Content-Length", "Content-Type", "Authorization", "X-Hospital-ID", "X-Hospital-Code"},
		AllowCredentials:       true,
		AllowWildcard:          false,
		AllowBrowserExtensions: false,
		AllowWebSockets:        true,
		AllowFiles:             false,
		MaxAge:                 12 * time.Hour,
	}

	// In production, restrict origins based on environment variable
	if config.Env.Env == constant.ProductionEnvironment {
		frontendURL := config.Env.Server.FrontendURL
		if frontendURL != "" {
			// Allow specific frontend URL
			corsCfg.AllowOrigins = []string{frontendURL}
			corsCfg.AllowOriginFunc = nil
		} else {
			// Fallback: allow all origins but log warning
			logrus.Warn("CORS: FrontendURL not set, allowing all origins (not recommended for production)")
			corsCfg.AllowOriginFunc = func(origin string) bool {
				return true
			}
		}
	} else {
		// Development: allow all origins
		corsCfg.AllowOriginFunc = func(origin string) bool {
			return true
		}
	}

	return corsCfg
}
