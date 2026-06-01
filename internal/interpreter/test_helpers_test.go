// Copyright (c) 2026 LealLang Contributors
// SPDX-License-Identifier: MIT

package interpreter

import (
	"bytes"
	"sync"
	"testing"

	"github.com/LealLang/leallang/internal/checker"
	"github.com/LealLang/leallang/internal/diagnostics"
	"github.com/LealLang/leallang/internal/lexer"
	"github.com/LealLang/leallang/internal/parser"
)

// safeWriter wraps a bytes.Buffer with a mutex for concurrent use.
type safeWriter struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (w *safeWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.buf.Write(p)
}

func (w *safeWriter) String() string {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.buf.String()
}

func runSource(t *testing.T, src string) string {
	t.Helper()
	out, diag, err := runSourceWithDiagnostics(t, src)
	if err != nil {
		t.Fatalf("runtime error: %v\n%s", err, diag.Format())
	}
	if diag.HasErrors() {
		t.Fatalf("unexpected diagnostics:\n%s", diag.Format())
	}
	return out
}

func runSourceWithDiagnostics(t *testing.T, src string) (string, *diagnostics.Diagnostics, error) {
	t.Helper()
	diag := diagnostics.New()
	diag.SetSource(src)
	tokens := lexer.New("test.ll", src, diag).Tokenize()
	program := parser.New(tokens, diag).Parse()
	if diag.HasErrors() {
		t.Fatalf("parse errors:\n%s", diag.Format())
	}
	checker.Check(program, diag)
	if diag.HasErrors() {
		t.Fatalf("checker errors:\n%s", diag.Format())
	}
	var stdout, stderr safeWriter
	interp := New(diag)
	interp.SetOutput(&stdout, &stderr)
	interp.SetArgs([]string{"one", "two"})
	err := interp.Run(program)
	return stdout.String(), diag, err
}

func requireRuntimeCode(t *testing.T, src, code string) {
	t.Helper()
	_, diag, err := runSourceWithDiagnostics(t, src)
	if err == nil && !diag.HasErrors() {
		t.Fatalf("expected runtime diagnostic %s", code)
	}
	for _, got := range diag.Errors() {
		if got.Code == code {
			return
		}
	}
	t.Fatalf("expected diagnostic %s, got:\n%s", code, diag.Format())
}

// runSourceLenient runs source and returns diagnostics without fatalling on checker errors.
func runSourceLenient(t *testing.T, src string) (string, *diagnostics.Diagnostics, error) {
	t.Helper()
	diag := diagnostics.New()
	diag.SetSource(src)
	tokens := lexer.New("test.ll", src, diag).Tokenize()
	program := parser.New(tokens, diag).Parse()
	if diag.HasErrors() {
		return "", diag, nil
	}
	checker.Check(program, diag)
	if diag.HasErrors() {
		return "", diag, nil
	}
	var stdout, stderr safeWriter
	interp := New(diag)
	interp.SetOutput(&stdout, &stderr)
	interp.SetArgs([]string{"one", "two"})
	err := interp.Run(program)
	return stdout.String(), diag, err
}
