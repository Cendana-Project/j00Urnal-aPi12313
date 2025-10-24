# Monolith Service

## Response
```json
{
	"message": "validation error", // string
	"data": {
		"id": "6759e56a-fa7c-49ee-9854-f32ab38083ae",
		"username": "username17",
		"email": "email17@gmail.com",
		"level": "USER",
		"createdAt": "2024-06-16T10:46:25Z",
		"updatedAt": "2024-06-16T10:46:25Z"
	}, // object | array of object | null
	"validationErrors": [
		{
			"field": "username",
			"tag": "required",
			"message": "Key: 'RegisterReq.Username' Error:Field validation for 'username' failed on the 'required' tag"
		}
	] // array of object | null
}
```

## API Contract

### Register
#### Request

**Method:** `POST`

**URL:** `${{HOST}}/v1/auth/register`

**Headers:**
- `Content-Type: application/json`

**Body:**
```json
{
	"username":"username",
	"email": "email@gmail.com",
	"password": "strongpassword"
}
```

**Example cURL Command:**

```bash
curl --request POST \
  --url ${{HOST}}/v1/auth/register \
  --header 'Content-Type: application/json' \
  --data '{
	"username":"username",
	"email": "email@gmail.com",
	"password": "strongpassword"
}'
```

**Example Response:**
```json
{
	"message": "validation error",
	"data": null,
	"validationErrors": [
		{
			"field": "username",
			"tag": "unique_db",
			"message": "username already taken"
		},
		{
			"field": "email",
			"tag": "unique_db",
			"message": "email already taken"
		}
	]
}
```

```json
{
	"message": "ok",
	"data": null,
	"validationErrors": null
}
```

### Login
#### Request

**Method:** `POST`

**URL:** `${{HOST}}/v1/auth/login`

**Headers:**
- `Content-Type: application/json`

**Body:**
```json
{
	"identifier":"username",
	"password": "strongpassword"
}
```

**Example cURL Command:**

```bash
curl --request POST \
  --url ${{HOST}}/v1/auth/login \
  --header 'Content-Type: application/json' \
  --data '{
	"identifier":"username",
	"password": "strongpassword"
}'
```

**Example Response:**
```json
{
	"message": "user not found",
	"data": null,
	"validationErrors": null
}
```

```json
{
	"message": "ok",
	"data": {
		"accessToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VySWQiOiI2NzU5ZTU2YS1mYTdjLTQ5ZWUtOTg1NC1mMzJhYjM4MDgzYWUiLCJleHAiOjE3MTg2MTk3ODQsImlhdCI6MTcxODYxNjE4NCwianRpIjoiYzViZjAyZTctNmNkMy00MjZiLThiMjctYzk2MTUyZjc2NmU4In0.aaUAM7Hl6Z-H8kzdnrLedVmmVJEuglxes7xQYHt1HKI",
		"accessTokenExpiredAt": "2024-06-17T09:23:04Z",
		"refreshToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VySWQiOiI2NzU5ZTU2YS1mYTdjLTQ5ZWUtOTg1NC1mMzJhYjM4MDgzYWUiLCJleHAiOjE3MTg3MjQxODQsImlhdCI6MTcxODYxNjE4NCwianRpIjoiNzllNTVkZDgtMzczMS00OWU2LThjZDItNzMxNDI0MzYzZjZjIn0.KQOZGZxz8-8JiJv68Xpdj-7z1Dp6dLe0a4IC0nZ5WcA",
		"refreshTokenExpiredAt": "2024-06-17T09:23:04Z"
	},
	"validationErrors": null
}
```

### Refresh Token
#### Request

**Method:** `POST`

**URL:** `${{HOST}}/v1/auth/refresh`

**Headers:**
- `Content-Type: application/json`
- `Authorization: Bearer <REFRESH_TOKEN>`

**Example cURL Command:**

```bash
curl --request POST \
  --url ${{HOST}}/v1/auth/refresh \
  --header 'Content-Type: application/json' \
  --header 'Authorization: Bearer <REFRESH_TOKEN>'
```

