package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/api-monolith-template/internal/config"
)

type Service struct {
	client *http.Client
}

func NewService() *Service {
	return &Service{
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

// Upload uploads a file to Supabase Storage and returns the public URL.
// path should be explicit like "journals/{id}/cover.jpg"
func (s *Service) Upload(ctx context.Context, file []byte, path, contentType string) (string, error) {
	url := fmt.Sprintf("%s/storage/v1/object/%s/%s", config.Env.Supabase.URL, config.Env.Supabase.Bucket, path)

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(file))
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+config.Env.Supabase.ServiceRoleKey)
	req.Header.Set("Content-Type", contentType)
	// x-upsert: true to overwrite
	req.Header.Set("x-upsert", "true")

	resp, err := s.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to upload to supabase: status=%d body=%s", resp.StatusCode, string(body))
	}

	// Construct public URL
	publicURL := fmt.Sprintf("%s/storage/v1/object/public/%s/%s", config.Env.Supabase.URL, config.Env.Supabase.Bucket, path)
	return publicURL, nil
}

// Delete removes a file from Supabase Storage.
func (s *Service) Delete(ctx context.Context, path string) error {
	// API: DELETE /storage/v1/object/{bucket}/{wildcard}
	// But standard API usually is DELETE /storage/v1/object/{bucket}/{path}
	// Let's verify standard Supabase Storage API.
	// It seems DELETE method on object URL should work.
	// Path should be "journals/uuid/cover.jpg" etc.

	url := fmt.Sprintf("%s/storage/v1/object/%s/%s", config.Env.Supabase.URL, config.Env.Supabase.Bucket, path)

	req, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+config.Env.Supabase.ServiceRoleKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to delete from supabase: status=%d body=%s", resp.StatusCode, string(body))
	}

	return nil
}
