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
	// Pattern: https://<project>.supabase.co/storage/v1/object/public/<bucket>/<path>
	publicURL := fmt.Sprintf("%s/storage/v1/object/public/%s/%s", config.Env.Supabase.URL, config.Env.Supabase.Bucket, path)
	return publicURL, nil
}
