package entity

type ManuscriptAuthor struct {
	ID              string  `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	ManuscriptID    string  `json:"manuscript_id" gorm:"type:uuid;not null"`
	UserID          *string `json:"user_id,omitempty" gorm:"type:uuid;nullable"`
	AuthorName      string  `json:"author_name" gorm:"type:varchar;not null"`
	AuthorEmail     string  `json:"author_email" gorm:"type:varchar;not null"`
	Affiliation     string  `json:"affiliation" gorm:"type:text"`
	IsCorresponding bool    `json:"is_corresponding" gorm:"type:boolean;default:false"`
	IsPrimaryAuthor bool    `json:"is_primary_author" gorm:"column:is_primary_author;type:boolean;default:false"`
	OrderPosition   int     `json:"order_position" gorm:"type:int;not null"`

	// Relationships
	User *User `json:"user,omitempty" gorm:"foreignKey:UserID;references:ID"`
}
