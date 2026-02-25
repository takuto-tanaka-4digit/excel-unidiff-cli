package report

import (
	"slices"
	"testing"
)

func TestDeriveStatus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		findings int
		errors   int
		wantS    Status
		wantC    int
	}{
		{name: "ok", findings: 0, errors: 0, wantS: StatusOK, wantC: 0},
		{name: "findings", findings: 2, errors: 0, wantS: StatusFindings, wantC: 1},
		{name: "error precedence", findings: 3, errors: 1, wantS: StatusError, wantC: 2},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s, c := DeriveStatus(tt.findings, tt.errors)
			if s != tt.wantS || c != tt.wantC {
				t.Fatalf("DeriveStatus(%d,%d) = (%s,%d), want (%s,%d)", tt.findings, tt.errors, s, c, tt.wantS, tt.wantC)
			}
		})
	}
}

func TestValidateInvariantsValid(t *testing.T) {
	t.Parallel()

	p := "repo"
	r := Report{
		SchemaVersion: SchemaVersion,
		Status:        StatusFindings,
		ExitCode:      1,
		Summary: Summary{
			TargetRepos: 1,
			Findings:    1,
			Errors:      0,
		},
		Results: []Result{{RepoPath: p, Kind: "unicode", Code: "UG011"}},
	}

	errList := ValidateInvariants(r, 1)
	if len(errList) != 0 {
		t.Fatalf("ValidateInvariants returned errors: %v", errList)
	}
}

func TestValidateInvariantsInvalid(t *testing.T) {
	t.Parallel()

	r := Report{
		SchemaVersion: "0.0.0",
		Status:        StatusOK,
		ExitCode:      1,
		Summary: Summary{
			TargetRepos: 0,
			Findings:    2,
			Errors:      2,
		},
		Results: []Result{{RepoPath: "repo1", Kind: "drift", Code: "UG004"}},
		Errors:  []ErrorItem{{Code: "UG002", RepoPath: ptr("repo2")}},
	}

	errList := ValidateInvariants(r, 2)
	if len(errList) == 0 {
		t.Fatal("ValidateInvariants returned no errors, want > 0")
	}
	mustContain(t, errList, "schemaVersion")
	mustContain(t, errList, "status-exit")
	mustContain(t, errList, "status invariant")
	mustContain(t, errList, "summary.findings")
	mustContain(t, errList, "summary.errors")
	mustContain(t, errList, "summary.targetRepos")
	mustContain(t, errList, "process exit code")
}

func TestSortResultsAndErrors(t *testing.T) {
	t.Parallel()

	p1 := "expected-b"
	p2 := "actual-b"
	pathA := "a\\x\\e\u0301.txt"
	pathB := "a/x/é.txt"

	r := Report{
		Results: []Result{
			{RepoPath: "repo/z", Kind: "unicode", Path: &pathB, Code: "UG005", Expected: &p1, Actual: &p2},
			{RepoPath: "repo/a", Kind: "drift", Path: nil, Code: "UG004"},
			{RepoPath: "repo/a", Kind: "drift", Path: &pathA, Code: "UG004"},
		},
		Errors: []ErrorItem{
			{Code: "UG010", RepoPath: ptr("repo/b"), Path: &pathB},
			{Code: "UG002", RepoPath: ptr("repo/a"), Path: &pathA},
		},
	}

	SortResultsAndErrors(&r)

	if r.Results[0].RepoPath != "repo/a" || r.Results[0].Path == nil {
		t.Fatalf("results[0] unexpected: %#v", r.Results[0])
	}
	if r.Results[1].RepoPath != "repo/a" || r.Results[1].Path != nil {
		t.Fatalf("results[1] unexpected: %#v", r.Results[1])
	}
	if r.Results[2].RepoPath != "repo/z" {
		t.Fatalf("results[2] unexpected: %#v", r.Results[2])
	}

	if !slices.Equal([]string{r.Errors[0].Code, r.Errors[1].Code}, []string{"UG002", "UG010"}) {
		t.Fatalf("errors order = [%s %s], want [UG002 UG010]", r.Errors[0].Code, r.Errors[1].Code)
	}
}

