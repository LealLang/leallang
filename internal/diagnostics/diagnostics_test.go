// Copyright (c) 2026 LealLang Contributors
// SPDX-License-Identifier: MIT

package diagnostics

import (
	"strings"
	"testing"

	"github.com/LealLang/leallang/internal/token"
)

func TestSeverityString(t *testing.T) {
	tests := []struct {
		sev  Severity
		want string
	}{
		{Error, "error"},
		{Warning, "warning"},
		{Severity(99), "unknown"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.sev.String(); got != tt.want {
				t.Fatalf("Severity.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDiagnosticsEmpty(t *testing.T) {
	d := New()
	if d.HasErrors() {
		t.Fatal("empty collector should not have errors")
	}
	if len(d.Errors()) != 0 {
		t.Fatalf("errors = %d, want 0", len(d.Errors()))
	}
	if len(d.All()) != 0 {
		t.Fatalf("all = %d, want 0", len(d.All()))
	}
}

func TestDiagnosticsReportError(t *testing.T) {
	d := New()
	pos := token.Position{File: "test.ll", Line: 1, Col: 1}
	d.ReportError("E001", "unexpected character", pos, "remove it", "x ~ y")

	if !d.HasErrors() {
		t.Fatal("should have errors")
	}
	errs := d.Errors()
	if len(errs) != 1 {
		t.Fatalf("errors = %d, want 1", len(errs))
	}
	if errs[0].Code != "E001" || errs[0].Severity != Error {
		t.Fatalf("bad error: %#v", errs[0])
	}
	if len(d.All()) != 1 {
		t.Fatalf("all = %d, want 1", len(d.All()))
	}
}

func TestDiagnosticsReportWarning(t *testing.T) {
	d := New()
	pos := token.Position{File: "test.ll", Line: 1, Col: 1}
	d.ReportWarning("W001", "unused variable", pos, "remove it", "x = 1")

	if d.HasErrors() {
		t.Fatal("warnings should not count as errors")
	}
	if len(d.Errors()) != 0 {
		t.Fatalf("errors = %d, want 0", len(d.Errors()))
	}
	all := d.All()
	if len(all) != 1 || all[0].Severity != Warning {
		t.Fatalf("bad warning: %#v", all[0])
	}
}

func TestDiagnosticsMixed(t *testing.T) {
	d := New()
	pos := token.Position{File: "test.ll", Line: 1, Col: 1}
	d.ReportWarning("W001", "unused variable", pos, "", "x = 1")
	d.ReportError("E001", "unexpected character", pos, "", "x ~ y")

	if !d.HasErrors() {
		t.Fatal("should have errors")
	}
	if len(d.Errors()) != 1 {
		t.Fatalf("errors = %d, want 1", len(d.Errors()))
	}
	if len(d.All()) != 2 {
		t.Fatalf("all = %d, want 2", len(d.All()))
	}
}

func TestFormatContainsSeverityAndCode(t *testing.T) {
	d := New()
	pos := token.Position{File: "test.ll", Line: 5, Col: 3}
	d.ReportError("E001", "unexpected character '~'", pos, "remove it", "x ~ y")

	out := d.Format()
	for _, want := range []string{"error[E001]", "unexpected character", "test.ll:5:3", "hint: remove it"} {
		if !strings.Contains(out, want) {
			t.Errorf("Format() missing %q:\n%s", want, out)
		}
	}
}

func TestFormatContainsSourceExcerpt(t *testing.T) {
	d := New()
	pos := token.Position{File: "test.ll", Line: 1, Col: 3}
	d.ReportError("E001", "unexpected character", pos, "", "x ~ y")

	out := d.Format()
	// The source excerpt contains ANSI color codes around the highlighted character,
	// so check for the parts surrounding it.
	if !strings.Contains(out, "x ") || !strings.Contains(out, " y") {
		t.Errorf("Format() missing source excerpt:\n%s", out)
	}
	if !strings.Contains(out, "unexpected character") {
		t.Errorf("Format() missing error message in excerpt:\n%s", out)
	}
}

func TestFormatNoSourceLine(t *testing.T) {
	d := New()
	pos := token.Position{File: "test.ll", Line: 1, Col: 1}
	d.ReportError("E004", "unterminated comment", pos, "add */", "")

	out := d.Format()
	if !strings.Contains(out, "error[E004]") {
		t.Errorf("Format() missing error code:\n%s", out)
	}
	if !strings.Contains(out, "hint: add */") {
		t.Errorf("Format() missing hint:\n%s", out)
	}
}
