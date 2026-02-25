package policy

import (
	"slices"
	"testing"
)

func TestNormalizeRelativePath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "current dir", in: ".", want: ""},
		{name: "slash normalize", in: `a\\b\\c.txt`, want: "a/b/c.txt"},
		{name: "dot cleanup", in: "./docs/../README.md", want: "README.md"},
		{name: "unicode nfc", in: "decomposed/e\u0301.xlsx", want: "decomposed/é.xlsx"},
		{name: "already clean", in: "src/main.go", want: "src/main.go"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := NormalizeRelativePath(tt.in)
			if got != tt.want {
				t.Fatalf("NormalizeRelativePath(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestDetectCaseOnlyCollisions(t *testing.T) {
	t.Parallel()

	collisions := DetectCaseOnlyCollisions([]string{
		"docs/Readme.md",
		"docs/readme.md",
		"src/main.go",
	})
	if len(collisions) != 1 {
		t.Fatalf("len(collisions) = %d, want 1", len(collisions))
	}
	if collisions[0].Normalized != "docs/readme.md" {
		t.Fatalf("Normalized = %q, want docs/readme.md", collisions[0].Normalized)
	}
	wantCandidates := []string{"docs/Readme.md", "docs/readme.md"}
	if !slices.Equal(collisions[0].Candidates, wantCandidates) {
		t.Fatalf("Candidates = %v, want %v", collisions[0].Candidates, wantCandidates)
	}

	none := DetectCaseOnlyCollisions([]string{"a.txt", "b.txt"})
	if len(none) != 0 {
		t.Fatalf("len(none) = %d, want 0", len(none))
	}
}
