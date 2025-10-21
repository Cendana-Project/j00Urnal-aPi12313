# Infrastructure

This directory contains code that interacts with external systems and frameworks, providing an abstraction layer over third-party libraries.

## Main Files

- **cache.go**: Provides cache implementation and abstraction.
- **common.go**: Contains common infrastructure code used across implementations.
- **gin.go**: Sets up and configures the Gin HTTP framework with appropriate middleware and settings.
- **gorm.go**: Configures database connections using GORM ORM and provides database access utilities.

## Purpose

The infrastructure layer:

1. Sets up connections to external systems (database, cache, message queue, etc.)
2. Provides abstractions over third-party libraries to isolate the application from external dependencies
3. Manages the lifecycle of external resources (connections, pools, etc.)
4. Configures frameworks with application-specific settings

## Example Code

### Database Configuration
```go
// SetupGormDB initializes the database connection
func SetupGormDB(cfg *config.Config) (*gorm.DB, error) {
    dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s search_path=%s",
        cfg.Database.Host,
        cfg.Database.Username,
        cfg.Database.Password,
        cfg.Database.Name,
        cfg.Database.Port,
        cfg.Database.SSLMode,
        cfg.Database.Schema,
    )

    // Configure based on environment
    gormConfig := &gorm.Config{
        Logger: logger.Default.LogMode(logger.Info),
    }
    
    if cfg.Server.Env == "production" {
        gormConfig.Logger = logger.Default.LogMode(logger.Error)
    }

    db, err := gorm.Open(postgres.Open(dsn), gormConfig)
    if err != nil {
        return nil, err
    }

    // Set connection pool configuration
    sqlDB, err := db.DB()
    if err != nil {
        return nil, err
    }

    sqlDB.SetMaxIdleConns(10)
    sqlDB.SetMaxOpenConns(100)
    sqlDB.SetConnMaxLifetime(time.Hour)

    DB = db
    return db, nil
}
```

### HTTP Server Configuration
```go
// SetupGinEngine initializes the Gin HTTP server
func SetupGinEngine(cfg *config.Config) *gin.Engine {
    if cfg.Server.Env == "production" {
        gin.SetMode(gin.ReleaseMode)
    }

    r := gin.Default()

    // Configure CORS
    r.Use(cors.New(cors.Config{
        AllowOrigins:     []string{"*"},
        AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
        AllowHeaders:     []string{"Origin", "Authorization", "Content-Type"},
        ExposeHeaders:    []string{"Content-Length"},
        AllowCredentials: true,
        MaxAge:           12 * time.Hour,
    }))

    return r
}
```