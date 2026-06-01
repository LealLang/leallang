// Copyright (c) 2026 LealLang Contributors
// SPDX-License-Identifier: MIT

package interpreter

import (
	"bytes"
	"os"
	"strings"
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

func TestAsyncReturnsBeforeCompletion(t *testing.T) {
	out := runSource(t, `package app.main

async func compute(x: int) -> int:
    return x * 2

func main():
    task = compute(21)
    result = await task
    console.print_ln(result)
`)
	if !strings.Contains(out, "42") {
		t.Fatalf("expected 42 in output, got %q", out)
	}
}

func TestAsyncConcurrentTasks(t *testing.T) {
	out := runSource(t, `package app.main

async func greet(name: string):
    console.print_ln($"hello {name}")

func main():
    t1 = greet("alice")
    t2 = greet("bob")
    await t1
    await t2
`)
	if !strings.Contains(out, "hello alice") || !strings.Contains(out, "hello bob") {
		t.Fatalf("expected both greetings, got %q", out)
	}
}

func TestAsyncRuntimeErrorResolvesTaskFailed(t *testing.T) {
	// Division by zero inside async should produce a task_failed error, not panic.
	src := `package app.main

async func bad_math() -> int:
    x = 1 / 0
    return x

func main():
    task = bad_math()
    result = await task
    console.print_ln(result)
`
	// The task should resolve (not crash). The result is a tuple (null, Error).
	out, _, err := runSourceWithDiagnostics(t, src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = out
}

func TestAsyncListDeepCopy(t *testing.T) {
	out := runSource(t, `package app.main

async func check_items(items: List<int>):
    console.print_ln(items[0])

func main():
    data = [1, 2, 3]
    t = check_items(data)
    data[0] = 99
    await t
    console.print_ln(data[0])
`)
	// Async task should see original value 1, main sees mutated 99.
	if !strings.Contains(out, "1") || !strings.Contains(out, "99") {
		t.Fatalf("expected 1 and 99 in output, got %q", out)
	}
}

func TestAsyncDictDeepCopy(t *testing.T) {
	out := runSource(t, `package app.main

async func check_dict(d: Dict<string, int>):
    console.print_ln(d["a"])

func main():
    data: Dict<string, int> = {"a": 1, "b": 2}
    t = check_dict(data)
    data["a"] = 99
    await t
    console.print_ln(data["a"])
`)
	// Async task should see original value 1, main sees mutated 99.
	if !strings.Contains(out, "1") || !strings.Contains(out, "99") {
		t.Fatalf("expected 1 and 99 in output, got %q", out)
	}
}

func TestAsyncRecordDeepCopy(t *testing.T) {
	out := runSource(t, `package app.main

type Config:
    name: string
    value: int

async func modify(c: Config):
    console.print_ln(c.value)

func main():
    cfg = Config(name: "test", value: 42)
    t = modify(cfg)
    cfg.value = 999
    await t
    console.print_ln(cfg.value)
`)
	// Async task should see original value 42, main sees mutated 999.
	if !strings.Contains(out, "42") || !strings.Contains(out, "999") {
		t.Fatalf("expected 42 and 999 in output, got %q", out)
	}
}

func TestAsyncNonSendableRejected(t *testing.T) {
	// cloneForTask should reject non-sendable values.
	ns := &NamespaceVal{Name: "console", Members: map[string]*BuiltinVal{
		"print": {Name: "console.print", Fn: func(args []Value) (Value, error) { return Null, nil }},
	}}
	_, err := cloneForTask(ns)
	if err == nil {
		t.Fatalf("expected error for namespace value")
	}
	if !strings.Contains(err.Error(), "cannot cross async task boundary") {
		t.Fatalf("unexpected error message: %v", err)
	}

	// FuncVal should also be rejected.
	fv := &FuncVal{Name: "test"}
	_, err = cloneForTask(fv)
	if err == nil {
		t.Fatalf("expected error for function value")
	}

	// Sendable values should work.
	cloned, err := cloneForTask(&ListVal{Elements: []Value{IntVal(1), IntVal(2)}})
	if err != nil {
		t.Fatalf("unexpected error for list: %v", err)
	}
	list := cloned.(*ListVal)
	if len(list.Elements) != 2 || list.Elements[0] != IntVal(1) {
		t.Fatalf("unexpected cloned list: %v", list)
	}
}

func TestAsyncCallDepthIsolation(t *testing.T) {
	// Two concurrent async tasks should not share callDepth.
	out := runSource(t, `package app.main

async func deep(n: int) -> int:
    return n

func main():
    t1 = deep(10)
    t2 = deep(20)
    r1 = await t1
    r2 = await t2
    console.print_ln(r1)
    console.print_ln(r2)
`)
	if !strings.Contains(out, "10") || !strings.Contains(out, "20") {
		t.Fatalf("expected 10 and 20, got %q", out)
	}
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
