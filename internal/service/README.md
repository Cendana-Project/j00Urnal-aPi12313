# Service

This directory contains the business logic of the application. This layer implements application use cases by orchestrating calls to repositories and other services.

## Subdirectories

- **auth/**: Contains implementations for authentication and authorization business logic.

## Purpose

The service layer:

1. Implements business rules and domain logic
2. Orchestrates operations across multiple repositories
3. Handles transaction boundaries
4. Performs data validation and transformation
5. Implements application-specific workflows

## Example Code

### Base Service
```go
// Auth service structure
type Service struct {
	userRepository  contract.UserRepository
	cacheRepository contract.CacheRepository
}

// Create new service
func NewService() *Service {
	return new(Service)
}

// Set user repository dependency
func (s *Service) WithUserRepository(repo contract.UserRepository) *Service {
	s.userRepository = repo
	return s
}

// Set cache repository dependency
func (s *Service) WithCacheRepository(repo contract.CacheRepository) *Service {
	s.cacheRepository = repo
	return s
}
```

### Login Implementation
```go
// Handle user login process
func (s *Service) Login(ctx context.Context, req *request.LoginReq) (*response.BaseResponse, error) {
	// Get user by identifier (email or username)
	user, err := s.userRepository.FindByIdentifier(ctx, req.Identifier)
	if err != nil {
		return response.NewErrorResponse(constant.ResponseUserNotFound, nil, nil), nil
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		return response.NewErrorResponse("invalid credentials", nil, nil), nil
	}

	// Generate tokens
	accessToken, refreshToken, err := s.generateTokens(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	// Create response
	return response.NewSuccessResponse("ok", &response.TokenResponse{
		AccessToken:           accessToken.Token,
		AccessTokenExpiredAt:  accessToken.ExpiredAt,
		RefreshToken:          refreshToken.Token,
		RefreshTokenExpiredAt: refreshToken.ExpiredAt,
	}), nil
}
```