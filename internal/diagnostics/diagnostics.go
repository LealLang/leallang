// Copyright (c) 2026 LealLang Contributors
// SPDX-License-Identifier: MIT

// Package diagnostics provides error and warning reporting for the LealLang compiler.
package diagnostics

import (
	"fmt"
	"strings"

	"github.com/LealLang/leallang/internal/token"
)

// ANSI color codes.
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[1;31m"
	colorGray   = "\033[90m"
	colorOrange = "\033[38;5;208m"
)

// Severity represents the severity level of a diagnostic.
type Severity int

const (
	// Error indicates a compilation error.
	Error Severity = iota
	// Warning indicates a non-fatal warning.
	Warning
)

// String returns the human-readable name of the severity.
func (s Severity) String() string {
	switch s {
	case Error:
		return "error"
	case Warning:
		return "warning"
	default:
		return "unknown"
	}
}

// Diagnostic represents a single diagnostic message.
type Diagnostic struct {
	Severity   Severity
	Code       string // e.g. "E001", "W001"
	Message    string
	Pos        token.Position
	Hint       string // optional fix suggestion
	SourceLine string // source text of the line for excerpt display
}

// Diagnostics collects multiple diagnostic messages during compilation.
type Diagnostics struct {
	items []Diagnostic
}

// New creates a new Diagnostics collector.
func New() *Diagnostics {
	return &Diagnostics{}
}

// ReportError adds an error diagnostic.
func (d *Diagnostics) ReportError(code, msg string, pos token.Position, hint, sourceLine string) {
	d.items = append(d.items, Diagnostic{
		Severity:   Error,
		Code:       code,
		Message:    msg,
		Pos:        pos,
		Hint:       hint,
		SourceLine: sourceLine,
	})
}

// ReportWarning adds a warning diagnostic.
func (d *Diagnostics) ReportWarning(code, msg string, pos token.Position, hint, sourceLine string) {
	d.items = append(d.items, Diagnostic{
		Severity:   Warning,
		Code:       code,
		Message:    msg,
		Pos:        pos,
		Hint:       hint,
		SourceLine: sourceLine,
	})
}

// HasErrors returns true if any error-level diagnostics have been collected.
func (d *Diagnostics) HasErrors() bool {
	for _, item := range d.items {
		if item.Severity == Error {
			return true
		}
	}
	return false
}

// Errors returns only the error-level diagnostics.
func (d *Diagnostics) Errors() []Diagnostic {
	var errs []Diagnostic
	for _, item := range d.items {
		if item.Severity == Error {
			errs = append(errs, item)
		}
	}
	return errs
}

// All returns all collected diagnostics.
func (d *Diagnostics) All() []Diagnostic {
	return d.items
}

// Format renders all diagnostics as human-readable colored output.
func (d *Diagnostics) Format() string {
	var b strings.Builder
	for i, diag := range d.items {
		if i > 0 {
			b.WriteString("\n")
		}
		d.formatOne(&b, diag)
	}
	return b.String()
}

// formatOne renders a single diagnostic with ANSI colors.
func (d *Diagnostics) formatOne(b *strings.Builder, diag Diagnostic) {
	// error[E001]: unexpected character '`'  (red)
	b.WriteString(fmt.Sprintf("%s%s[%s]%s: %s\n",
		colorRed, diag.Severity, diag.Code, colorReset, diag.Message))

	//   --> src/main.ll:12:8  (gray arrow, reset path)
	b.WriteString(fmt.Sprintf("  %s-->%s %s:%d:%d\n",
		colorGray, colorReset, diag.Pos.File, diag.Pos.Line, diag.Pos.Col))

	// Use right-aligned line number for consistent padding
	lineNumStr := fmt.Sprintf("%d", diag.Pos.Line)
	sepPrefix := strings.Repeat(" ", len(lineNumStr)) + colorGray + " |" + colorReset + " "

	//   |  (gray)
	b.WriteString(sepPrefix + "\n")

	if diag.SourceLine != "" {
		// 12 |     label: `name`  (gray line num, source with red highlight)
		col := diag.Pos.Col // 1-based
		b.WriteString(colorGray + lineNumStr + " |" + colorReset + " ")

		// Write source with the "guilty" character highlighted in red
		if col >= 1 && col <= len(diag.SourceLine) {
			before := diag.SourceLine[:col-1]
			char := diag.SourceLine[col-1 : col]
			after := ""
			if col < len(diag.SourceLine) {
				after = diag.SourceLine[col:]
			}
			b.WriteString(before + colorRed + string(char) + colorReset + after)
		} else {
			b.WriteString(diag.SourceLine)
		}
		b.WriteString("\n")

		//   |            ^ unexpected character  (gray pipe, red caret + message)
		b.WriteString(sepPrefix + strings.Repeat(" ", col-1) +
			colorRed + "^ " + diag.Message + colorReset + "\n")
	}

	// = hint: ...  (orange)
	if diag.Hint != "" {
		b.WriteString(colorOrange + "   = hint: " + diag.Hint + colorReset + "\n")
	}
}