**Example Response:**
```json
{
	"message": "invalid token",
	"data": null,
	"validationErrors": null
}
```

```json
{
	"message": "ok",
	"data": {
		"accessToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VySWQiOiI2NzU5ZTU2YS1mYTdjLTQ5ZWUtOTg1NC1mMzJhYjM4MDgzYWUiLCJleHAiOjE3MTg2MTk3ODQsImlhdCI6MTcxODYxNjE4NCwianRpIjoiYzViZjAyZTctNmNkMy00MjZiLThiMjctYzk2MTUyZjc2NmU4In0.aaUAM7Hl6Z-H8kzdnrLedVmmVJEuglxes7xQYHt1HKI",
		"accessTokenExpiredAt": "2024-06-17T09:23:04Z",
		"refreshToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VySWQiOiI2NzU5ZTU2YS1mYTdjLTQ5ZWUtOTg1NC1mMzJhYjM4MDgzYWUiLCJleHAiOjE3MTg3MjQxODQsImlhdCI6MTcxODYxNjE4NCwianRpIjoiNzllNTVkZDgtMzczMS00OWU2LThjZDItNzMxNDI0MzYzZjZjIn0.KQOZGZxz8-8JiJv68Xpdj-7z1Dp6dLe0a4IC0nZ5WcA",
		"refreshTokenExpiredAt": "2024-06-17T09:23:04Z"
	},
	"validationErrors": null
}
```

### Logout
#### Request

**Method:** `POST`

**URL:** `${{HOST}}/v1/auth/logout`

**Headers:**
- `Content-Type: application/json`
- `Authorization: Bearer <ACCESS_TOKEN>`

**Example cURL Command:**

```bash
curl --request POST \
  --url ${{HOST}}/v1/auth/logout \
  --header 'Content-Type: application/json' \
  --header 'Authorization: Bearer <ACCESS_TOKEN>'
```

**Example Response:**
```json
{
	"message": "invalid token",
	"data": null,
	"validationErrors": null
}
```

```json
{
	"message": "ok",
	"data": null,
	"validationErrors": null
}
```

### User Info
#### Request

**Method:** `GET`

**URL:** `${{HOST}}/v1/auth/info`

**Headers:**
- `Content-Type: application/json`
- `Authorization: Bearer <ACCESS_TOKEN>`

**Example cURL Command:**

```bash
curl --request GET \
  --url ${{HOST}}/v1/auth/info \
  --header 'Authorization: Bearer <ACCESS_TOKEN>'
```

**Example Response:**
```json
{
	"message": "ok",
	"data": {
		"id": "6759e56a-fa7c-49ee-9854-f32ab38083ae",
		"username": "username",
		"email": "email@gmail.com",
		"level": "USER",
		"createdAt": "2024-06-16T10:46:25Z",
		"updatedAt": "2024-06-16T10:46:25Z"
	},
	"validationErrors": null
}
```

### User Management Endpoints

### Get All Users
#### Request

**Method:** `GET`

**URL:** `${{HOST}}/v1/users`

**Headers:**
- `Content-Type: application/json`

**Example cURL Command:**

```bash
curl --request GET \
  --url ${{HOST}}/v1/users \
  --header 'Content-Type: application/json'
```

**Example Response:**
```json
{
	"message": "ok",
	"data": {
		"users": [
			{
				"id": "6759e56a-fa7c-49ee-9854-f32ab38083ae",
				"username": "username",
				"firstName": "First",
				"lastName": "Last",
				"email": "email@gmail.com",
				"phone": "1234567890",
				"location": "Jakarta",
				"createdAt": "2024-06-16T10:46:25Z",
				"updatedAt": "2024-06-16T10:46:25Z"
			}
		],
		"total": 1
	},
	"validationErrors": null
}
```

### Get User by ID
#### Request

**Method:** `GET`

**URL:** `${{HOST}}/v1/users/:id`

**Headers:**
- `Content-Type: application/json`

**Example cURL Command:**

