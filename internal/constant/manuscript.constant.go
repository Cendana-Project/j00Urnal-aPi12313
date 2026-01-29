package constant

type ManuscriptStatus string
type ManuscriptFileType string

const (
	// Manuscript Statuses
	ManuscriptStatusDraft            ManuscriptStatus = "DRAFT"
	ManuscriptStatusSubmitted        ManuscriptStatus = "SUBMITTED"
	ManuscriptStatusUnderChiefReview ManuscriptStatus = "UNDER_CHIEF_REVIEW"
	ManuscriptStatusAssignedToEditor ManuscriptStatus = "ASSIGNED_TO_EDITOR"
	ManuscriptStatusUnderReview      ManuscriptStatus = "UNDER_REVIEW"
	ManuscriptStatusRevisionRequired ManuscriptStatus = "REVISION_REQUIRED"
	ManuscriptStatusRevised          ManuscriptStatus = "REVISED"
	ManuscriptStatusAccepted         ManuscriptStatus = "ACCEPTED"
	ManuscriptStatusRejected         ManuscriptStatus = "REJECTED"
	ManuscriptStatusPublished        ManuscriptStatus = "PUBLISHED"

	// Manuscript File Types
	ManuscriptFileTypeMain        ManuscriptFileType = "MAIN"
	ManuscriptFileTypeFigure      ManuscriptFileType = "FIGURE"
	ManuscriptFileTypeTable       ManuscriptFileType = "TABLE"
	ManuscriptFileTypeSupplement  ManuscriptFileType = "SUPPLEMENT"
	ManuscriptFileTypeRevision    ManuscriptFileType = "REVISION"
	ManuscriptFileTypeTurnitin    ManuscriptFileType = "TURNITIN"
	ManuscriptFileTypeCoverLetter ManuscriptFileType = "COVER_LETTER"
)
