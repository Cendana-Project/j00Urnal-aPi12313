package constant

type ManuscriptStatus string
type ManuscriptFileType string

const (
	// My Queue
	ManuscriptStatusSubmitted        ManuscriptStatus = "SUBMITTED"
	ManuscriptStatusAssignedToEditor ManuscriptStatus = "ASSIGNED_TO_EDITOR"
	ManuscriptStatusUnderReview      ManuscriptStatus = "UNDER_REVIEW"
	ManuscriptStatusRevisionRequired ManuscriptStatus = "REVISION_REQUIRED"
	ManuscriptStatusRevised          ManuscriptStatus = "REVISED"

	// Production
	ManuscriptStatusInProduction     ManuscriptStatus = "IN_PRODUCTION"

	// Archived
	ManuscriptStatusAccepted  ManuscriptStatus = "ACCEPTED"
	ManuscriptStatusRejected  ManuscriptStatus = "REJECTED"
	ManuscriptStatusPublished ManuscriptStatus = "PUBLISHED"
)

var (
	StatusQueue = []ManuscriptStatus{
		ManuscriptStatusSubmitted,
		ManuscriptStatusAssignedToEditor,
		ManuscriptStatusUnderReview,
		ManuscriptStatusRevisionRequired,
		ManuscriptStatusRevised,
		ManuscriptStatusAccepted,
		ManuscriptStatusInProduction,
	}

	StatusArchives = []ManuscriptStatus{
		ManuscriptStatusAccepted,
		ManuscriptStatusRejected,
		ManuscriptStatusPublished,
	}
)

func (s ManuscriptStatus) IsQueue() bool {
	switch s {
	case ManuscriptStatusSubmitted,
		ManuscriptStatusAssignedToEditor,
		ManuscriptStatusUnderReview,
		ManuscriptStatusRevisionRequired,
		ManuscriptStatusRevised:
		return true
	}
	return false
}

func (s ManuscriptStatus) IsArchived() bool {
	switch s {
	case ManuscriptStatusAccepted,
		ManuscriptStatusRejected,
		ManuscriptStatusPublished:
		return true
	}
	return false
}

const (
	// Manuscript File Types
	ManuscriptFileTypeMain        ManuscriptFileType = "MAIN"
	ManuscriptFileTypeFigure      ManuscriptFileType = "FIGURE"
	ManuscriptFileTypeTable       ManuscriptFileType = "TABLE"
	ManuscriptFileTypeSupplement  ManuscriptFileType = "SUPPLEMENT"
	ManuscriptFileTypeRevision    ManuscriptFileType = "REVISION"
	ManuscriptFileTypeTurnitin    ManuscriptFileType = "TURNITIN"
	ManuscriptFileTypeCoverLetter ManuscriptFileType = "COVER_LETTER"
	ManuscriptFileTypeCopyedited  ManuscriptFileType = "COPYEDITED"
	ManuscriptFileTypeCopyeditedRevision ManuscriptFileType = "COPYEDITED_REVISION"
)

// ====== Review Round ======

type ReviewRoundStatus string

const (
	ReviewRoundStatusPending           ReviewRoundStatus = "PENDING"
	ReviewRoundStatusInReview          ReviewRoundStatus = "IN_REVIEW"
	ReviewRoundStatusRevisionRequested ReviewRoundStatus = "REVISION_REQUESTED"
	ReviewRoundStatusCompleted         ReviewRoundStatus = "COMPLETED"
)

// ====== Review Assignment ======

type ReviewAssignmentStatus string

const (
	ReviewAssignmentStatusInvited   ReviewAssignmentStatus = "INVITED"
	ReviewAssignmentStatusAccepted  ReviewAssignmentStatus = "ACCEPTED"
	ReviewAssignmentStatusDeclined  ReviewAssignmentStatus = "DECLINED"
	ReviewAssignmentStatusCompleted ReviewAssignmentStatus = "COMPLETED"
	ReviewAssignmentStatusExpired   ReviewAssignmentStatus = "EXPIRED"
	ReviewAssignmentStatusWithdrawn ReviewAssignmentStatus = "WITHDRAWN"
)

// ====== Review Recommendation ======

type ReviewRecommendation string

const (
	ReviewRecommendationAccept        ReviewRecommendation = "ACCEPT"
	ReviewRecommendationReject        ReviewRecommendation = "REJECT"
	ReviewRecommendationMajorRevision ReviewRecommendation = "MAJOR_REVISION"
	ReviewRecommendationMinorRevision ReviewRecommendation = "MINOR_REVISION"
)

// ====== Review File Type ======

type ReviewFileType string

const (
	ReviewFileTypeComment    ReviewFileType = "REVIEW_COMMENT"
	ReviewFileTypeRevision   ReviewFileType = "REVISION"
	ReviewFileTypeAttachment ReviewFileType = "REVIEW_ATTACHMENT"
	ReviewFileTypeReviewerPDF ReviewFileType = "REVIEWER_PDF"
)

// Review assignment report row status (review_assignment_reports.status)
type ReviewReportStatus string

const (
	ReviewReportStatusDraft     ReviewReportStatus = "DRAFT"
	ReviewReportStatusSubmitted ReviewReportStatus = "SUBMITTED"
)

// Review extension request workflow
type ReviewExtensionRequestStatus string

const (
	ReviewExtensionStatusPending   ReviewExtensionRequestStatus = "PENDING"
	ReviewExtensionStatusApproved  ReviewExtensionRequestStatus = "APPROVED"
	ReviewExtensionStatusRejected  ReviewExtensionRequestStatus = "REJECTED"
)
