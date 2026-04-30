package output

import (
	"bytes"
	"strings"
	"testing"

	jwch "github.com/west2-online/jwch"
)

func TestJSON(t *testing.T) {
	var buf bytes.Buffer
	if err := JSON(&buf, map[string]string{"name": "FZU"}); err != nil {
		t.Fatalf("JSON() error = %v", err)
	}
	if got := buf.String(); !strings.Contains(got, "\"name\": \"FZU\"") {
		t.Fatalf("JSON() = %q", got)
	}
}

func TestCoursesTable(t *testing.T) {
	var buf bytes.Buffer
	err := Courses(&buf, "2025-2026-1", []*jwch.Course{
		{Name: "Math", Credits: "4", Teacher: "Lin", Type: "required", RawScheduleRules: "Mon\n1-2"},
	})
	if err != nil {
		t.Fatalf("Courses() error = %v", err)
	}
	got := buf.String()
	for _, want := range []string{"Term", "2025-2026-1", "Math", "Mon 1-2"} {
		if !strings.Contains(got, want) {
			t.Fatalf("Courses() = %q, want substring %q", got, want)
		}
	}
}
