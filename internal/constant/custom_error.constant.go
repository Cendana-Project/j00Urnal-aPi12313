package constant

import (
	"net/http"

	"github.com/api-monolith-template/internal/model/response"
)

// NOTE:
// - File ini menyatukan semua konstanta error yang digunakan oleh util, transport, dan service (auth).
// - Beberapa error punya pasangan "ValidationFailed" vs "ValidationError" agar kompatibel dengan util lama.
// - Detail pesan memakai GetMessageDetail dari message.constant.go. Jika suatu Msg* belum ada, hapus Detail-nya atau tambahkan di message.constant.go.

var (
	// ====== Generic / Common ======
	ErrInternalServerError = response.CustomError{
		Code:       "INTERNAL_SERVER_ERROR",
		StatusCode: http.StatusInternalServerError,
		Message:    "internal server error",
		Detail:     GetMessageDetail(MsgInternalServerError),
	}
	ErrValidationError = response.CustomError{
		Code:       "VALIDATION_ERROR",
		StatusCode: http.StatusBadRequest,
		Message:    "validation error",
		Detail:     GetMessageDetail(MsgValidationError),
	}
	// Alias gaya penamaan yang lain (dipakai di beberapa tempat)
	ErrValidationFailed = response.CustomError{
		Code:       "VALIDATION_FAILED",
		StatusCode: http.StatusBadRequest,
		Message:    "validation failed",
		Detail:     GetMessageDetail(MsgValidationFailed),
	}

	ErrEndpointNotFound = response.CustomError{
		Code:       "ENDPOINT_NOT_FOUND",
		StatusCode: http.StatusNotFound,
		Message:    "endpoint not found",
		Detail:     GetMessageDetail(MsgEndpointNotFound),
	}

	ErrUnauthorized = response.CustomError{
		Code:       "UNAUTHORIZED",
		StatusCode: http.StatusForbidden,
		Message:    "unauthorized",
		Detail:     GetMessageDetail(MsgUnauthorized),
	}
	ErrForbidden = response.CustomError{
		Code:       "FORBIDDEN",
		StatusCode: http.StatusForbidden,
		Message:    "forbidden",
		Detail:     GetMessageDetail(MsgForbidden),
	}
	ErrUserNotAuthenticated = response.CustomError{
		Code:       "USER_NOT_AUTHENTICATED",
		StatusCode: http.StatusUnauthorized,
		Message:    "user not authenticated",
		Detail:     GetMessageDetail(MsgUserNotAuthenticated),
	}

	ErrTooManyRequests = response.CustomError{
		Code:       "TOO_MANY_REQUESTS",
		StatusCode: http.StatusTooManyRequests,
		Message:    "too many requests, please try again later",
	}

	// ====== Input / Format ======
	ErrInvalidEmail = response.CustomError{
		Code:       "INVALID_EMAIL",
		StatusCode: http.StatusBadRequest,
		Message:    "invalid email format",
	}
	ErrInvalidPassword = response.CustomError{
		Code:       "INVALID_PASSWORD",
		StatusCode: http.StatusBadRequest,
		Message:    "Password must meet all requirements: at least 8 characters, contain uppercase letter, lowercase letter, number, and special character. Must not contain sequential numbers (e.g. 1234) or repeated characters (e.g. aaaa).",
	}
	ErrPasswordSimilarToUserInfo = response.CustomError{
		Code:       "PASSWORD_SIMILAR_TO_USER_INFO",
		StatusCode: http.StatusBadRequest,
		Message:    "password cannot be similar to your username or email",
	}
	ErrInvalidDateFormat = response.CustomError{
		Code:       "INVALID_DATE_FORMAT",
		StatusCode: http.StatusBadRequest,
		Message:    "invalid date format, use RFC3339 format (example: 1997-12-22T00:00:00Z)",
	}
	ErrInvalidUUIDFormat = response.CustomError{
		Code:       "INVALID_UUID_FORMAT",
		StatusCode: http.StatusBadRequest,
		Message:    "invalid UUID format",
		Detail:     GetMessageDetail(MsgInvalidUUIDFormat),
	}

	// ====== Resource / Data ======
	ErrRecordNotFound = response.CustomError{
		Code:       "RECORD_NOT_FOUND",
		StatusCode: http.StatusNotFound,
		Message:    "record not found",
		Detail:     GetMessageDetail(MsgRecordNotFound),
	}
	ErrUserNotFound = response.CustomError{
		Code:       "USER_NOT_FOUND",
		StatusCode: http.StatusNotFound,
		Message:    "user not found",
		Detail:     GetMessageDetail(MsgUserNotFound),
	}
	ErrDuplicateUsernameOrEmail = response.CustomError{
		Code:       "DUPLICATE_USERNAME_OR_EMAIL",
		StatusCode: http.StatusConflict,
		Message:    "Username or email already exists",
	}
	ErrDuplicateNIK = response.CustomError{ // <=== added
		Code:       "DUPLICATE_NIK",
		StatusCode: http.StatusConflict,
		Message:    "NIK already exists",
		Detail:     GetMessageDetail(MsgConflict),
	}
	ErrConflict = response.CustomError{
		Code:       "CONFLICT",
		StatusCode: http.StatusConflict,
		Message:    "conflict",
		Detail:     GetMessageDetail(MsgConflict), // Ensure generic conflict has detail if needed, or leave as is
	}
	ErrVolumeNotActive = response.CustomError{
		Code:       "VOLUME_NOT_ACTIVE",
		StatusCode: http.StatusConflict,
		Message:    "parent volume is not active",
		Detail:     GetMessageDetail(MsgVolumeNotActive),
	}
	ErrJournalNotActive = response.CustomError{
		Code:       "JOURNAL_NOT_ACTIVE",
		StatusCode: http.StatusConflict,
		Message:    "parent journal is not active",
		Detail:     GetMessageDetail(MsgJournalNotActive),
	}

	// ====== Auth / Token / OTP/PIN ======
	ErrInvalidToken = response.CustomError{
		Code:       "INVALID_TOKEN",
		StatusCode: http.StatusUnauthorized,
		Message:    "invalid token",
	}
	ErrTokenExpired = response.CustomError{
		Code:       "TOKEN_EXPIRED",
		StatusCode: http.StatusUnauthorized,
		Message:    "token expired",
	}
	ErrInvalidOTP = response.CustomError{
		Code:       "INVALID_OTP",
		StatusCode: http.StatusBadRequest,
		Message:    "invalid or expired OTP",
	}
	ErrEmailNotVerified = response.CustomError{
		Code:       "EMAIL_NOT_VERIFIED",
		StatusCode: http.StatusForbidden,
		Message:    "please verify your email before logging in",
	}

	// ====== Role / Permission ======
	ErrInvalidRoleID = response.CustomError{
		Code:       "INVALID_ROLE_ID",
		StatusCode: http.StatusBadRequest,
		Message:    "invalid role_id",
		Detail:     GetMessageDetail(MsgInvalidRoleID),
	}
	ErrRoleAlreadyExist = response.CustomError{
		Code:       "ROLE_ALREADY_EXIST",
		StatusCode: http.StatusConflict,
		Message:    "role already exists",
		Detail:     GetMessageDetail(MsgRoleAlreadyExist),
	}
	ErrRoleNotFound = response.CustomError{
		Code:       "ROLE_NOT_FOUND",
		StatusCode: http.StatusNotFound,
		Message:    "role not found",
		Detail:     GetMessageDetail(MsgRoleNotFound),
	}
	ErrRoleNotAssigned = response.CustomError{
		Code:       "ROLE_NOT_ASSIGNED",
		StatusCode: http.StatusBadRequest,
		Message:    "role not assigned to user",
		Detail:     GetMessageDetail(MsgRoleNotAssigned),
	}
	ErrRoleAlreadyAssigned = response.CustomError{
		Code:       "ROLE_ALREADY_ASSIGNED",
		StatusCode: http.StatusBadRequest,
		Message:    "role already assigned to user",
		Detail:     GetMessageDetail(MsgRoleAlreadyAssigned),
	}
	ErrRoleInUse = response.CustomError{
		Code:       "ROLE_IN_USE",
		StatusCode: http.StatusConflict,
		Message:    "role currently in use",
		Detail:     GetMessageDetail(MsgRoleInUse),
	}
	ErrOnlySuperAdmin = response.CustomError{
		Code:       "ONLY_SUPER_ADMIN",
		StatusCode: http.StatusForbidden,
		Message:    "only super admin",
		Detail:     GetMessageDetail(MsgOnlySuperAdmin),
	}

	// ====== MedikaOne – Auth (registrasi PIN) ======
	ErrEmailAlreadyActive = response.CustomError{
		Code:       "EMAIL_ALREADY_ACTIVE",
		StatusCode: http.StatusConflict,
		Message:    "email already registered and active",
		Detail:     GetMessageDetail(MsgConflict),
	}
	ErrEmailSendFailed = response.CustomError{
		Code:       "EMAIL_SEND_FAILED",
		StatusCode: http.StatusBadGateway,
		Message:    "failed to send verification email, please try again later",
		Detail:     GetMessageDetail(MsgEmailSendFailed),
	}
	ErrAccountRoleNotFound = response.CustomError{
		Code:       "ACCOUNT_ROLE_NOT_FOUND",
		StatusCode: http.StatusNotFound,
		Message:    "account role not found",
		Detail:     GetMessageDetail(MsgRoleNotFound),
	}

	// ====== Review / Invitation ======
	ErrInvitationExpired = response.CustomError{
		Code:       "INVITATION_EXPIRED",
		StatusCode: http.StatusGone,
		Message:    "Reviewer invitation has expired (7-day limit exceeded)",
		Detail:     GetMessageDetail(MsgInvitationExpired),
	}
	ErrInvitationAlreadyAccepted = response.CustomError{
		Code:       "INVITATION_ALREADY_ACCEPTED",
		StatusCode: http.StatusConflict,
		Message:    "This invitation has already been accepted",
		Detail:     GetMessageDetail(MsgInvitationAlreadyAccepted),
	}
	ErrInvitationNotFound = response.CustomError{
		Code:       "INVITATION_NOT_FOUND",
		StatusCode: http.StatusNotFound,
		Message:    "Invitation not found or invalid token",
		Detail:     GetMessageDetail(MsgNotFound),
	}
	ErrReviewNotCompleted = response.CustomError{
		Code:       "REVIEW_NOT_COMPLETED",
		StatusCode: http.StatusBadRequest,
		Message:    "Not all reviews have been completed for this round",
		Detail:     GetMessageDetail(MsgReviewNotCompleted),
	}
	ErrInvalidManuscriptStatus = response.CustomError{
		Code:       "INVALID_MANUSCRIPT_STATUS",
		StatusCode: http.StatusBadRequest,
		Message:    "Manuscript is not in the correct status for this action",
		Detail:     GetMessageDetail(MsgInvalidManuscriptStatus),
	}
	ErrEditorNotAssigned = response.CustomError{
		Code:       "EDITOR_NOT_ASSIGNED",
		StatusCode: http.StatusForbidden,
		Message:    "You are not assigned as editor for this manuscript",
		Detail:     GetMessageDetail(MsgEditorNotAssigned),
	}

	// ====== Lain-lain yang masih direferensi di util/transport lama ======
	ErrUnauthorizedUpdate = response.CustomError{
		Code:       "UNAUTHORIZED_UPDATE",
		StatusCode: http.StatusForbidden,
		Message:    "You can only update your own user data",
	}
	ErrNewPasswordSame = response.CustomError{
		Code:       "NEW_PASSWORD_SAME",
		StatusCode: http.StatusBadRequest,
		Message:    "new password cannot be the same as old password",
	}
	ErrPasswordNotMatch = response.CustomError{
		Code:       "PASSWORD_NOT_MATCH",
		StatusCode: http.StatusUnauthorized,
		Message:    "password not match",
		Detail:     GetMessageDetail(MsgPasswordNotMatch),
	}
	ErrHospitalNotFound = response.CustomError{
		Code:       "HOSPITAL_NOT_FOUND",
		StatusCode: http.StatusNotFound,
		Message:    "hospital not found",
		Detail:     GetMessageDetail(MsgNotFound),
	}
	ErrUserNotLinkedToHospital = response.CustomError{
		Code:       "USER_NOT_LINKED_TO_HOSPITAL",
		StatusCode: http.StatusForbidden,
		Message:    "user is not linked to this hospital",
		Detail:     GetMessageDetail(MsgForbidden),
	}
	ErrProfileAlreadySet = response.CustomError{
		Code:       "PROFILE_ALREADY_SET",
		StatusCode: http.StatusConflict,
		Message:    "profile has already been set",
		Detail:     GetMessageDetail(MsgConflict),
	}
	ErrRegistrationError = response.CustomError{
		Code:       "REGISTRATION_ERROR",
		StatusCode: http.StatusInternalServerError,
		Message:    "internal server error",
		Detail:     GetMessageDetail(MsgInternalServerError),
	}
)
