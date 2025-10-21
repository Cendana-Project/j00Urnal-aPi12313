package constant

// Permission slugs
const (
	// User & Auth
	PermUserView   = "user.view"
	PermUserCreate = "user.create"
	PermUserUpdate = "user.update"
	PermUserDelete = "user.delete"

	// Roles & Permissions
	PermRoleView   = "role.view"
	PermRoleAssign = "role.assign"
	PermPermView   = "permission.view"

	// Patient data
	PermPatientView = "patient.view"
	PermPatientEdit = "patient.edit"

	// Doctor data
	PermDoctorView = "doctor.view"
	PermDoctorEdit = "doctor.edit"

	// Medical records (EMR)
	PermEMRView = "emr.view"
	PermEMREdit = "emr.edit"

	// Appointments & Schedules
	PermAppointmentView = "appointment.view"
	PermAppointmentEdit = "appointment.edit"

	// Billing
	PermBillingView = "billing.view"
	PermBillingEdit = "billing.edit"
)

// All permissions set helper
func allPermsSlice() []string {
	return []string{
		PermUserView, PermUserCreate, PermUserUpdate, PermUserDelete,
		PermRoleView, PermRoleAssign, PermPermView,
		PermPatientView, PermPatientEdit,
		PermDoctorView, PermDoctorEdit,
		PermEMRView, PermEMREdit,
		PermAppointmentView, PermAppointmentEdit,
		PermBillingView, PermBillingEdit,
	}
}

// Matrix per role
var (
	PermissionsPatient = []string{
		PermUserView,
		PermAppointmentView,
	}

	PermissionsDoctor = []string{
		PermUserView,
		PermPatientView, PermPatientEdit,
		PermEMRView, PermEMREdit,
		PermAppointmentView, PermAppointmentEdit,
	}

	PermissionsNurse = []string{
		PermUserView,
		PermPatientView, PermPatientEdit,
		PermEMRView,
		PermAppointmentView, PermAppointmentEdit,
	}

	PermissionsReceptionist = []string{
		PermUserView,
		PermPatientView,
		PermAppointmentView, PermAppointmentEdit,
	}

	PermissionsBOD = []string{
		PermUserView, PermRoleView, PermPermView,
		PermPatientView, PermDoctorView,
		PermAppointmentView,
		PermBillingView,
	}

	PermissionsAdmin = []string{
		PermUserView, PermUserCreate, PermUserUpdate, PermUserDelete,
		PermRoleView, PermRoleAssign, PermPermView,
		PermPatientView, PermPatientEdit,
		PermDoctorView, PermDoctorEdit,
		PermEMRView, PermEMREdit,
		PermAppointmentView, PermAppointmentEdit,
		PermBillingView, PermBillingEdit,
	}

	// super_admin = semua permission
	PermissionsSuperAdmin = allPermsSlice()
)

// AllPermissions merangkum semua slug untuk seeding
func AllPermissions() []string { return allPermsSlice() }
