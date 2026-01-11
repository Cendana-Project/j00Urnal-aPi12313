package constant

// ==========================
// PERMISSION SLUGS — MedikaOne
// ==========================

// User & Role & Permission (admin area)
const (
	PermissionUserView   = "user.view"
	PermissionUserCreate = "user.create"
	PermissionUserUpdate = "user.update"
	PermissionUserDelete = "user.delete"

	PermissionRoleView   = "role.view"
	PermissionRoleAssign = "role.assign" // assign role ke user

	PermissionPermissionView = "permission.view"
)

// Patient / Doctor profile
const (
	PermissionPatientView = "patient.view"
	PermissionPatientEdit = "patient.edit"

	PermissionDoctorView = "doctor.view"
	PermissionDoctorEdit = "doctor.edit"
)

// EMR & Billing
const (
	PermissionEMRView = "emr.view"
	PermissionEMREdit = "emr.edit"

	PermissionBillingView = "billing.view"
	PermissionBillingEdit = "billing.edit"
)

// Appointment
const (
	PermissionAppointmentView = "appointment.view"
	PermissionAppointmentEdit = "appointment.edit"
)

// Manuscript
const (
	PermissionManuscriptManage = "manuscript.manage"
)

// Journal
const (
	PermissionJournalManage = "journal.manage"
)

// ==========================
// DEFAULT ROLE → PERMISSIONS
// ==========================
//
// Catatan:
// - super_admin mendapat SEMUA permission
// - admin (rumah sakit) = manajemen RS
// - nurse & receptionist hak operasional
// - bod pengawasan & laporan (read mostly)
// - patient & doctor minimal yang diperlukan
var DefaultRolePermissions = map[string][]string{
	// super admin: semua
	RoleSuperAdmin: {
		PermissionUserView, PermissionUserCreate, PermissionUserUpdate, PermissionUserDelete,
		PermissionRoleView, PermissionRoleAssign, PermissionPermissionView,
		PermissionPatientView, PermissionPatientEdit,
		PermissionDoctorView, PermissionDoctorEdit,
		PermissionEMRView, PermissionEMREdit,
		PermissionBillingView, PermissionBillingEdit,
		PermissionAppointmentView, PermissionAppointmentEdit,
		PermissionManuscriptManage,
		PermissionJournalManage,
	},

	RoleEditor: {
		PermissionManuscriptManage,
		PermissionJournalManage,
	},

	RoleChiefEditor: {
		PermissionManuscriptManage,
		PermissionJournalManage,
	},

	RoleAuthor: {
		PermissionAppointmentView, // Maybe needed? Or just minimal
		// Author should have permissions relevant to them if any
	},

	RoleReviewer: {
		// Reviewer permissions
	},
}
