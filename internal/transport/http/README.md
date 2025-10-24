# HTTP Transport

This directory contains HTTP communication implementation using the Gin framework. It defines how the application exposes its functionality via HTTP.

## Main Files

- **http.transport.go**: Contains the main HTTP transport setup and initialization.
- **route.transport.go**: Defines HTTP routes and endpoints for the application.

## Subdirectories

- **auth/**: Contains HTTP handlers for authentication-related endpoints.
- **middleware/**: Contains HTTP middleware implementations for cross-cutting concerns.

## Purpose

The HTTP transport layer:

1. Defines HTTP routes and endpoints
2. Handles HTTP requests and responses
3. Converts between HTTP-specific formats and domain models
4. Manages HTTP-specific concerns (status codes, headers, cookies)
5. Implements HTTP middleware for cross-cutting concerns

## Example Code

### Transport Setup
```go
type Transport struct {
	router *gin.Engine
	authController       *auth.Controller
	middlewareController *middleware.Controller
}

func NewTransport() *Transport {
	return new(Transport)
}

func (t *Transport) WithGinEngine(r *gin.Engine) *Transport {
	t.router = r
	return t
}

// InitRoute initializes all HTTP routes
func (t *Transport) InitRoute() {
	api := t.router.Group("/api")
	v1 := api.Group("/v1")
    
    // Initialize routes
    t.initAuthRoutes(v1)
}
```

### Route Definition
```go
// initAuthRoutes defines authentication routes
func (t *Transport) initAuthRoutes(v1Group *gin.RouterGroup) {
	authGroup := v1Group.Group("/auth")
	
	// Public endpoints
	authGroup.POST("/register", t.authController.Register)
	authGroup.POST("/login", t.authController.Login)
	
	// Endpoints requiring refresh token
	authRefreshToken := authGroup.Group("/refresh", 
	    t.middlewareController.AuthMiddleware(constant.RefreshTokenType))
	authRefreshToken.POST("/", t.authController.RefreshToken)

	// Endpoints requiring access token
	authProtected := authGroup.Use(
	    t.middlewareController.AuthMiddleware(constant.AccessTokenType))
	authProtected.GET("/info", t.authController.Info)
	authProtected.POST("/logout", t.authController.Logout)
}
```

### HTTP Handler
```go
// Login handles HTTP login requests
func (c *Controller) Login(ctx *gin.Context) {
	var req request.LoginReq

	// Bind JSON request body to request struct
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid request",
			"error":   err.Error(),
		})
		return
	}

	// Call service to handle login
	response, err := c.authService.Login(ctx, &req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "internal server error",
			"error":   err.Error(),
		})
		return
	}

	// Return successful or error response
	if response.IsSuccess() {
		ctx.JSON(http.StatusOK, response)
	} else {
		ctx.JSON(http.StatusBadRequest, response)
	}
}
```