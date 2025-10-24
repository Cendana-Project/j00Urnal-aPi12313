# Repository

This directory contains implementations for data access operations. This layer handles all interactions with data storage systems like databases and caches.

## Subdirectories

- **cache/**: Implementations for interacting with cache systems.
- **user/**: Implementations for user-related data operations.

## Purpose

The repository layer:

1. Provides abstraction over data storage mechanisms
2. Implements CRUD operations for domain entities
3. Handles data persistence and retrieval logic
4. Manages transactions and data consistency
5. Translates between domain entities and storage formats

## Example Code

### Base Repository
```go
// User repository base structure
type Repository struct {
	db        *gorm.DB
	cacheRepo contract.CacheRepository
}

// Create a new repository
func NewRepository() *Repository {
	return new(Repository)
}

// Provide database connection
func (r *Repository) WithGormDB(db *gorm.DB) *Repository {
	r.db = db
	return r
}

// Provide cache repository
func (r *Repository) WithCacheRepository(repo contract.CacheRepository) *Repository {
	r.cacheRepo = repo
	return r
}
```

### Method Implementation
```go
// FindByID gets user by ID
func (r *Repository) FindByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	// Try getting from cache first
	cacheKey := cachekey.GetUserKey(id)
	user := new(entity.User)
	
	err := r.cacheRepo.Get(ctx, cacheKey, user)
	if err == nil {
		// Data found in cache
		return user, nil
	}
	
	// If not in cache, get from database
	result := r.db.WithContext(ctx).Where("id = ?", id).First(user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("user with id %s not found", id.String())
		}
		return nil, result.Error
	}

	// Save to cache for future use
	r.cacheRepo.Set(ctx, cacheKey, user, 30*time.Minute)

	return user, nil
}
```