func TestValidateInvariantsUnknownStatus(t *testing.T) {
	t.Parallel()

	r := Report{
		SchemaVersion: SchemaVersion,
		Status:        Status("unknown"),
		ExitCode:      2,
		Summary:       Summary{TargetRepos: 0, Findings: 0, Errors: 0},
	}

	errList := ValidateInvariants(r, -1)
	mustContain(t, errList, "status-exit")
	mustContain(t, errList, "status invariant")
}

func TestMapStatusToExitCode(t *testing.T) {
	t.Parallel()

	tests := []struct {
		status Status
		wantC  int
		wantOK bool
	}{
		{status: StatusOK, wantC: 0, wantOK: true},
		{status: StatusFindings, wantC: 1, wantOK: true},
		{status: StatusError, wantC: 2, wantOK: true},
		{status: Status("x"), wantC: 0, wantOK: false},
	}
	for _, tt := range tests {
		c, ok := mapStatusToExitCode(tt.status)
		if c != tt.wantC || ok != tt.wantOK {
			t.Fatalf("mapStatusToExitCode(%q) = (%d,%t), want (%d,%t)", tt.status, c, ok, tt.wantC, tt.wantOK)
		}
	}
}

func TestIsStatusInvariantValid(t *testing.T) {
	t.Parallel()

	if !isStatusInvariantValid(Report{Status: StatusOK, Summary: Summary{Findings: 0, Errors: 0}}) {
		t.Fatal("ok invariant should be valid")
	}
	if !isStatusInvariantValid(Report{Status: StatusFindings, Summary: Summary{Findings: 1, Errors: 0}}) {
		t.Fatal("findings invariant should be valid")
	}
	if !isStatusInvariantValid(Report{Status: StatusError, ExitCode: 2, Summary: Summary{Errors: 1}}) {
		t.Fatal("error invariant should be valid")
	}
	if isStatusInvariantValid(Report{Status: Status("bad")}) {
		t.Fatal("unknown invariant should be invalid")
	}
}

func TestCompareOptionalString(t *testing.T) {
	t.Parallel()

	a := "a"
	b := "b"
	pathA := "c\\x\\e\u0301"
	pathB := "c/x/é"

	if compareOptionalString(nil, nil, false) != 0 {
		t.Fatal("nil,nil should be equal")
	}
	if compareOptionalString(nil, &a, false) != 1 {
		t.Fatal("nil,left should be nulls-last")
	}
	if compareOptionalString(&a, nil, false) != -1 {
		t.Fatal("right nil should be nulls-last")
	}
	if compareOptionalString(&a, &b, false) >= 0 {
		t.Fatal("a should be less than b")
	}
	if compareOptionalString(&pathA, &pathB, true) != 0 {
		t.Fatal("canonical path compare should treat equivalent NFC paths as equal")
	}
}

func TestCanonicalForSort(t *testing.T) {
	t.Parallel()

	got := canonicalForSort("c:\\tmp\\e\u0301")
	if got != "C:/tmp/é" {
		t.Fatalf("canonicalForSort = %q, want C:/tmp/é", got)
	}
}

func TestSortResultsAndErrorsTieBreakers(t *testing.T) {
	t.Parallel()

	e1 := "a"
	e2 := "b"
	a1 := "x"
	a2 := "y"
	p := "repo/a"
	path := "repo/a.txt"

	r := Report{
		Results: []Result{
			{RepoPath: p, Kind: "drift", Path: &path, Code: "UG004", Expected: &e2, Actual: &a1},
			{RepoPath: p, Kind: "drift", Path: &path, Code: "UG004", Expected: &e1, Actual: &a2},
		},
		Errors: []ErrorItem{
			{Code: "UG001", RepoPath: ptr(p), Path: nil},
			{Code: "UG001", RepoPath: ptr(p), Path: &path},
		},
	}

	SortResultsAndErrors(&r)
	if *r.Results[0].Expected != "a" {
		t.Fatalf("results tie-break by expected failed: %#v", r.Results)
	}
	if r.Errors[0].Path == nil {
		t.Fatalf("errors nulls-last failed: %#v", r.Errors)
	}
}

