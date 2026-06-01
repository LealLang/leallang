// Copyright (c) 2026 LealLang Contributors
// SPDX-License-Identifier: MIT

package interpreter

import (
	"strings"
	"testing"
)

func TestConstDeclaration(t *testing.T) {
	out := runSource(t, `package app.main

const PI: float = 3.14

func main():
    console.print_ln(PI)
`)
	if !strings.Contains(out, "3.14") {
		t.Fatalf("expected 3.14, got %q", out)
	}
}

func TestConstDeclarationWithExplicitType(t *testing.T) {
	out := runSource(t, `package app.main

const MAX: int = 100

func main():
    console.print_ln(MAX)
`)
	if !strings.Contains(out, "100") {
		t.Fatalf("expected 100, got %q", out)
	}
}

func TestConstDeclarationFloat(t *testing.T) {
	out := runSource(t, `package app.main

const PI: float = 3.14

func main():
    console.print_ln(PI)
`)
	if !strings.Contains(out, "3.14") {
		t.Fatalf("expected 3.14, got %q", out)
	}
}

func TestConstDeclarationString(t *testing.T) {
	out := runSource(t, `package app.main

const NAME: string = "leal"

func main():
    console.print_ln(NAME)
`)
	if !strings.Contains(out, "leal") {
		t.Fatalf("expected leal, got %q", out)
	}
}

func TestConstDeclarationBool(t *testing.T) {
	out := runSource(t, `package app.main

const FLAG: bool = true

func main():
    console.print_ln(FLAG)
`)
	if !strings.Contains(out, "true") {
		t.Fatalf("expected true, got %q", out)
	}
}

func TestConstWithInferredType(t *testing.T) {
	out := runSource(t, `package app.main

const MAX = 100

func main():
    console.print_ln(MAX)
`)
	if !strings.Contains(out, "100") {
		t.Fatalf("expected 100, got %q", out)
	}
}

func TestConstDeclInsideFunc(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    const X: int = 42
    console.print_ln(X)
`)
	if !strings.Contains(out, "42") {
		t.Fatalf("expected 42, got %q", out)
	}
}

func TestConstGroupAccess(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    c = colors.blue
    console.print_ln(c)
    d = dock.left
    console.print_ln(d)
    o = orientation.horizontal
    console.print_ln(o)
    t = toast_type.error
    console.print_ln(t)
`)
	if !strings.Contains(out, "blue") || !strings.Contains(out, "left") || !strings.Contains(out, "horizontal") || !strings.Contains(out, "error") {
		t.Fatalf("expected const group values, got %q", out)
	}
}

func TestStubNamespaceCall(t *testing.T) {
	out, diag, _ := runSourceWithDiagnostics(t, `package app.main

func main():
    x = true
    window.open(x)
`)
	// Stubs write to stderr, so just verify no errors.
	_ = out
	if diag.HasErrors() {
		t.Fatalf("unexpected errors: %s", diag.Format())
	}
}

func TestStubMsgNamespace(t *testing.T) {
	out, diag, _ := runSourceWithDiagnostics(t, `package app.main

func main():
    msg.alert("hello")
`)
	_ = out
	if diag.HasErrors() {
		t.Fatalf("unexpected errors: %s", diag.Format())
	}
}

func TestStubModalNamespace(t *testing.T) {
	out, diag, _ := runSourceWithDiagnostics(t, `package app.main

func main():
    x = true
    modal.open(x)
`)
	_ = out
	if diag.HasErrors() {
		t.Fatalf("unexpected errors: %s", diag.Format())
	}
}

func TestStubToastNamespace(t *testing.T) {
	out, diag, _ := runSourceWithDiagnostics(t, `package app.main

func main():
    t = toast_type.success
    toast.show("hello", 1000, t)
`)
	_ = out
	if diag.HasErrors() {
		t.Fatalf("unexpected errors: %s", diag.Format())
	}
}

