package pagination

import "gorm.io/gorm"

type Pagination struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
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
			p.PageSize = 10
		}

		offset := (p.Page - 1) * p.PageSize
		return db.Offset(offset).Limit(p.PageSize)
	}
}
