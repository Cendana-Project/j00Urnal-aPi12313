# Constants

This directory contains constants and enums used throughout the application code. It provides a central location for all fixed values.

## Main Files

- **common.go**: Defines common constants used in various parts of the application.
- **custom_error.constant.go**: Contains error constants and error codes for standardized error handling.
- **user.constant.go**: Defines user-related constants such as user role codes or status codes.

## Purpose

The constant layer:

1. Provides a central location for all application constants
2. Ensures consistency of string literals, error codes, and other values used throughout the application
3. Makes code more maintainable by avoiding hardcoded values
4. Defines type-safe enums and constants for better code quality

## Example Code

### common.go
```go
package constant

// Token type constants
const (
    AccessTokenType  = "access_token"
    RefreshTokenType = "refresh_token"
)

// Time format constants
const (
    TimeFormat = "2006-01-02T15:04:05Z"
)

// Response message constants
const (
    ResponseOK             = "ok"
    ResponseInvalidToken   = "invalid token"
    ResponseUserNotFound   = "user not found"
    ResponseValidationError = "validation error"
)
```

### user.constant.go
```go
package constant

// User level constants
const (
    UserLevelAdmin = "ADMIN"
    UserLevelUser  = "USER"
)

// User validation error messages
const (
    UserErrorUsernameRequired = "username is required"
    UserErrorEmailRequired    = "email is required"
    UserErrorPasswordRequired = "password is required"
    UserErrorUsernameAlreadyTaken = "username already taken"
    UserErrorEmailAlreadyTaken    = "email already taken"
)
```

### custom_error.constant.go
```go
package constant

// Error codes for different types of errors
const (
    ErrorCodeNotFound        = "ERR_NOT_FOUND"
    ErrorCodeUnauthorized    = "ERR_UNAUTHORIZED"
    ErrorCodeValidation      = "ERR_VALIDATION"
    ErrorCodeDatabase        = "ERR_DATABASE"
    ErrorCodeInternal        = "ERR_INTERNAL"
)

// Custom error types
const (
    ErrorTypeDatabase  = "DATABASE_ERROR"
    ErrorTypeValidation = "VALIDATION_ERROR"
    ErrorTypeAuth      = "AUTHENTICATION_ERROR"
)
```