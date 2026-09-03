package config

import (
	"strings"
	"testing"
)

func validTestConfig() *EnvConfig {
	return &EnvConfig{
		Env: "development",
		Server: Server{
			Port: "8080",
		},
		Database: Database{
			DSN: "postgres://localhost/journal",
		},
		Redis: Redis{
			IsCacheDisable: true,
		},
		Token: Token{
			PasswordSalt:       "password-salt-16",
			AccessTokenSecret:  "access-secret-16",
			RefreshTokenSecret: "refresh-secret-16",
		},
		Supabase: Supabase{
			URL:            "https://example.supabase.co",
			ServiceRoleKey: "service-role-key",
		},
	}
}

func TestEnvConfigValidateSupabaseRequired(t *testing.T) {
	tests := []struct {
		name        string
		configure   func(*EnvConfig)
		wantMessage string
	}{
		{
			name: "missing URL",
			configure: func(cfg *EnvConfig) {
				cfg.Supabase.URL = ""
			},
			wantMessage: "SUPABASE_URL is required",
		},
		{
			name: "blank URL",
			configure: func(cfg *EnvConfig) {
				cfg.Supabase.URL = "   "
			},
			wantMessage: "SUPABASE_URL is required",
		},
		{
			name: "missing service role key",
			configure: func(cfg *EnvConfig) {
				cfg.Supabase.ServiceRoleKey = ""
			},
			wantMessage: "SUPABASE_SERVICE_ROLE_KEY is required",
		},
		{
			name: "blank service role key",
			configure: func(cfg *EnvConfig) {
				cfg.Supabase.ServiceRoleKey = "\t"
			},
			wantMessage: "SUPABASE_SERVICE_ROLE_KEY is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := validTestConfig()
			tt.configure(cfg)

			err := cfg.Validate()
			if err == nil {
				t.Fatalf("Validate() returned nil; want error containing %q", tt.wantMessage)
			}
			if !strings.Contains(err.Error(), tt.wantMessage) {
				t.Fatalf("Validate() error = %q; want it to contain %q", err, tt.wantMessage)
			}
		})
	}
}

func TestEnvConfigValidateAggregatesSupabaseErrors(t *testing.T) {
	cfg := validTestConfig()
	cfg.Database.DSN = ""
	cfg.Supabase.URL = ""
	cfg.Supabase.ServiceRoleKey = ""

	err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate() returned nil; want validation error")
	}

	for _, want := range []string{
		"DATABASE_DSN is required",
		"SUPABASE_URL is required",
		"SUPABASE_SERVICE_ROLE_KEY is required",
	} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("Validate() error = %q; want it to contain %q", err, want)
		}
	}
}

func TestEnvConfigValidateAcceptsCompleteConfig(t *testing.T) {
	if err := validTestConfig().Validate(); err != nil {
		t.Fatalf("Validate() returned unexpected error: %v", err)
	}
}