```bash
curl --request GET \
  --url ${{HOST}}/v1/users/6759e56a-fa7c-49ee-9854-f32ab38083ae \
  --header 'Content-Type: application/json'
```

**Example Response:**
```json
{
	"message": "ok",
	"data": {
		"id": "6759e56a-fa7c-49ee-9854-f32ab38083ae",
		"username": "username",
		"firstName": "First",
		"lastName": "Last",
		"email": "email@gmail.com",
		"phone": "1234567890",
		"location": "Jakarta",
		"createdAt": "2024-06-16T10:46:25Z",
		"updatedAt": "2024-06-16T10:46:25Z"
	},
	"validationErrors": null
}
```

### Find User by Email
#### Request

**Method:** `GET`

**URL:** `${{HOST}}/v1/users/email/:email`

**Headers:**
- `Content-Type: application/json`

**Example cURL Command:**

```bash
curl --request GET \
  --url ${{HOST}}/v1/users/email/email@gmail.com \
  --header 'Content-Type: application/json'
```

**Example Response:**
```json
{
	"message": "ok",
	"data": {
		"id": "6759e56a-fa7c-49ee-9854-f32ab38083ae",
		"username": "username",
		"firstName": "First",
		"lastName": "Last",
		"email": "email@gmail.com",
		"phone": "1234567890",
		"location": "Jakarta",
		"createdAt": "2024-06-16T10:46:25Z",
		"updatedAt": "2024-06-16T10:46:25Z"
	},
	"validationErrors": null
}
```

### Find User by Username
#### Request

**Method:** `GET`

**URL:** `${{HOST}}/v1/users/username/:username`

**Headers:**
- `Content-Type: application/json`

**Example cURL Command:**

```bash
curl --request GET \
  --url ${{HOST}}/v1/users/username/username \
  --header 'Content-Type: application/json'
```

**Example Response:**
```json
{
	"message": "ok",
	"data": {
		"id": "6759e56a-fa7c-49ee-9854-f32ab38083ae",
		"username": "username",
		"firstName": "First",
		"lastName": "Last",
		"email": "email@gmail.com",
		"phone": "1234567890",
		"location": "Jakarta",
		"createdAt": "2024-06-16T10:46:25Z",
		"updatedAt": "2024-06-16T10:46:25Z"
	},
	"validationErrors": null
}
```

### Find User by Identifier
#### Request

**Method:** `GET`

**URL:** `${{HOST}}/v1/users/identifier/:identifier`

**Headers:**
- `Content-Type: application/json`

**Example cURL Command:**

```bash
curl --request GET \
  --url ${{HOST}}/v1/users/identifier/username \
  --header 'Content-Type: application/json'
```

**Example Response:**
```json
{
	"message": "ok",
	"data": {
		"id": "6759e56a-fa7c-49ee-9854-f32ab38083ae",
		"username": "username",
		"firstName": "First",
		"lastName": "Last",
		"email": "email@gmail.com",
		"phone": "1234567890",
		"location": "Jakarta",
		"createdAt": "2024-06-16T10:46:25Z",
		"updatedAt": "2024-06-16T10:46:25Z"
	},
	"validationErrors": null
}
```

### Update User
#### Request

**Method:** `POST`

**URL:** `${{HOST}}/v1/users`

**Headers:**
- `Content-Type: application/json`
- `Authorization: Bearer <ACCESS_TOKEN>`

**Body:**
```json
{
	"username": "newusername",
	"firstName": "New",
	"lastName": "Name",
	"email": "newemail@gmail.com",
	"phone": "1234567890",
	"location": "Jakarta"
}
```

**Example cURL Command:**

```bash
curl --request POST \
  --url ${{HOST}}/v1/users \
  --header 'Content-Type: application/json' \
  --header 'Authorization: Bearer <ACCESS_TOKEN>' \
  --data '{
	"username": "newusername",
	"firstName": "New",
	"lastName": "Name",
	"email": "newemail@gmail.com",
	"phone": "1234567890",
	"location": "Jakarta"
}'
```

**Example Response:**
```json
{
	"message": "Username or email already exists",
	"data": null,
	"validationErrors": null
}
```

