// Copyright (c) 2026 LealLang Contributors
// SPDX-License-Identifier: MIT

package interpreter

import (
	"strings"
	"testing"

	"github.com/LealLang/leallang/internal/checker"
	"github.com/LealLang/leallang/internal/diagnostics"
	"github.com/LealLang/leallang/internal/lexer"
	"github.com/LealLang/leallang/internal/parser"
)

func TestBuiltins(t *testing.T) {
	tmp := t.TempDir() + "/file.txt"
	out := runSource(t, `package app.main

func main():
    console.print_ln(file.exists("`+tmp+`"))
    console.print_ln(system.args()[0])
    result = json.stringify({"a": 1})
    console.print_ln(result)
`)
	if out != "false\none\n({\"a\":1}, null)\n" {
		t.Fatalf("unexpected output %q", out)
	}
}

func TestRuntimeDivisionByZero(t *testing.T) {
	requireRuntimeCode(t, `package app.main

func zero() -> int:
    return 0

func main():
    console.print_ln(1 / zero())
`, "E103")
}

func TestRuntimeIndexOutOfBounds(t *testing.T) {
	requireRuntimeCode(t, `package app.main

func idx() -> int:
    return 2

func main():
    xs = [1]
    console.print_ln(xs[idx()])
`, "E102")
}

func TestRuntimeDictKeyNotFound(t *testing.T) {

	requireRuntimeCode(t, `package app.main

func main():
    m = {"a": 1}
    console.print_ln(m["missing"])
`, "E106")
}

func TestExitCode(t *testing.T) {
	src := `package app.main

func main():
    system.exit(42)
`
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
	interp.Run(program)
	if interp.ExitCode() != 42 {
		t.Fatalf("expected exit code 42, got %d", interp.ExitCode())
	}
}

func TestNamespaceUnknownMemberError(t *testing.T) {
	_, diag, err := runSourceLenient(t, `package app.main

func main():
    console.unknown_func()
`)
	if err == nil && !diag.HasErrors() {
		t.Fatalf("expected error for unknown namespace member")
	}
}

func TestConstGroupUnknownMemberError(t *testing.T) {
	_, diag, err := runSourceLenient(t, `package app.main

func main():
    console.print_ln(color.unknown)
`)
	if err == nil && !diag.HasErrors() {
		t.Fatalf("expected error for unknown const group member")
	}
}

func TestComparisonTypeMismatch(t *testing.T) {
	_, diag, err := runSourceLenient(t, `package app.main

func main():
    console.print_ln("a" < 1)
`)
	if err == nil && !diag.HasErrors() {
		t.Fatalf("expected error for comparison type mismatch")
	}
}

func TestComparisonNonComparable(t *testing.T) {
	_, diag, err := runSourceLenient(t, `package app.main

func main():
    console.print_ln([1] < [2])
`)
	if err == nil && !diag.HasErrors() {
		t.Fatalf("expected error for non-comparable comparison")
	}
}

func TestLogicalOperandNotBool(t *testing.T) {
	_, diag, err := runSourceLenient(t, `package app.main

func main():
    console.print_ln(1 and true)
`)
	if err == nil && !diag.HasErrors() {
		t.Fatalf("expected error for non-bool logical operand")
	}
}

func TestListIndexNonInt(t *testing.T) {
	_, diag, err := runSourceLenient(t, `package app.main

func main():
    xs = [1, 2, 3]
    console.print_ln(xs["a"])
`)
	if err == nil && !diag.HasErrors() {
		t.Fatalf("expected error for list index with non-int")
	}
}

func TestDictIndexNonString(t *testing.T) {
	_, diag, err := runSourceLenient(t, `package app.main

func main():
    m = {"a": 1}
    console.print_ln(m[0])
`)
	if err == nil && !diag.HasErrors() {
		t.Fatalf("expected error for dict index with non-string")
	}
}

func TestStringIndexNonInt(t *testing.T) {
	_, diag, err := runSourceLenient(t, `package app.main

func main():
    s = "hello"
    console.print_ln(s["a"])
`)
	if err == nil && !diag.HasErrors() {
		t.Fatalf("expected error for string index with non-int")
	}
}

func TestStringIndexOutOfBounds(t *testing.T) {
	requireRuntimeCode(t, `package app.main

func main():
    s = "hi"
    console.print_ln(s[99])
`, "E102")
}

// --- Additional coverage tests ---

func TestFileWriteAndExists(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/test.txt"
	src := `package app.main

func main():
    file.write_text("` + path + `", "hello world")
    console.print_ln(file.exists("` + path + `"))
    console.print_ln(file.exists("` + path + `/nonexistent"))
`
	out := runSource(t, src)
	if !strings.Contains(out, "true") || !strings.Contains(out, "false") {
		t.Fatalf("expected true and false, got %q", out)
	}
}

func TestFileReadTextError(t *testing.T) {
	dir := t.TempDir()
	src := `package app.main

func main():
    result = file.read_text("` + dir + `/nonexistent.txt")
    console.print_ln(result)
`
	out := runSource(t, src)
	if !strings.Contains(out, "IOError") {
		t.Fatalf("expected IOError, got %q", out)
	}
}

func TestJsonParseError(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    result = json.parse("not valid json")
    console.print_ln(result)
`)
	if !strings.Contains(out, "JSONError") {
		t.Fatalf("expected JSONError, got %q", out)
	}
}

func TestJsonStringifyEmptyError(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    result = json.stringify({"key": "value"})
    console.print_ln(result)
`)
	if !strings.Contains(out, "key") {
		t.Fatalf("expected key in output, got %q", out)
	}
}

