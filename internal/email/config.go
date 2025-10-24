package email

import (
	"os"
	"time"
)

type Config struct {
	Enabled     bool
	Provider    string
	Host        string
	Port        int
	Username    string
	Password    string
	FromEmail   string
	FromName    string
	Timeout     time.Duration
	UseSTARTTLS bool
}

// NewConfigFromMap membaca dari map (hasil unmarshal config.yml) dan ENV override.
func NewConfigFromMap(m map[string]any) *Config {
	getStr := func(key string) string {
		if v, ok := m[key]; ok && v != nil {
			if s, ok2 := v.(string); ok2 {
				return s
			}
		}
		return ""
	}
	getInt := func(key string) int {
		if v, ok := m[key]; ok && v != nil {
			switch t := v.(type) {
			case int:
				return t
			case int64:
				return int(t)
			case float64:
				return int(t)
			}
		}
		return 0
	}
	getBool := func(key string) bool {
		if v, ok := m[key]; ok && v != nil {
			if b, ok2 := v.(bool); ok2 {
				return b
			}
		}
		return false
	}

	c := &Config{
		Enabled:     getBool("enabled"),
		Provider:    getStr("provider"),
		Host:        getStr("host"),
		Port:        getInt("port"),
		Username:    getStr("username"),
		Password:    getStr("password"),
		FromEmail:   getStr("from_email"),
		FromName:    getStr("from_name"),
		Timeout:     time.Duration(getInt("timeout_seconds")) * time.Second,
		UseSTARTTLS: getBool("use_starttls"),
	}

	// ENV override (prioritas)
	if v := os.Getenv("EMAIL_ENABLED"); v == "true" || v == "1" {
		c.Enabled = true
	}
	if v := os.Getenv("EMAIL_PROVIDER"); v != "" {
		c.Provider = v
	}
	if v := os.Getenv("EMAIL_HOST"); v != "" {
		c.Host = v
	}
	if v := os.Getenv("EMAIL_PORT"); v != "" {
		// biarkan sederhana; bisa parse int jika mau
	}
	if v := os.Getenv("EMAIL_USERNAME"); v != "" {
		c.Username = v
	}
	if v := os.Getenv("EMAIL_PASSWORD"); v != "" {
		c.Password = v
	}
	if v := os.Getenv("EMAIL_FROM"); v != "" {
		c.FromEmail = v
	}
	if v := os.Getenv("EMAIL_FROM_NAME"); v != "" {
		c.FromName = v
	}
	return c
}
