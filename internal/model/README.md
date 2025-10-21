# Model

This directory contains data structures and interfaces used throughout the application.

## Subdirectories

- **cachekey/**: Cache key patterns for data storage
- **contract/**: Interfaces defining component behaviors
- **entity/**: Core domain entities
- **request/**: Structures for incoming API requests
- **response/**: Structures for outgoing API responses

## Purpose

The model layer:

1. Defines data shapes
2. Builds contracts between layers 
3. Provides type-safe request/response structures
4. Defines business entities

## Example

```go
// User entity
type User struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	Username  string    `gorm:"type:varchar(50);uniqueIndex"`
	Email     string    `gorm:"type:varchar(100);uniqueIndex"`
	Password  string    `gorm:"type:varchar(255)"`
	Level     string    `gorm:"type:varchar(20)"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
```