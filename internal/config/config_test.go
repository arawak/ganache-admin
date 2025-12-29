package config

import "testing"

func TestNormalizeBasePath(t *testing.T) {
	tests := []struct {
		name string
		in   string
		out  string
	}{
		{"empty", "", ""},
		{"slash", "/", ""},
		{"plain segment", "media", "/media"},
		{"nested segments", "media/admin", "/media/admin"},
		{"prefixed", "/media/admin", "/media/admin"},
		{"trailing slash", "/media/admin/", "/media/admin"},
		{"double slashes", "//media///admin//", "/media/admin"},
		{"spaces", "  /media/admin  ", "/media/admin"},
		{"root alias", "//", ""},
		{"dot segments", "/media/admin/../library", "/media/library"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := normalizeBasePath(tt.in); got != tt.out {
				t.Fatalf("normalizeBasePath(%q) = %q, want %q", tt.in, got, tt.out)
			}
		})
	}
}
