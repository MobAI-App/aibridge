package patterns

import "testing"

func TestGetPattern(t *testing.T) {
	tests := []struct {
		tool  string
		regex string
	}{
		{"claude", "esc to interrupt"},
		{"codex", "esc to interrupt"},
		{"gemini", "esc to cancel"},
	}

	for _, tt := range tests {
		t.Run(tt.tool, func(t *testing.T) {
			p := GetPattern(tt.tool)
			if p == nil || p.Regex != tt.regex {
				t.Errorf("GetPattern(%q) = %v, want Regex=%q", tt.tool, p, tt.regex)
			}
		})
	}
}

func TestGetPatternUnknown(t *testing.T) {
	p := GetPattern("unknown")
	if p != nil {
		t.Error("GetPattern for unknown tool should return nil")
	}
}

func TestDefaultPattern(t *testing.T) {
	p := DefaultPattern()
	if p == nil || p.Regex == "" {
		t.Errorf("DefaultPattern() = %v, want non-nil with non-empty Regex", p)
	}
}
