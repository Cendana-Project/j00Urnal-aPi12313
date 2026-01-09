package constant

type PublicationStatus string
type EntityType string
type FileType string

const (
	// Publication Statuses
	PublicationStatusDraft     PublicationStatus = "DRAFT"
	PublicationStatusActive    PublicationStatus = "ACTIVE"
	PublicationStatusArchived  PublicationStatus = "ARCHIVED"
	PublicationStatusPublished PublicationStatus = "PUBLISHED"

	// Entity Types
	EntityTypeJournal    EntityType = "JOURNAL"
	EntityTypeVolume     EntityType = "VOLUME"
	EntityTypeIssue      EntityType = "ISSUE"
	EntityTypeManuscript EntityType = "MANUSCRIPT"

	// File Types
	FileTypeCover        FileType = "COVER"
	FileTypeFullIssuePDF FileType = "FULL_ISSUE_PDF"
	FileTypeMain         FileType = "MAIN"
	FileTypeFigure       FileType = "FIGURE"
	FileTypeTable        FileType = "TABLE"
	FileTypeSupplement   FileType = "SUPPLEMENT"
)
