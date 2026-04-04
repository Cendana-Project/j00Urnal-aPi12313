package pagination

import "gorm.io/gorm"

// DefaultPageSize and MaxPageSize are used for list endpoints that read page/page_size from queries.
const (
	DefaultPageSize = 10
	MaxPageSize     = 100
)

type Pagination struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
}

// NormalizeQuery returns safe page (>= 1) and page size in [1, MaxPageSize] for HTTP query parsing.
func NormalizeQuery(page, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = DefaultPageSize
	}
	if pageSize > MaxPageSize {
		pageSize = MaxPageSize
	}
	return page, pageSize
}

func New(page, pageSize int) *Pagination {
	return &Pagination{
		Page:     page,
		PageSize: pageSize,
	}
}

func (p *Pagination) Paginate() func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if p.Page <= 0 {
			p.Page = 1
		}

		if p.PageSize <= 0 {
			p.PageSize = DefaultPageSize
		}

		offset := (p.Page - 1) * p.PageSize
		return db.Offset(offset).Limit(p.PageSize)
	}
}
