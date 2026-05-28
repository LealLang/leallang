// Copyright (c) 2026 LealLang Contributors
// SPDX-License-Identifier: MIT

package interpreter

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/LealLang/leallang/internal/checker"
	"github.com/LealLang/leallang/internal/diagnostics"
	"github.com/LealLang/leallang/internal/lexer"
	"github.com/LealLang/leallang/internal/parser"
)

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
	var stdout, stderr bytes.Buffer
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

func TestHelloWorld(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    console.print_ln("Hello, world!")
`)
	if !strings.Contains(out, "Hello, world!\n") {
		t.Fatalf("unexpected output %q", out)
	}
}

func TestArithmeticAndStringConcat(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    console.print_ln(2 + 3 * 4)
    console.print_ln("Hello, " + "world")
`)
	want := "14\nHello, world\n"
	if out != want {
		t.Fatalf("got %q, want %q", out, want)
	}
}

func TestVarDeclAndIfElse(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    x: int = 2
    y = 3
    if x + y == 4:
        console.print_ln("if")
    else if x + y == 5:
        console.print_ln("elseif")
    else:
        console.print_ln("else")
`)
	if out != "elseif\n" {
		t.Fatalf("unexpected output %q", out)
	}
}

func TestSwitchStmtAndExpr(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    x = 2
    switch x:
        1:
            console.print_ln("one")
        _:
            console.print_ln("other")
    result: string = switch x:
        2 -> "two"
        _ -> "many"
    console.print_ln(result)
`)
	if out != "other\ntwo\n" {
		t.Fatalf("unexpected output %q", out)
	}
}

func TestRangeLoopVariants(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    loop i in 0..3:
        console.print(i)
    console.print("|")
    loop i in 0..<3:
        console.print(i)
    console.print("|")
    loop i in 0..10, step 3:
        console.print(i)
`)
	if out != "0123|012|0369" {
		t.Fatalf("unexpected output %q", out)
	}
}

func TestLoopModifiersBreakContinueAndInfinite(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    n = 0
    loop i in 0..10, while n < 5, if i != 1:
        n = n + 1
        if i == 3:
            continue
        console.print(i)
        if i == 5:
            break
    console.print("|")
    count = 0
    loop true:
        count = count + 1
        if count == 3:
            break
    console.print(count)
`)
	if out != "0245|3" {
		t.Fatalf("unexpected output %q", out)
	}
}

func TestCollectionLoopsAndIndexing(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    xs = [10, 20, 30]
    console.print_ln(xs[1])
    xs[1] = 25
    loop x in xs:
        console.print(x)
    console.print("|")
    m = {"a": 1}
    console.print(m["a"])
`)
	if out != "20\n102530|1" {
		t.Fatalf("unexpected output %q", out)
	}
}

func TestFunctionCallsReturnAndRecursion(t *testing.T) {
	out := runSource(t, `package app.main

func add(a: int, b: int) -> int:
    return a + b

func fact(n: int) -> int:
    if n <= 1:
        return 1
    return n * fact(n - 1)

func main():
    console.print_ln(add(b: 3, a: 2))
    console.print_ln(fact(5))
`)
	if out != "5\n120\n" {
		t.Fatalf("unexpected output %q", out)
	}
}

func TestRecordsConstructorsMethodsAndInterpolation(t *testing.T) {
	out := runSource(t, `package app.main

type Person:
    name: string
    age: int
    func Greet() -> string:
        return $"{name}:{age}"

type Point:
    X: int
    Y: int
    constructor(x: int, y: int):
        X = x
        Y = y
    func Sum() -> int:
        return X + Y

func main():
    p = Person(name: "Ana", age: 7)
    point = Point(x: 10, y: 20)
    console.print_ln(p.Greet())
    console.print_ln(point.Sum())
    console.print_ln($"Value: {point.X + 1}")
`)
	want := "Ana:7\n30\nValue: 11\n"
	if out != want {
		t.Fatalf("got %q, want %q", out, want)
	}
}

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

func TestExamplesFullRuns(t *testing.T) {
	src, err := os.ReadFile("../../examples/full.ll")
	if err != nil {
		t.Fatal(err)
	}
	out := runSource(t, string(src))
	if !strings.Contains(out, "i is 0\n") || !strings.Contains(out, "i is 48\n") {
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

func TestJsonRoundTrip(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    result = json.stringify({"name": "test", "value": "42"})
    console.print_ln(result)
`)
	if !strings.Contains(out, `"name"`) || !strings.Contains(out, `"value"`) {
		t.Fatalf("unexpected output %q", out)
	}
}
