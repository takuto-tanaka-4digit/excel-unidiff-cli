package scan

import (
	"slices"
	"testing"
)

func TestAnalyzePaths(t *testing.T) {
	t.Parallel()

	paths := []string{
		"plain/ascii.txt",
		"unicode/é.xlsx",       // NFC-only
		"unicode/e\u0301.xlsx", // NFD-only, collides with previous after NFC
		"unicode/a\u030A.txt",  // combining mark
	}

	result := AnalyzePaths(paths)

	if result.Metrics.NFCOnly != 1 {
		t.Fatalf("NFCOnly = %d, want 1", result.Metrics.NFCOnly)
	}
	if result.Metrics.NFDOnly != 2 {
		t.Fatalf("NFDOnly = %d, want 2", result.Metrics.NFDOnly)
	}
	if result.Metrics.NFCCollisions != 1 {
		t.Fatalf("NFCCollisions = %d, want 1", result.Metrics.NFCCollisions)
	}
	if result.Metrics.CombiningMarkPaths != 2 {
		t.Fatalf("CombiningMarkPaths = %d, want 2", result.Metrics.CombiningMarkPaths)
	}

	if len(result.Collisions) != 1 {
		t.Fatalf("len(Collisions) = %d, want 1", len(result.Collisions))
	}
	collision := result.Collisions[0]
	if collision.NormalizedPath != "unicode/é.xlsx" {
		t.Fatalf("NormalizedPath = %q, want unicode/é.xlsx", collision.NormalizedPath)
	}
	wantColliding := []string{"unicode/e\u0301.xlsx", "unicode/é.xlsx"}
	if !slices.Equal(collision.CollidingPaths, wantColliding) {
		t.Fatalf("CollidingPaths = %v, want %v", collision.CollidingPaths, wantColliding)
	}
}

func TestAnalyzePathsEmpty(t *testing.T) {
	t.Parallel()

	result := AnalyzePaths(nil)
	if result.Metrics != (Metrics{}) {
		t.Fatalf("Metrics = %#v, want zero", result.Metrics)
	}
	if len(result.Collisions) != 0 {
		t.Fatalf("len(Collisions) = %d, want 0", len(result.Collisions))
	}
}

func TestAnalyzePathsCollisionSort(t *testing.T) {
	t.Parallel()

	paths := []string{
		"b/é.txt",
		"b/e\u0301.txt",
		"a/Å.txt",
		"a/A\u030A.txt",
	}

	result := AnalyzePaths(paths)
	if result.Metrics.NFCCollisions != 2 {
		t.Fatalf("NFCCollisions = %d, want 2", result.Metrics.NFCCollisions)
	}
	if len(result.Collisions) != 2 {
		t.Fatalf("len(Collisions) = %d, want 2", len(result.Collisions))
	}
	if result.Collisions[0].NormalizedPath != "a/Å.txt" {
		t.Fatalf("Collisions[0] = %q, want a/Å.txt", result.Collisions[0].NormalizedPath)
	}
	if result.Collisions[1].NormalizedPath != "b/é.txt" {
		t.Fatalf("Collisions[1] = %q, want b/é.txt", result.Collisions[1].NormalizedPath)
	}
}

func TestAnalyzePathsCollisionIncludesAllCandidates(t *testing.T) {
	t.Parallel()

	paths := []string{
		"multi/Å.txt",
		"multi/A\u030A.txt",
		"multi/Å.txt",
	}

	result := AnalyzePaths(paths)
	if result.Metrics.NFCCollisions != 1 {
		t.Fatalf("NFCCollisions = %d, want 1", result.Metrics.NFCCollisions)
	}
	if len(result.Collisions) != 1 {
		t.Fatalf("len(Collisions) = %d, want 1", len(result.Collisions))
	}
	want := []string{
		"multi/A\u030A.txt",
		"multi/Å.txt",
		"multi/Å.txt",
	}
	if !slices.Equal(result.Collisions[0].CollidingPaths, want) {
		t.Fatalf("CollidingPaths = %v, want %v", result.Collisions[0].CollidingPaths, want)
	}
}
