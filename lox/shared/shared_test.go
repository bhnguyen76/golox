package shared

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func captureStderr(f func()) string {
	// Save original stderr
	old := os.Stderr

	r, w, err := os.Pipe()
	if err != nil {
		panic(err)
	}
	os.Stderr = w

	// Run the function while stderr is redirected
	f()

	// Restore stderr and close writer
	_ = w.Close()
	os.Stderr = old

	// Read captured output
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	_ = r.Close()
	return buf.String()
}

func TestErrorAtSetsHadErrorAndPrints(t *testing.T) {
	ResetErrors()

	out := captureStderr(func() {
		ErrorAt(42, "Something went wrong")
	})

	if !HadError {
		t.Fatalf("expected HadError to be true after ErrorAt")
	}

	if !strings.Contains(out, "[Line 42] Error: Something went wrong") {
		t.Fatalf("unexpected stderr output: %q", out)
	}
}

func TestReportWithWhereFormatsMessage(t *testing.T) {
	ResetErrors()

	out := captureStderr(func() {
		Report(7, " at 'foo'", "Bad token")
	})

	if !HadError {
		t.Fatalf("expected HadError to be true after Report")
	}

	expected := "[Line 7] Error at 'foo': Bad token"
	if !strings.Contains(out, expected) {
		t.Fatalf("expected stderr to contain %q, got %q", expected, out)
	}
}
