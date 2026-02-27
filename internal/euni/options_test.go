package euni

import "testing"

func TestParseOptionsApplyRepairUnicodeDeletes(t *testing.T) {
	t.Parallel()

	opts, err := ParseOptions([]string{
		"apply",
		"--repo", ".",
		"--policy", "./.euni.yml",
		"--repair-unicode-deletes",
	})
	if err != nil {
		t.Fatalf("ParseOptions returned error: %v", err)
	}
	if !opts.RepairUnicodeDeletes {
		t.Fatalf("RepairUnicodeDeletes = %v, want true", opts.RepairUnicodeDeletes)
	}
}

func TestParseOptionsRejectRepairUnicodeDeletesOutsideApply(t *testing.T) {
	t.Parallel()

	_, err := ParseOptions([]string{
		"check",
		"--repair-unicode-deletes",
	})
	if err == nil {
		t.Fatalf("ParseOptions returned nil error, want UG009")
	}

	ug, ok := err.(UGError)
	if !ok {
		t.Fatalf("error type = %T, want UGError", err)
	}
	if ug.Code != "UG009" {
		t.Fatalf("UG code = %s, want UG009", ug.Code)
	}
}
