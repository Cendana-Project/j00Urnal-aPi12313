# Config

This directory contains code related to application configuration management. This layer handles various settings required for the application to run properly.

## Main Files

- **config.go**: Defines the configuration structure and provides functionality to load settings from various sources (environment variables, configuration files, etc.)

## Purpose

The config layer is responsible for:

1. Defining the application configuration structure
2. Loading configuration values from various sources (YAML files, environment variables)
3. Validating configuration values
4. Providing centralized access to configuration values throughout the application

This ensures the application has access to all necessary configurations to run properly, such as database connection strings, API keys, feature flags, and other environment-specific settings.

## Example Code

```go
// Application configuration structure
type Config struct {
    Server struct {
        Port             string `yaml:"port" validate:"required"`
        Env              string `yaml:"env" validate:"required,oneof=development staging production"`
        ShutdownTimeout  int    `yaml:"shutdownTimeout" validate:"required"`
    } `yaml:"server"`
    
    Database struct {
        Host     string `yaml:"host" validate:"required"`
        Port     string `yaml:"port" validate:"required"`
        Username string `yaml:"username" validate:"required"`
        Password string `yaml:"password" validate:"required"`
        Name     string `yaml:"name" validate:"required"`
        SSLMode  string `yaml:"sslMode" validate:"required"`
        Schema   string `yaml:"schema" validate:"required"`
    } `yaml:"database"`
    
    Redis struct {
        Host     string `yaml:"host" validate:"required"`
        Port     string `yaml:"port" validate:"required"`
        Password string `yaml:"password"`
        DB       int    `yaml:"db" validate:"required"`
    } `yaml:"redis"`
    
    JWT struct {
        AccessTokenExpiration  int    `yaml:"accessTokenExpiration" validate:"required"`
        RefreshTokenExpiration int    `yaml:"refreshTokenExpiration" validate:"required"`
        Secret                 string `yaml:"secret" validate:"required"`
    } `yaml:"jwt"`
}

// Load configuration from file
func LoadConfig() (*Config, error) {
    filename := "config.yml"
    if os.Getenv("CONFIG_FILE") != "" {
        filename = os.Getenv("CONFIG_FILE")
    }
    
    f, err := os.Open(filename)
    if err != nil {
        return nil, err
    }
    defer f.Close()
    
    var cfg Config
    decoder := yaml.NewDecoder(f)
    if err := decoder.Decode(&cfg); err != nil {
        return nil, err
    }
    
    // Validate configuration
    validate := validator.New()
    if err := validate.Struct(&cfg); err != nil {
        return nil, err
    }
    
    return &cfg, nil
}

// Get configuration values
func GetConfig() *Config {
    // Singleton pattern for configuration
    if config == nil {
        cfg, err := LoadConfig()
        if err != nil {
            log.Fatalf("Failed to load configuration: %v", err)
        }
        config = cfg
    }
    return config
}
````