```json
{
	"message": "ok",
	"data": {
		"id": "6759e56a-fa7c-49ee-9854-f32ab38083ae",
		"username": "newusername",
		"email": "newemail@gmail.com",
		"updatedAt": "2024-06-16T10:46:25Z"
	},
	"validationErrors": null
}
```

## Create new domain

### Define contract

Define your contract domain (repository and service) on /internal/model/contract/<DOMAIN>.contract.go, example:

```go
package contract

import (
	"context"

	"github.com/api-monolith-template/internal/model/entity"
    "github.com/api-monolith-template/internal/model/request"
	"github.com/api-monolith-template/internal/model/response"
	"github.com/google/uuid"
)

type UserRepository interface {
	FindByEmail(ctx context.Context, email string) (*entity.User, error)
	FindByUsername(ctx context.Context, username string) (*entity.User, error)
	FindByIdentifier(ctx context.Context, identifier string) (*entity.User, error)
	FindByID(ctx context.Context, id uuid.UUID) (*entity.User, error)
	Upsert(ctx context.Context, user *entity.User) error
}

type AuthService interface {
	Register(ctx context.Context, req *request.RegisterReq) (*response.BaseResponse, error)
	Login(ctx context.Context, req *request.LoginReq) (*response.BaseResponse, error)
	RefreshToken(ctx context.Context, req *request.AuthRefreshReq) (*response.BaseResponse, error)
	Info(ctx context.Context, req *request.AuthInfoReq) (*response.BaseResponse, error)
	Logout(ctx context.Context, req *request.AuthLogoutReq) (*response.BaseResponse, error)
}

```

### Create your repository

Create your base repository on /internal/repository/<DOMAIN>/<DOMAIN>.repository.go, example:

```go
package user

import (
	"github.com/api-monolith-template/internal/model/contract"
	"gorm.io/gorm"
)

type Repository struct {
	db        *gorm.DB
	cacheRepo contract.CacheRepository
}

func NewRepository() *Repository {
	return new(Repository)
}

func (r *Repository) WithGormDB(db *gorm.DB) *Repository {
	r.db = db
	return r
}

func (r *Repository) WithCacheRepository(repo contract.CacheRepository) *Repository {
	r.cacheRepo = repo
	return r
}
```

### Create your service

Create your base service on /internal/service/<DOMAIN>/<DOMAIN>.service.go, example:

```go
package auth

import "github.com/api-monolith-template/internal/model/contract"

type Service struct {
	userRepository  contract.UserRepository
	cacheRepository contract.CacheRepository
}

func NewService() *Service {
	return new(Service)
}

func (s *Service) WithUserRepository(repo contract.UserRepository) *Service {
	s.userRepository = repo
	return s
}

func (s *Service) WithCacheRepository(repo contract.CacheRepository) *Service {
	s.cacheRepository = repo
	return s
}
```

### Create your transport layer

Create base http transport layer on /internal/transport/http/<DOMAIN>/<DOMAIN>.http_transport.go, example:

```go
package auth

import "github.com/api-monolith-template/internal/model/contract"

type Controller struct {
	authService contract.AuthService
}

func NewController() *Controller {
	return new(Controller)
}

func (c *Controller) WithAuthService(svc contract.AuthService) *Controller {
	c.authService = svc
	return c
}
```

inject all your transport domain on /internal/transport/http/http.transport.go

```go
package http

import (
	"github.com/api-monolith-template/internal/transport/http/auth"
	"github.com/api-monolith-template/internal/transport/http/middleware"
	"github.com/gin-gonic/gin"
)

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

func (t *Transport) WithAuthController(c *auth.Controller) *Transport {
	t.authController = c
	return t
}

func (t *Transport) WithMiddlewareController(c *middleware.Controller) *Transport {
	t.middlewareController = c
	return t
}

```

after create http transport layer, create http route on /internal/transport/http/route.transport.go

