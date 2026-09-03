package user

import "testing"

func TestPrimaryGlobalRoleSlug(t *testing.T) {
	tests := []struct {
		name  string
		roles []string
		want  string
	}{
		{name: "empty", roles: nil, want: ""},
		{name: "normalizes", roles: []string{" reviewer "}, want: "REVIEWER"},
		{name: "reviewer beats author independent of DB order", roles: []string{"AUTHOR", "REVIEWER"}, want: "REVIEWER"},
		{name: "chief editor beats reviewer", roles: []string{"REVIEWER", "CHIEF_EDITOR", "AUTHOR"}, want: "CHIEF_EDITOR"},
		{name: "super admin wins", roles: []string{"EDITOR", "SUPER_ADMIN", "CHIEF_EDITOR"}, want: "SUPER_ADMIN"},
		{name: "known role beats custom role", roles: []string{"Z_CUSTOM", "AUTHOR"}, want: "AUTHOR"},
		{name: "custom roles are deterministic", roles: []string{"Z_CUSTOM", "A_CUSTOM"}, want: "A_CUSTOM"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := PrimaryGlobalRoleSlug(tt.roles); got != tt.want {
				t.Fatalf("PrimaryGlobalRoleSlug(%v) = %q, want %q", tt.roles, got, tt.want)
			}
		})
	}
}
