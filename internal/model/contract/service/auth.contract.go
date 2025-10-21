package service

import (
	"context"

	"github.com/api-monolith-template/internal/model/request"
	"github.com/api-monolith-template/internal/model/response"
)

type AuthService interface {
	// Register akun (generate PIN, kirim email)
	Register(ctx context.Context, req *request.RegisterRequest) (*response.RegisterResponse, error)

	// Verifikasi PIN 6 digit yang dikirim via email
	VerifyPIN(ctx context.Context, email, pin string) (*response.VerifyEmailResponse, error)
}
