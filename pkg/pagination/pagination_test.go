package pagination

import "testing"

func TestNormalizeQuery(t *testing.T) {
	tests := []struct {
		name         string
		page         int
		pageSize     int
		wantPage     int
		wantPageSize int
	}{
		{"defaults", 1, 10, 1, 10},
		{"page_zero", 0, 10, 1, 10},
		{"page_negative", -3, 10, 1, 10},
		{"pageSize_zero", 2, 0, 2, DefaultPageSize},
		{"pageSize_negative", 1, -1, 1, DefaultPageSize},
		{"pageSize_capped", 1, 5000, 1, MaxPageSize},
		{"pageSize_at_max", 3, MaxPageSize, 3, MaxPageSize},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotP, gotPS := NormalizeQuery(tt.page, tt.pageSize)
			if gotP != tt.wantPage || gotPS != tt.wantPageSize {
				t.Fatalf("NormalizeQuery(%d,%d) = (%d,%d), want (%d,%d)",
					tt.page, tt.pageSize, gotP, gotPS, tt.wantPage, tt.wantPageSize)
			}
		})
	}
}