```go
authGroup := v1Group.Group("/auth")
authGroup.POST("/register", t.authController.Register)
authGroup.POST("/login", t.authController.Login)
authRefreshToken := authGroup.Group("/refresh", t.middlewareController.AuthMiddleware(constant.RefreshTokenType))
authRefreshToken.POST("/", t.authController.RefreshToken)

authProtected := authGroup.Use(t.middlewareController.AuthMiddleware(constant.AccessTokenType))
authProtected.GET("/info", t.authController.Info)
```

### Inject all new depedency

after create a domain for each layer, now init new domain and inject to all layer

```go
// init repository
cacheRepository := cacheRepo.
    NewRepository().
    WithRedisDB(rdb)
userRepository := userRepo.
    NewRepository().
    WithGormDB(infrastructure.DB).
    WithCacheRepository(cacheRepository)

// init service
authService := authSvc.
    NewService().
    WithUserRepository(userRepository).
    WithCacheRepository(cacheRepository)

// init controller
middlewareController := middlewareCtrl.
    NewController().
    WithAuthService(authService).
    WithCacheRepository(cacheRepository)
authController := authCtrl.
    NewController().
    WithAuthService(authService)

// init http transport
httpTransport.
    NewTransport().
    WithGinEngine(r).
    WithMiddlewareController(middlewareController).
    WithAuthController(authController).
    InitRoute()
```

# Role-Based Access Control API

This repository implements a role-based access control (RBAC) system for a Go application, allowing for fine-grained permission management based on user roles within specific buildings.

## Features

- User authentication (register, login, logout)
- Role management (create, update, delete roles)
- Building-specific role assignments (assign/remove roles to users in specific buildings)
- Permission-based access control with granular permissions (e.g., "owner:show", "owner:create")
- Role hierarchy (Admin, Manager, User)

## Testing the API

### Prerequisites

- Ensure PostgreSQL is installed and running
- Create a database named "Medikaone"
- Set up the configuration in `config.yml`

### Setup

1. Initialize and seed the database:

```bash
make seed
```

This creates:
- Default roles (Admin, Manager, User)
- Default permissions (owner:show, owner:create, user:list)
- Test users with passwords "password123":
  - admin@example.com (Admin role)
  - manager@example.com (Manager role)
  - user@example.com (User role)
- Test buildings (Building A, Building B)
- Role assignments for users in buildings

2. Start the server:

```bash
make run
```

### Testing with Postman

Import the provided Postman collection:

1. Open Postman
2. Click "Import" and select the `role_management_api.postman_collection.json` file
3. Create a Postman environment with the variables:
   - `access_token` (leave it empty, it will be filled automatically)
   - `refresh_token` (leave it empty, it will be filled automatically)
   - `user_id` (default: bf7ad1c8-a873-4915-9e60-2cd15b451292)
   - `role_id` (default: e52b1dac-7751-451c-98d5-f81401926cf7)
   - `building_id` (default: e0ffcd6c-a2f2-453f-801e-cbb351850932)
4. Start with the "Login" request to get an access token
5. After successful login, the environment variables will be automatically updated
6. Test the other endpoints as needed

### Testing Flow

1. Login as admin user
2. Create a new role
3. Assign the role to a user in a building
4. Verify the role assignment by fetching user roles for that building
5. Remove the role from the user
6. Verify the role was removed

## Test Users and Default IDs

After seeding, the following entities are available:

### Users
- Admin User: `bf7ad1c8-a873-4915-9e60-2cd15b451292`
- Manager User: `9eff1130-2aa4-40f8-a3f1-cb3d461b6682`
- Regular User: `40a91ed3-5057-4bca-be65-91c7da59feca`

### Roles
- Admin Role: `e52b1dac-7751-451c-98d5-f81401926cf7`
- Manager Role: `63de13e6-9847-4b6c-bcc4-20145d0e1bec`
- User Role: `b7efdfce-bc6a-451a-9b92-28b42e6eb3bc`

### Buildings
- Building A: `e0ffcd6c-a2f2-453f-801e-cbb351850932`
- Building B: `1f66bcdc-8e1e-40a0-a037-1b364c70ac79`

Note: The actual UUIDs may vary based on your seeded data.