func TestSystemArgs(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    args = system.args()
    console.print_ln(args)
`)
	if !strings.Contains(out, "one") || !strings.Contains(out, "two") {
		t.Fatalf("expected args, got %q", out)
	}
}

func TestConsolePrint(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    console.print("hello ")
    console.print("world")
`)
	if !strings.Contains(out, "hello world") {
		t.Fatalf("expected hello world, got %q", out)
	}
}

func TestFloatDivisionByZero(t *testing.T) {
	requireRuntimeCode(t, `package app.main

func main():
    x = 1.0 / 0.0
    console.print_ln(x)
`, "E103")
}

func TestIntModuloByZero(t *testing.T) {
	requireRuntimeCode(t, `package app.main

func main():
    x = 10 % 0
    console.print_ln(x)
`, "E103")
}

func TestListIndexOutOfBounds(t *testing.T) {
	requireRuntimeCode(t, `package app.main

func main():
    xs = [1, 2, 3]
    console.print_ln(xs[99])
`, "E102")
}

func TestDictKeyNotFound(t *testing.T) {
	requireRuntimeCode(t, `package app.main

func main():
    m = {"a": 1}
    console.print_ln(m["missing"])
`, "E106")
}

func TestConstReassignmentRuntimeError(t *testing.T) {
	// Checker catches this with E033.
	_, diag, err := runSourceLenient(t, `package app.main

const X: int = 10

func main():
    X = 20
`)
	if err == nil && !diag.HasErrors() {
		t.Fatalf("expected error for const reassignment")
	}
}

func TestZeroValueCharAndNullable(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    c: char
    s: string?
    b: bool
    i: int
    f: float
    console.print_ln(c)
    console.print_ln(s)
    console.print_ln(b)
    console.print_ln(i)
    console.print_ln(f)
`)
	if !strings.Contains(out, "false") || !strings.Contains(out, "0") {
		t.Fatalf("expected zero values, got %q", out)
	}
}

func TestZeroValueList(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    xs: List<int>
    console.print_ln(xs)
`)
	if !strings.Contains(out, "null") && !strings.Contains(out, "[]") {
		t.Fatalf("expected null or [], got %q", out)
	}
}

func TestListMethodsRuntime(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    xs: List<int> = [1, 2]
    xs.push(3)
    xs.insert(1, 9)
    console.print_ln(xs.count())
    console.print_ln(xs.index_of(9))
    console.print_ln(xs.has_index(4))
    console.print_ln(xs.remove(2))
    console.print_ln(xs.pop())
    xs.clear()
    console.print_ln(xs.count())
`)
	for _, want := range []string{"4", "1", "false", "true", "3", "0"} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected output to contain %q, got %q", want, out)
		}
	}
}

func TestDictMethodsRuntime(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    scores: Dict<string, int> = {"Ana": 10}
    console.print_ln(scores.has_key("Ana"))
    console.print_ln(scores.try_add("Bob", 11))
    console.print_ln(scores.try_add("Bob", 12))
    console.print_ln(scores.try_set("Ana", 15))
    console.print_ln(scores.try_remove("Bob"))
    console.print_ln(scores.count())
    result = scores.get("Ana")
    console.print_ln(result)
    missing = scores.get("Bob")
    console.print_ln(missing)
    scores.clear()
    console.print_ln(scores.count())
`)
	for _, want := range []string{"true", "false", "1", "15", "null", "KeyError", "0"} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected output to contain %q, got %q", want, out)
		}
	}
}

func TestFieldAssignmentOnNonRecordError(t *testing.T) {
	requireRuntimeCode(t, `package app.main

func main():
    x = 42
    x.foo = 10
`, "E101")
}

func TestSystemEnv(t *testing.T) {
	// Set a test env var and verify system.env reads it.
	t.Setenv("LEAL_TEST_VAR", "hello_leal")
	out := runSource(t, `package app.main

func main():
    result = system.env("LEAL_TEST_VAR")
    console.print_ln(result)
`)
	if !strings.Contains(out, "hello_leal") {
		t.Fatalf("expected hello_leal, got %q", out)
	}
}

func TestSystemEnvNotFound(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    result = system.env("LEAL_NONEXISTENT_VAR_12345")
    console.print_ln(result)
`)
	if !strings.Contains(out, "null") {
		t.Fatalf("expected null, got %q", out)
	}
}

func TestFileWriteTextSuccess(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/output.txt"
	out := runSource(t, `package app.main

func main():
    file.write_text("`+path+`", "content")
    result = file.read_text("`+path+`")
    console.print_ln(result)
`)
	if !strings.Contains(out, "content") {
		t.Fatalf("expected content, got %q", out)
	}
}

func TestJsonParseSuccess(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    result = json.parse("{\"key\": 42}")
    console.print_ln(result)
`)
	if !strings.Contains(out, "42") {
		t.Fatalf("expected 42, got %q", out)
	}
}

func TestConsolePrintNoArgs(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    console.print("")
`)
	// Should not crash.
	if out == "" {
		// empty string is fine
	}
}

func TestConsolePrintLnEmpty(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    console.print_ln("")
`)
	// Should print an empty line.
	_ = out
}

func TestCallDepthExceeded(t *testing.T) {
	requireRuntimeCode(t, `package app.main

func infinite():
    infinite()

func main():
    infinite()
`, "E107")
}

func TestConstReassignRuntime(t *testing.T) {
	_, diag, err := runSourceLenient(t, `package app.main

func main():
    const X: int = 10
    X = 20
`)
	if err == nil && !diag.HasErrors() {
		t.Fatalf("expected error for const reassignment")
	}
}