func TestSortResultsAndErrorsByKindCodeActual(t *testing.T) {
	t.Parallel()

	path := "repo/a.txt"
	expected := "same"
	actualA := "a"
	actualB := "b"

	r := Report{
		Results: []Result{
			{RepoPath: "repo/a", Kind: "unicode", Path: &path, Code: "UG011", Expected: &expected, Actual: &actualB},
			{RepoPath: "repo/a", Kind: "drift", Path: &path, Code: "UG011", Expected: &expected, Actual: &actualA},
			{RepoPath: "repo/a", Kind: "drift", Path: &path, Code: "UG011", Expected: &expected, Actual: &actualB},
			{RepoPath: "repo/a", Kind: "drift", Path: &path, Code: "UG004", Expected: &expected, Actual: &actualA},
		},
	}

	SortResultsAndErrors(&r)
	if r.Results[0].Kind != "drift" || r.Results[0].Code != "UG004" {
		t.Fatalf("first result order mismatch: %#v", r.Results)
	}
	if r.Results[1].Kind != "drift" || r.Results[1].Code != "UG011" {
		t.Fatalf("second result order mismatch: %#v", r.Results)
	}
	if r.Results[2].Kind != "drift" || *r.Results[2].Actual != "b" {
		t.Fatalf("third result order mismatch: %#v", r.Results)
	}
	if r.Results[3].Kind != "unicode" {
		t.Fatalf("fourth result order mismatch: %#v", r.Results)
	}
}

func TestSortErrorsByRepoAndPath(t *testing.T) {
	t.Parallel()

	pathA := "repo/a.txt"
	pathB := "repo/b.txt"
	r := Report{
		Errors: []ErrorItem{
			{Code: "UG001", RepoPath: ptr("repo/b"), Path: &pathB},
			{Code: "UG001", RepoPath: ptr("repo/a"), Path: &pathA},
		},
	}

	SortResultsAndErrors(&r)
	if r.Errors[0].RepoPath == nil || *r.Errors[0].RepoPath != "repo/a" {
		t.Fatalf("errors repo sort mismatch: %#v", r.Errors)
	}
}

func TestSortResultsAndErrorsCanonicalRepoPath(t *testing.T) {
	t.Parallel()

	repoA := "c:\\repo\\a"
	repoB := "C:/repo/a"
	path := "x.txt"

	r := Report{
		Results: []Result{
			{RepoPath: repoA, Kind: "drift", Path: &path, Code: "UG004"},
			{RepoPath: repoB, Kind: "drift", Path: &path, Code: "UG004"},
		},
		Errors: []ErrorItem{
			{Code: "UG001", RepoPath: ptr(repoA), Path: &path},
			{Code: "UG001", RepoPath: ptr(repoB), Path: &path},
		},
	}

	SortResultsAndErrors(&r)
	if r.Results[0].RepoPath != repoA || r.Results[1].RepoPath != repoB {
		t.Fatalf("results canonical sort mismatch: %#v", r.Results)
	}
	if r.Errors[0].RepoPath == nil || *r.Errors[0].RepoPath != repoA {
		t.Fatalf("errors canonical sort mismatch: %#v", r.Errors)
	}
}

func mustContain(t *testing.T, list []string, token string) {
	t.Helper()
	for _, item := range list {
		if contains(item, token) {
			return
		}
	}
	t.Fatalf("expected token %q in %v", token, list)
}

func contains(s, token string) bool {
	return len(token) == 0 || (len(s) >= len(token) && indexOf(s, token) >= 0)
}

func indexOf(s, token string) int {
	for i := 0; i+len(token) <= len(s); i++ {
		if s[i:i+len(token)] == token {
			return i
		}
	}
	return -1
}

func ptr(s string) *string {
	return &s
}
