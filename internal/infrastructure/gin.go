package infrastructure

import (
	"fmt"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"log"
	"net/http"
	"runtime"
	"time"

	"github.com/api-monolith-template/internal/config"
	"github.com/api-monolith-template/internal/constant"
	"github.com/api-monolith-template/internal/model/response"
	transportmw "github.com/api-monolith-template/internal/transport/middleware"
	"github.com/api-monolith-template/internal/util"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func NewGinEngine() *gin.Engine {
	r := gin.New()

	if config.Env.Env == constant.ProductionEnvironment {
		gin.SetMode(gin.ReleaseMode)
	}

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		util.RegisterCustomValidators(v) // pastikan fungsi ini ada di util/validation.go
	}

	// HANYA middleware umum (tidak ada Auth di sini)
	r.Use(gin.Recovery())
	r.Use(transportmw.TraceID())

	// Access log bawaan Gin
	r.Use(gin.Logger())

	// Panic handler
	r.Use(gin.Recovery())

	// TraceID (punyamu)
	r.Use(transportmw.TraceID())

	// (OPSIONAL) Access log custom yang menyertakan trace_id dan user_id
	r.Use(func(c *gin.Context) {
		start := time.Now()
		c.Next()
		latency := time.Since(start)
		traceID := c.GetString("trace_id")
		userID, _ := c.Get("user_id")
		status := c.Writer.Status()
		method := c.Request.Method
		path := c.Request.URL.Path
		clientIP := c.ClientIP()

		// pakai logrus atau log bawaan
		log.Printf("[ACCESS] %s %s %d %s ip=%s user_id=%v trace_id=%s",
			method, path, status, latency, clientIP, userID, traceID)
	})

	corsConfig := cors.Config{
		AllowMethods:           []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders:           []string{"Origin", "Content-Length", "Content-Type", "Authorization", "X-Hospital-ID", "X-Hospital-Code"},
		AllowCredentials:       true,
		AllowWildcard:          false,
		AllowBrowserExtensions: false,
		AllowWebSockets:        true,
		AllowFiles:             false,
	}

	if config.Env.Env == constant.ProductionEnvironment {
		corsConfig.AllowOrigins = []string{
			"https://dashboard-staging.soccernearu.tech",
			"https://www.dashboard-staging.soccernearu.tech",
		}
		if config.Env.Server.FrontendURL != "" {
			corsConfig.AllowOrigins = append(corsConfig.AllowOrigins, config.Env.Server.FrontendURL)
		}
	} else {
		corsConfig.AllowOrigins = []string{
			"http://localhost:3000",
			"http://localhost:3001",
			"http://localhost:8080",
			"https://dashboard-staging.soccernearu.tech",
			"https://www.dashboard-staging.soccernearu.tech",
		}
		if config.Env.Server.FrontendURL != "" {
			corsConfig.AllowOrigins = append(corsConfig.AllowOrigins, config.Env.Server.FrontendURL)
		}
	}

	r.Use(cors.New(corsConfig))

	// register custom validation
	util.AddValidation(DB)

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
