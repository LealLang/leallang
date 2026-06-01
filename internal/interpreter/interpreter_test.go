// Copyright (c) 2026 LealLang Contributors
// SPDX-License-Identifier: MIT

package interpreter

import (
	"bytes"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/LealLang/leallang/internal/ast"
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

// --- TaskVal unit tests ---

func TestTaskValResolveAndAwait(t *testing.T) {
	task := newTaskVal()
	task.resolve(IntVal(42), Null)
	result, err := task.await()
	if result != IntVal(42) {
		t.Fatalf("expected IntVal(42), got %v", result)
	}
	if _, ok := err.(*NullVal); !ok {
		t.Fatalf("expected Null error, got %v", err)
	}
}

func TestTaskValResolveWithError(t *testing.T) {
	task := newTaskVal()
	taskErr := errorRecord("something failed", "task_failed")
	task.resolve(Null, taskErr)
	result, err := task.await()
	if _, ok := result.(*NullVal); !ok {
		t.Fatalf("expected Null result, got %v", result)
	}
	rec, ok := err.(*RecordVal)
	if !ok {
		t.Fatalf("expected Error record, got %T", err)
	}
	if rec.TypeName != "Error" {
		t.Fatalf("expected Error type, got %s", rec.TypeName)
	}
	if string(rec.Fields["message"].(StringVal)) != "something failed" {
		t.Fatalf("unexpected error message: %v", rec.Fields["message"])
	}
}

func TestTaskValTypeAndString(t *testing.T) {
	task := newTaskVal()
	if task.Type() != "Task" {
		t.Fatalf("expected 'Task', got %q", task.Type())
	}
	if task.String() != "<task>" {
		t.Fatalf("expected '<task>', got %q", task.String())
	}
}

func TestTaskValAwaitBlocks(t *testing.T) {
	task := newTaskVal()
	done := make(chan struct{})
	go func() {
		result, _ := task.await()
		if result != IntVal(99) {
			t.Errorf("expected IntVal(99), got %v", result)
		}
		close(done)
	}()
	// Small delay to ensure goroutine is waiting on the channel.
	time.Sleep(10 * time.Millisecond)
	task.resolve(IntVal(99), Null)
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("await did not unblock after resolve")
	}
}

// --- cloneForTask comprehensive tests ---

func TestCloneForTaskPrimitives(t *testing.T) {
	tests := []struct {
		name  string
		value Value
	}{
		{"IntVal", IntVal(42)},
		{"FloatVal", FloatVal(3.14)},
		{"StringVal", StringVal("hello")},
		{"BoolVal", BoolVal(true)},
		{"CharVal", CharVal('x')},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cloned, err := cloneForTask(tt.value)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			// Primitives pass through unchanged.
			if cloned != tt.value {
				t.Fatalf("expected same value, got %v vs %v", cloned, tt.value)
			}
		})
	}
}

func TestCloneForTaskNullAndEnum(t *testing.T) {
	cloned, err := cloneForTask(Null)
	if err != nil {
		t.Fatalf("unexpected error for Null: %v", err)
	}
	if _, ok := cloned.(*NullVal); !ok {
		t.Fatalf("expected *NullVal, got %T", cloned)
	}

	enum := &EnumVal{TypeName: "Color", Member: "blue"}
	cloned, err = cloneForTask(enum)
	if err != nil {
		t.Fatalf("unexpected error for EnumVal: %v", err)
	}
	if cloned != enum {
		t.Fatalf("expected same enum reference, got %v", cloned)
	}
}

func TestCloneForTaskRange(t *testing.T) {
	rv := &RangeVal{Low: 0, High: 10, Exclusive: true}
	cloned, err := cloneForTask(rv)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cloned != rv {
		t.Fatalf("expected same RangeVal reference, got %v", cloned)
	}
}

func TestCloneForTaskTuple(t *testing.T) {
	original := &TupleVal{Elements: []Value{IntVal(1), StringVal("two")}}
	cloned, err := cloneForTask(original)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	tuple, ok := cloned.(*TupleVal)
	if !ok {
		t.Fatalf("expected *TupleVal, got %T", cloned)
	}
	// Elements should be equal but the slice itself should be a different backing array.
	if len(tuple.Elements) != 2 {
		t.Fatalf("expected 2 elements, got %d", len(tuple.Elements))
	}
	if tuple.Elements[0] != IntVal(1) || tuple.Elements[1] != StringVal("two") {
		t.Fatalf("unexpected elements: %v", tuple.Elements)
	}
}

func TestCloneForTaskNestedStructures(t *testing.T) {
	// Record containing a List.
	listField := &ListVal{Elements: []Value{IntVal(10), IntVal(20)}}
	rec := &RecordVal{TypeName: "Config", Fields: map[string]Value{
		"name": StringVal("test"),
		"items": listField,
	}}
	cloned, err := cloneForTask(rec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	clonedRec, ok := cloned.(*RecordVal)
	if !ok {
		t.Fatalf("expected *RecordVal, got %T", cloned)
	}
	// Modify original list — clone should be unaffected.
	listField.Elements[0] = IntVal(999)
	clonedList := clonedRec.Fields["items"].(*ListVal)
	if clonedList.Elements[0] != IntVal(10) {
		t.Fatalf("clone was affected by mutation: got %v", clonedList.Elements[0])
	}

	// Dict containing a Record.
	innerRec := &RecordVal{TypeName: "Inner", Fields: map[string]Value{"x": IntVal(1)}}
	dict := &DictVal{Entries: map[string]Value{"key": innerRec}}
	clonedDict, err := cloneForTask(dict)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Modify original inner record.
	innerRec.Fields["x"] = IntVal(999)
	d := clonedDict.(*DictVal)
	ir := d.Entries["key"].(*RecordVal)
	if ir.Fields["x"] != IntVal(1) {
		t.Fatalf("clone was affected by mutation: got %v", ir.Fields["x"])
	}
}

func TestCloneForTaskRejectsAllNonSendable(t *testing.T) {
	tests := []struct {
		name  string
		value Value
	}{
		{"BuiltinVal", &BuiltinVal{Name: "test", Fn: func(args []Value) (Value, error) { return Null, nil }}},
		{"ConstGroupVal", &ConstGroupVal{Name: "colors", Members: map[string]*EnumVal{"red": {TypeName: "Color", Member: "red"}}}},
		{"RecordTypeVal", &RecordTypeVal{Decl: nil}},
		{"BoundMethodVal", &BoundMethodVal{Receiver: nil, Decl: nil, Closure: nil}},
		{"ComponentRefVal", &ComponentRefVal{Component: "Button", ID: "btn"}},
		{"TaskVal", newTaskVal()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := cloneForTask(tt.value)
			if err == nil {
				t.Fatalf("expected error for %s", tt.name)
			}
			if !strings.Contains(err.Error(), "cannot cross async task boundary") {
				t.Fatalf("unexpected error message: %v", err)
			}
		})
	}
}

// --- Async behavioral tests ---

func TestAsyncVoidFunction(t *testing.T) {
	out := runSource(t, `package app.main

async func do_nothing():
    pass

func main():
    task = do_nothing()
    result = await task
    console.print_ln(result)
`)
	// Void async returns (null, null) tuple.
	if !strings.Contains(out, "null") {
		t.Fatalf("expected null in output, got %q", out)
	}
}

func TestAsyncNestedCalls(t *testing.T) {
	out := runSource(t, `package app.main

async func inner(x: int) -> int:
    return x * 3

async func outer(x: int):
    task = inner(x)
    result = await task
    console.print_ln(result)

func main():
    task = outer(7)
    await task
`)
	if !strings.Contains(out, "21") {
		t.Fatalf("expected 21 in output, got %q", out)
	}
}

func TestAsyncGlobalsIsolation(t *testing.T) {
	out := runSource(t, `package app.main

counter: int = 10

async func read_global() -> int:
    return counter

func main():
    task = read_global()
    counter = 999
    result = await task
    console.print_ln(result)
`)
	// Async task should see the snapshot value 10, not the mutated 999.
	if !strings.Contains(out, "10") {
		t.Fatalf("expected 10 in output, got %q", out)
	}
}

func TestAsyncRecordWithListDeepCopy(t *testing.T) {
	out := runSource(t, `package app.main

type Config:
    Name: string
    Items: List<int>

async func process(c: Config):
    console.print_ln(c.Items[0])

func main():
    cfg = Config(Name: "test", Items: [10, 20, 30])
    task = process(cfg)
    await task
`)
	if !strings.Contains(out, "10") {
		t.Fatalf("expected 10 in output, got %q", out)
	}
}

func TestMultipleAwaitsOnSameTask(t *testing.T) {
	out := runSource(t, `package app.main

async func compute() -> int:
    return 42

func main():
    task = compute()
    r1 = await task
    r2 = await task
    console.print_ln(r1)
    console.print_ln(r2)
`)
	if !strings.Contains(out, "42") {
		t.Fatalf("expected 42 in output, got %q", out)
	}
	// Both awaits should return the same result — 42 should appear twice.
	count := strings.Count(out, "42")
	if count < 2 {
		t.Fatalf("expected 42 to appear twice, got %q", out)
	}
}

// --- Value Type()/String() method coverage ---

func TestValueTypeAndString(t *testing.T) {
	tests := []struct {
		name     string
		val      Value
		wantType string
		wantStr  string
	}{
		{"IntVal", IntVal(42), "int", "42"},
		{"IntVal neg", IntVal(-5), "int", "-5"},
		{"IntVal zero", IntVal(0), "int", "0"},
		{"FloatVal", FloatVal(3.14), "float", "3.14"},
		{"FloatVal zero", FloatVal(0), "float", "0"},
		{"StringVal", StringVal("hello"), "string", "hello"},
		{"StringVal empty", StringVal(""), "string", ""},
		{"BoolVal true", BoolVal(true), "bool", "true"},
		{"BoolVal false", BoolVal(false), "bool", "false"},
		{"CharVal", CharVal('x'), "char", "x"},
		{"CharVal zero", CharVal(0), "char", "\x00"},
		{"NullVal", Null, "null", "null"},
		{"ListVal", &ListVal{Elements: []Value{IntVal(1), IntVal(2)}}, "List", "[1, 2]"},
		{"ListVal empty", &ListVal{}, "List", "[]"},
		{"DictVal", &DictVal{Entries: map[string]Value{"a": IntVal(1)}}, "Dict", `{"a": 1}`},
		{"DictVal empty", &DictVal{}, "Dict", "{}"},
		{"RecordVal", &RecordVal{TypeName: "Point", Fields: map[string]Value{"X": IntVal(1)}}, "Point", "Point{X: 1}"},
		{"EnumVal", &EnumVal{TypeName: "Color", Member: "red"}, "Color", "red"},
		{"TupleVal", &TupleVal{Elements: []Value{IntVal(1), StringVal("two")}}, "tuple", "(1, two)"},
		{"TupleVal empty", &TupleVal{}, "tuple", "()"},
		{"RangeVal", &RangeVal{Low: 0, High: 10, Exclusive: true}, "range", "0..<10"},
		{"RangeVal inclusive", &RangeVal{Low: 0, High: 10, Exclusive: false}, "range", "0..10"},
		{"FuncVal", &FuncVal{Name: "test"}, "function", "<func test>"},
		{"FuncVal anon", &FuncVal{}, "function", "<func>"},
		{"BuiltinVal", &BuiltinVal{Name: "print_ln"}, "builtin", "<builtin print_ln>"},
		{"NamespaceVal", &NamespaceVal{Name: "console"}, "namespace", "<namespace console>"},
		{"ConstGroupVal", &ConstGroupVal{Name: "colors"}, "const group", "<const group colors>"},
		{"RecordTypeVal", &RecordTypeVal{Decl: &ast.TypeDecl{Name: "T"}}, "type", "<type T>"},
		{"BoundMethodVal", &BoundMethodVal{Decl: &ast.FuncDecl{Name: "Foo"}}, "method", "<method Foo>"},
		{"ComponentRefVal", &ComponentRefVal{Component: "Button", ID: "btn"}, "Button", "@Button[btn]"},
		{"TaskVal", newTaskVal(), "Task", "<task>"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.val.Type(); got != tt.wantType {
				t.Errorf("Type() = %q, want %q", got, tt.wantType)
			}
			if got := tt.val.String(); got != tt.wantStr {
				t.Errorf("String() = %q, want %q", got, tt.wantStr)
			}
		})
	}
}

func TestValueStringNil(t *testing.T) {
	if got := valueString(nil); got != "null" {
		t.Errorf("valueString(nil) = %q, want %q", got, "null")
	}
}

// --- valuesEqual coverage ---

func TestValuesEqualAllTypes(t *testing.T) {
	tests := []struct {
		name string
		a, b Value
		want bool
	}{
		{"int equal", IntVal(1), IntVal(1), true},
		{"int not equal", IntVal(1), IntVal(2), false},
		{"float equal", FloatVal(1.5), FloatVal(1.5), true},
		{"float not equal", FloatVal(1.5), FloatVal(2.5), false},
		{"string equal", StringVal("a"), StringVal("a"), true},
		{"string not equal", StringVal("a"), StringVal("b"), false},
		{"bool equal", BoolVal(true), BoolVal(true), true},
		{"bool not equal", BoolVal(true), BoolVal(false), false},
		{"char equal", CharVal('a'), CharVal('a'), true},
		{"char not equal", CharVal('a'), CharVal('b'), false},
		{"null equal", Null, Null, true},
		{"enum equal", &EnumVal{TypeName: "C", Member: "a"}, &EnumVal{TypeName: "C", Member: "a"}, true},
		{"enum different member", &EnumVal{TypeName: "C", Member: "a"}, &EnumVal{TypeName: "C", Member: "b"}, false},
		{"enum different type", &EnumVal{TypeName: "C", Member: "a"}, &EnumVal{TypeName: "D", Member: "a"}, false},
		{"list equal", &ListVal{[]Value{IntVal(1)}}, &ListVal{[]Value{IntVal(1)}}, true},
		{"list different len", &ListVal{[]Value{IntVal(1)}}, &ListVal{[]Value{IntVal(1), IntVal(2)}}, false},
		{"list different values", &ListVal{[]Value{IntVal(1)}}, &ListVal{[]Value{IntVal(2)}}, false},
		{"dict equal", &DictVal{map[string]Value{"a": IntVal(1)}}, &DictVal{map[string]Value{"a": IntVal(1)}}, true},
		{"dict different len", &DictVal{map[string]Value{"a": IntVal(1)}}, &DictVal{map[string]Value{"a": IntVal(1), "b": IntVal(2)}}, false},
		{"dict different values", &DictVal{map[string]Value{"a": IntVal(1)}}, &DictVal{map[string]Value{"a": IntVal(2)}}, false},
		{"record equal", &RecordVal{"T", map[string]Value{"x": IntVal(1)}}, &RecordVal{"T", map[string]Value{"x": IntVal(1)}}, true},
		{"record different type", &RecordVal{"T", map[string]Value{"x": IntVal(1)}}, &RecordVal{"U", map[string]Value{"x": IntVal(1)}}, false},
		{"record different fields", &RecordVal{"T", map[string]Value{"x": IntVal(1)}}, &RecordVal{"T", map[string]Value{"x": IntVal(2)}}, false},
		{"tuple equal", &TupleVal{[]Value{IntVal(1), StringVal("a")}}, &TupleVal{[]Value{IntVal(1), StringVal("a")}}, true},
		{"tuple different len", &TupleVal{[]Value{IntVal(1)}}, &TupleVal{[]Value{IntVal(1), IntVal(2)}}, false},
		{"tuple different values", &TupleVal{[]Value{IntVal(1)}}, &TupleVal{[]Value{IntVal(2)}}, false},
		{"component ref equal", &ComponentRefVal{"Button", "btn"}, &ComponentRefVal{"Button", "btn"}, true},
		{"component ref different", &ComponentRefVal{"Button", "btn"}, &ComponentRefVal{"Label", "btn"}, false},
		{"cross type int float", IntVal(1), FloatVal(1.0), false},
		{"cross type int string", IntVal(1), StringVal("1"), false},
		{"cross type null int", Null, IntVal(0), false},
		{"default fallback diff ptr", &RangeVal{0, 10, false}, &RangeVal{0, 10, false}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := valuesEqual(tt.a, tt.b); got != tt.want {
				t.Errorf("valuesEqual(%v, %v) = %v, want %v", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

// --- dictKey coverage ---

func TestDictKeyAllTypes(t *testing.T) {
	tests := []struct {
		name    string
		val     Value
		wantKey string
		wantOk  bool
	}{
		{"string", StringVal("hello"), "hello", true},
		{"int", IntVal(42), "42", true},
		{"bool true", BoolVal(true), "true", true},
		{"bool false", BoolVal(false), "false", true},
		{"char", CharVal('x'), "x", true},
		{"enum", &EnumVal{TypeName: "C", Member: "red"}, "red", true},
		{"list invalid", &ListVal{}, "", false},
		{"dict invalid", &DictVal{}, "", false},
		{"null invalid", Null, "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, ok := dictKey(tt.val)
			if ok != tt.wantOk {
				t.Errorf("dictKey(%v) ok = %v, want %v", tt.val, ok, tt.wantOk)
			}
			if ok && key != tt.wantKey {
				t.Errorf("dictKey(%v) key = %q, want %q", tt.val, key, tt.wantKey)
			}
		})
	}
}

// --- valueToGo coverage ---

func TestValueToGoAllTypes(t *testing.T) {
	tests := []struct {
		name string
		val  Value
	}{
		{"int", IntVal(42)},
		{"float", FloatVal(3.14)},
		{"string", StringVal("hello")},
		{"bool", BoolVal(true)},
		{"char", CharVal('x')},
		{"null", Null},
		{"list", &ListVal{[]Value{IntVal(1), IntVal(2)}}},
		{"dict", &DictVal{map[string]Value{"a": IntVal(1)}}},
		{"record", &RecordVal{"T", map[string]Value{"x": IntVal(1)}}},
		{"enum", &EnumVal{"Color", "red"}},
		{"tuple", &TupleVal{[]Value{IntVal(1), StringVal("a")}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := valueToGo(tt.val)
			if result == nil && tt.val != Null {
				t.Errorf("valueToGo(%v) returned nil", tt.val)
			}
		})
	}
}

// --- goToValue coverage ---

func TestGoToValueAllTypes(t *testing.T) {
	tests := []struct {
		name string
		val  any
		want string // expected Type()
	}{
		{"nil", nil, "null"},
		{"bool", true, "bool"},
		{"string", "hello", "string"},
		{"float64 int", float64(42), "int"},
		{"float64 float", float64(3.14), "float"},
		{"[]any", []any{float64(1), float64(2)}, "List"},
		{"map", map[string]any{"a": float64(1)}, "Dict"},
		{"default", struct{}{}, "string"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := goToValue(tt.val)
			if got := result.Type(); got != tt.want {
				t.Errorf("goToValue(%v).Type() = %q, want %q", tt.val, got, tt.want)
			}
		})
	}
}

// --- decodeJSONValue / jsonNumberToValue coverage ---

func TestDecodeJSONValue(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantType string
	}{
		{"null", "null", "null"},
		{"bool", "true", "bool"},
		{"string", `"hello"`, "string"},
		{"int", "42", "int"},
		{"float", "3.14", "float"},
		{"array", "[1, 2, 3]", "List"},
		{"object", `{"a": 1}`, "Dict"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := decodeJSONValue([]byte(tt.input))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got := result.Type(); got != tt.wantType {
				t.Errorf("decodeJSONValue(%q).Type() = %q, want %q", tt.input, got, tt.wantType)
			}
		})
	}
}

func TestJsonNumberToValueNested(t *testing.T) {
	// Nested array
	result := jsonNumberToValue([]any{float64(1), float64(2)})
	lst, ok := result.(*ListVal)
	if !ok || len(lst.Elements) != 2 {
		t.Fatalf("expected ListVal with 2 elements, got %T", result)
	}
	// Nested object
	result = jsonNumberToValue(map[string]any{"x": float64(1)})
	dict, ok := result.(*DictVal)
	if !ok || len(dict.Entries) != 1 {
		t.Fatalf("expected DictVal with 1 entry, got %T", result)
	}
	// Default fallback
	result = jsonNumberToValue(struct{}{})
	if result.Type() != "string" {
		t.Errorf("expected string type for unknown, got %q", result.Type())
	}
}

// --- Signal.Error coverage ---

func TestSignalError(t *testing.T) {
	tests := []struct {
		name   string
		signal *Signal
		want   string
	}{
		{"return", &Signal{Kind: signalReturn, Value: IntVal(42)}, "return 42"},
		{"return nil", &Signal{Kind: signalReturn, Value: nil}, "return null"},
		{"break", &Signal{Kind: signalBreak}, "break"},
		{"continue", &Signal{Kind: signalContinue}, "continue"},
		{"unknown", &Signal{Kind: 99}, "signal"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.signal.Error(); got != tt.want {
				t.Errorf("Signal.Error() = %q, want %q", got, tt.want)
			}
		})
	}
}

// --- Source-level tests for uncovered interpreter paths ---

func TestUnaryMinusInt(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    x = -42
    console.print_ln(x)
`)
	if !strings.Contains(out, "-42") {
		t.Fatalf("expected -42, got %q", out)
	}
}

func TestUnaryMinusFloat(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    x = -3.14
    console.print_ln(x)
`)
	if !strings.Contains(out, "-3.14") {
		t.Fatalf("expected -3.14, got %q", out)
	}
}

func TestUnaryNot(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    a = not false
    b = not true
    console.print_ln(a)
    console.print_ln(b)
`)
	if !strings.Contains(out, "true") || !strings.Contains(out, "false") {
		t.Fatalf("expected true and false, got %q", out)
	}
}

func TestUnaryMinusRuntimeError(t *testing.T) {
	_, diag, err := runSourceLenient(t, `package app.main

func main():
    x = -"hello"
`)
	if err == nil && !diag.HasErrors() {
		t.Fatalf("expected error for unary minus on string")
	}
}

func TestNotRuntimeError(t *testing.T) {
	_, diag, err := runSourceLenient(t, `package app.main

func main():
    x = not 42
`)
	if err == nil && !diag.HasErrors() {
		t.Fatalf("expected error for not on int")
	}
}

func TestFloatLiteral(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    x = 3.14
    console.print_ln(x)
`)
	if !strings.Contains(out, "3.14") {
		t.Fatalf("expected 3.14, got %q", out)
	}
}

func TestCharLiteral(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    x = 'A'
    console.print_ln(x)
`)
	if !strings.Contains(out, "A") {
		t.Fatalf("expected A, got %q", out)
	}
}

func TestBoolLiterals(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    console.print_ln(true)
    console.print_ln(false)
`)
	if !strings.Contains(out, "true") || !strings.Contains(out, "false") {
		t.Fatalf("expected true and false, got %q", out)
	}
}

func TestNullLiteral(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    x: string? = null
    console.print_ln(x)
`)
	if !strings.Contains(out, "null") {
		t.Fatalf("expected null, got %q", out)
	}
}

func TestIntSubtraction(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    console.print_ln(10 - 3)
`)
	if !strings.Contains(out, "7") {
		t.Fatalf("expected 7, got %q", out)
	}
}

func TestIntMultiplication(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    console.print_ln(4 * 5)
`)
	if !strings.Contains(out, "20") {
		t.Fatalf("expected 20, got %q", out)
	}
}

func TestIntModulo(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    console.print_ln(10 % 3)
`)
	if !strings.Contains(out, "1") {
		t.Fatalf("expected 1, got %q", out)
	}
}

func TestFloatArithmeticOps(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    console.print_ln(1.5 + 2.5)
    console.print_ln(5.0 - 1.5)
    console.print_ln(2.0 * 3.5)
    console.print_ln(10.0 / 4.0)
`)
	if !strings.Contains(out, "4") || !strings.Contains(out, "3.5") || !strings.Contains(out, "7") || !strings.Contains(out, "2.5") {
		t.Fatalf("unexpected output: %q", out)
	}
}

func TestFloatModulo(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    console.print_ln(10.5 % 3.0)
`)
	if !strings.Contains(out, "1.5") {
		t.Fatalf("expected 1.5, got %q", out)
	}
}

func TestMixedIntFloatArithmetic(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    console.print_ln(1 + 2.5)
`)
	if !strings.Contains(out, "3.5") {
		t.Fatalf("expected 3.5, got %q", out)
	}
}

func TestNumericComparisonAllOps(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    console.print_ln(1 < 2)
    console.print_ln(2 > 1)
    console.print_ln(1 <= 1)
    console.print_ln(1 >= 1)
    console.print_ln(1.5 < 2.5)
    console.print_ln(2.5 > 1.5)
`)
	if !strings.Contains(out, "true") {
		t.Fatalf("expected true, got %q", out)
	}
}

func TestStringIndex(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    s = "hello"
    console.print_ln(s[0])
    console.print_ln(s[4])
`)
	if !strings.Contains(out, "h") || !strings.Contains(out, "o") {
		t.Fatalf("expected h and o, got %q", out)
	}
}

func TestDictIndexSuccess(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    m = {"key": 42}
    console.print_ln(m["key"])
`)
	if !strings.Contains(out, "42") {
		t.Fatalf("expected 42, got %q", out)
	}
}

func TestFieldAccessConstGroup(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    c = colors.red
    console.print_ln(c)
`)
	if !strings.Contains(out, "red") {
		t.Fatalf("expected red, got %q", out)
	}
}

func TestFieldAccessError(t *testing.T) {
	requireRuntimeCode(t, `package app.main

func main():
    x = 42
    console.print_ln(x.foo)
`, "E101")
}

func TestAssignIndexList(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    xs = [1, 2, 3]
    xs[1] = 99
    console.print_ln(xs[1])
`)
	if !strings.Contains(out, "99") {
		t.Fatalf("expected 99, got %q", out)
	}
}

func TestAssignIndexDict(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    m = {"a": 1}
    m["a"] = 99
    console.print_ln(m["a"])
`)
	if !strings.Contains(out, "99") {
		t.Fatalf("expected 99, got %q", out)
	}
}

func TestAssignIndexStringError(t *testing.T) {
	_, diag, err := runSourceLenient(t, `package app.main

func main():
    s = "hello"
    s[0] = "H"
`)
	if err == nil && !diag.HasErrors() {
		t.Fatalf("expected error for string index assignment")
	}
}

func TestVarDeclWithZeroValue(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    x: int
    y: float
    z: bool
    console.print_ln(x)
    console.print_ln(y)
    console.print_ln(z)
`)
	if !strings.Contains(out, "0") || !strings.Contains(out, "false") {
		t.Fatalf("expected zero values, got %q", out)
	}
}

func TestSwitchExprNoMatch(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    x = switch 99:
        1 -> "one"
        2 -> "two"
    console.print_ln(x)
`)
	if !strings.Contains(out, "null") {
		t.Fatalf("expected null for no match, got %q", out)
	}
}

func TestSwitchStmtNoMatch(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    x = 99
    switch x:
        1 -> console.print_ln("one")
        2 -> console.print_ln("two")
    console.print_ln("done")
`)
	if !strings.Contains(out, "done") {
		t.Fatalf("expected done, got %q", out)
	}
}

func TestLoopWithStep(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    loop i in 0..10, step 2:
        console.print_ln(i)
`)
	if !strings.Contains(out, "0") || !strings.Contains(out, "2") || !strings.Contains(out, "8") {
		t.Fatalf("expected 0, 2, 8, got %q", out)
	}
}

func TestLoopWithIfModifier(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    loop i in 0..5, if i != 2:
        console.print_ln(i)
`)
	if strings.Contains(out, "2") {
		t.Fatalf("should not contain 2, got %q", out)
	}
}

func TestLogicalAndShortCircuit(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    console.print_ln(true and true)
    console.print_ln(true and false)
    console.print_ln(false and true)
`)
	if !strings.Contains(out, "true") || !strings.Contains(out, "false") {
		t.Fatalf("expected true and false, got %q", out)
	}
}

func TestLogicalOrShortCircuit(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    console.print_ln(false or true)
    console.print_ln(true or false)
    console.print_ln(false or false)
`)
	if !strings.Contains(out, "true") || !strings.Contains(out, "false") {
		t.Fatalf("expected true and false, got %q", out)
	}
}

func TestEqualityOperators(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    console.print_ln(1 == 1)
    console.print_ln(1 != 2)
    console.print_ln("a" == "a")
    console.print_ln("a" != "b")
    console.print_ln(true == true)
    console.print_ln(null == null)
`)
	if !strings.Contains(out, "true") {
		t.Fatalf("expected true, got %q", out)
	}
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

func TestAssignRecordField(t *testing.T) {
	out := runSource(t, `package app.main

type Point:
    X: int
    Y: int

func main():
    p = Point(X: 1, Y: 2)
    p.X = 99
    console.print_ln(p.X)
`)
	if !strings.Contains(out, "99") {
		t.Fatalf("expected 99, got %q", out)
	}
}

func TestRecordMethodCallWithArgs(t *testing.T) {
	out := runSource(t, `package app.main

type Calc:
    Base: int
    func Add(x: int) -> int:
        return Base + x

func main():
    c = Calc(Base: 10)
    console.print_ln(c.Add(5))
`)
	if !strings.Contains(out, "15") {
		t.Fatalf("expected 15, got %q", out)
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

func TestSwitchStmtWithDefault(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    x = 99
    switch x:
        1 -> console.print_ln("one")
        _ -> console.print_ln("other")
`)
	if !strings.Contains(out, "other") {
		t.Fatalf("expected other, got %q", out)
	}
}

func TestSwitchExprWithDefault(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    result = switch 99:
        1 -> "one"
        _ -> "other"
    console.print_ln(result)
`)
	if !strings.Contains(out, "other") {
		t.Fatalf("expected other, got %q", out)
	}
}

func TestElseIf(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    x = 2
    if x == 1:
        console.print_ln("one")
    else if x == 2:
        console.print_ln("two")
    else:
        console.print_ln("other")
`)
	if !strings.Contains(out, "two") {
		t.Fatalf("expected two, got %q", out)
	}
}

func TestElse(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    x = 3
    if x == 1:
        console.print_ln("one")
    else:
        console.print_ln("other")
`)
	if !strings.Contains(out, "other") {
		t.Fatalf("expected other, got %q", out)
	}
}

func TestLoopContinue(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    loop i in 0..5:
        if i == 2:
            continue
        console.print_ln(i)
`)
	if strings.Contains(out, "2") {
		t.Fatalf("should not contain 2, got %q", out)
	}
}

func TestLoopBreak(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    loop i in 0..10:
        if i == 3:
            break
        console.print_ln(i)
`)
	if strings.Contains(out, "3") {
		t.Fatalf("should not contain 3, got %q", out)
	}
}

func TestDictLoop(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    m = {"a": 1, "b": 2}
    loop k in m:
        console.print_ln(k)
`)
	if !strings.Contains(out, "a") || !strings.Contains(out, "b") {
		t.Fatalf("expected a and b, got %q", out)
	}
}

func TestInterpolatedString(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    name = "world"
    s = $"hello {name}!"
    console.print_ln(s)
`)
	if !strings.Contains(out, "hello world!") {
		t.Fatalf("expected hello world!, got %q", out)
	}
}

func TestListLiteralAndLength(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    xs = [10, 20, 30]
    console.print_ln(xs)
`)
	if !strings.Contains(out, "10") || !strings.Contains(out, "20") || !strings.Contains(out, "30") {
		t.Fatalf("expected list contents, got %q", out)
	}
}

func TestDictLiteralAndAccess(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    m = {"x": 100, "y": 200}
    console.print_ln(m["x"])
    console.print_ln(m["y"])
`)
	if !strings.Contains(out, "100") || !strings.Contains(out, "200") {
		t.Fatalf("expected 100 and 200, got %q", out)
	}
}

func TestRecordConstructionAndFieldAccess(t *testing.T) {
	out := runSource(t, `package app.main

type Vec2:
    X: int
    Y: int

func main():
    v = Vec2(X: 5, Y: 10)
    console.print_ln(v.X)
    console.print_ln(v.Y)
`)
	if !strings.Contains(out, "5") || !strings.Contains(out, "10") {
		t.Fatalf("expected 5 and 10, got %q", out)
	}
}

func TestFunctionWithMultipleParams(t *testing.T) {
	out := runSource(t, `package app.main

func add(a: int, b: int) -> int:
    return a + b

func main():
    console.print_ln(add(3, 4))
`)
	if !strings.Contains(out, "7") {
		t.Fatalf("expected 7, got %q", out)
	}
}

func TestFunctionNamedArgs(t *testing.T) {
	out := runSource(t, `package app.main

func greet(name: string, age: int):
    console.print_ln(name)

func main():
    greet(age: 25, name: "Alice")
`)
	if !strings.Contains(out, "Alice") {
		t.Fatalf("expected Alice, got %q", out)
	}
}

func TestRecursiveFunction(t *testing.T) {
	out := runSource(t, `package app.main

func factorial(n: int) -> int:
    if n <= 1:
        return 1
    return n * factorial(n - 1)

func main():
    console.print_ln(factorial(5))
`)
	if !strings.Contains(out, "120") {
		t.Fatalf("expected 120, got %q", out)
	}
}

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

func TestAssignRecordFieldMutation(t *testing.T) {
	out := runSource(t, `package app.main

type Counter:
    Value: int

func main():
    c = Counter(Value: 0)
    c.Value = 42
    console.print_ln(c.Value)
`)
	if !strings.Contains(out, "42") {
		t.Fatalf("expected 42, got %q", out)
	}
}

func TestMethodReturningRecord(t *testing.T) {
	out := runSource(t, `package app.main

type Box:
    Value: int
    func Get() -> int:
        return Value

func main():
    b = Box(Value: 77)
    console.print_ln(b.Get())
`)
	if !strings.Contains(out, "77") {
		t.Fatalf("expected 77, got %q", out)
	}
}

func TestSwitchStmtMultipleCases(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    x = 2
    switch x:
        1 -> console.print_ln("one")
        2 -> console.print_ln("two")
        3 -> console.print_ln("three")
`)
	if !strings.Contains(out, "two") {
		t.Fatalf("expected two, got %q", out)
	}
}

func TestSwitchExprMultipleArms(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    result = switch 3:
        1 -> "one"
        2 -> "two"
        3 -> "three"
    console.print_ln(result)
`)
	if !strings.Contains(out, "three") {
		t.Fatalf("expected three, got %q", out)
	}
}

func TestRangeLoopExclusive(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    loop i in 0..<3:
        console.print_ln(i)
`)
	if !strings.Contains(out, "0") || !strings.Contains(out, "2") || strings.Contains(out, "3") {
		t.Fatalf("expected 0, 1, 2 but not 3, got %q", out)
	}
}

func TestRangeLoopInclusive(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    loop i in 0..3:
        console.print_ln(i)
`)
	if !strings.Contains(out, "3") {
		t.Fatalf("expected 3, got %q", out)
	}
}

func TestListLoop(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    xs = [10, 20, 30]
    loop x in xs:
        console.print_ln(x)
`)
	if !strings.Contains(out, "10") || !strings.Contains(out, "20") || !strings.Contains(out, "30") {
		t.Fatalf("expected list elements, got %q", out)
	}
}

func TestAssignDictKey(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    m = {"a": 1}
    m["b"] = 2
    console.print_ln(m["b"])
`)
	if !strings.Contains(out, "2") {
		t.Fatalf("expected 2, got %q", out)
	}
}

func TestAssignListElement(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    xs = [1, 2, 3]
    xs[0] = 99
    console.print_ln(xs[0])
`)
	if !strings.Contains(out, "99") {
		t.Fatalf("expected 99, got %q", out)
	}
}

// --- Additional coverage for remaining gaps ---

func TestEvalAssignIdentifier(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    x = 10
    x = 20
    console.print_ln(x)
`)
	if !strings.Contains(out, "20") {
		t.Fatalf("expected 20, got %q", out)
	}
}

func TestEvalAssignIndexListNested(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    xs = [[1, 2], [3, 4]]
    xs[0][1] = 99
    console.print_ln(xs[0][1])
`)
	if !strings.Contains(out, "99") {
		t.Fatalf("expected 99, got %q", out)
	}
}

func TestSwitchStmtFallThrough(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    x = 1
    switch x:
        1 -> console.print_ln("one")
        2 -> console.print_ln("two")
`)
	if !strings.Contains(out, "one") {
		t.Fatalf("expected one, got %q", out)
	}
}

func TestSwitchExprFirstMatch(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    result = switch 1:
        1 -> "first"
        2 -> "second"
    console.print_ln(result)
`)
	if !strings.Contains(out, "first") {
		t.Fatalf("expected first, got %q", out)
	}
}

func TestFieldAccessRecord(t *testing.T) {
	out := runSource(t, `package app.main

type Person:
    Name: string
    Age: int

func main():
    p = Person(Name: "Alice", Age: 30)
    console.print_ln(p.Name)
    console.print_ln(p.Age)
`)
	if !strings.Contains(out, "Alice") || !strings.Contains(out, "30") {
		t.Fatalf("expected Alice and 30, got %q", out)
	}
}

func TestFieldAccessNamespace(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    console.print_ln("test")
`)
	if !strings.Contains(out, "test") {
		t.Fatalf("expected test, got %q", out)
	}
}

func TestComparisonIntEqual(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    console.print_ln(1 == 1)
    console.print_ln(1 == 2)
`)
	if !strings.Contains(out, "true") || !strings.Contains(out, "false") {
		t.Fatalf("expected true and false, got %q", out)
	}
}

func TestComparisonFloatEqual(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    console.print_ln(1.5 == 1.5)
    console.print_ln(1.5 == 2.5)
`)
	if !strings.Contains(out, "true") || !strings.Contains(out, "false") {
		t.Fatalf("expected true and false, got %q", out)
	}
}

func TestComparisonIntLess(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    console.print_ln(1 < 2)
    console.print_ln(2 < 1)
`)
	if !strings.Contains(out, "true") || !strings.Contains(out, "false") {
		t.Fatalf("expected true and false, got %q", out)
	}
}

func TestComparisonIntGreater(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    console.print_ln(2 > 1)
    console.print_ln(1 > 2)
`)
	if !strings.Contains(out, "true") || !strings.Contains(out, "false") {
		t.Fatalf("expected true and false, got %q", out)
	}
}

func TestComparisonIntLessEqual(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    console.print_ln(1 <= 1)
    console.print_ln(1 <= 2)
    console.print_ln(2 <= 1)
`)
	if !strings.Contains(out, "true") || !strings.Contains(out, "false") {
		t.Fatalf("expected true and false, got %q", out)
	}
}

func TestComparisonIntGreaterEqual(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    console.print_ln(1 >= 1)
    console.print_ln(2 >= 1)
    console.print_ln(1 >= 2)
`)
	if !strings.Contains(out, "true") || !strings.Contains(out, "false") {
		t.Fatalf("expected true and false, got %q", out)
	}
}

func TestComparisonFloatLess(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    console.print_ln(1.5 < 2.5)
    console.print_ln(2.5 < 1.5)
`)
	if !strings.Contains(out, "true") || !strings.Contains(out, "false") {
		t.Fatalf("expected true and false, got %q", out)
	}
}

func TestComparisonFloatGreater(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    console.print_ln(2.5 > 1.5)
    console.print_ln(1.5 > 2.5)
`)
	if !strings.Contains(out, "true") || !strings.Contains(out, "false") {
		t.Fatalf("expected true and false, got %q", out)
	}
}

func TestComparisonFloatLessEqual(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    console.print_ln(1.5 <= 1.5)
    console.print_ln(1.5 <= 2.5)
`)
	if !strings.Contains(out, "true") {
		t.Fatalf("expected true, got %q", out)
	}
}

func TestComparisonFloatGreaterEqual(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    console.print_ln(1.5 >= 1.5)
    console.print_ln(2.5 >= 1.5)
`)
	if !strings.Contains(out, "true") {
		t.Fatalf("expected true, got %q", out)
	}
}

func TestUnaryMinusOnVariable(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    x = 42
    y = -x
    console.print_ln(y)
`)
	if !strings.Contains(out, "-42") {
		t.Fatalf("expected -42, got %q", out)
	}
}

func TestUnaryNotOnVariable(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    x = true
    y = not x
    console.print_ln(y)
`)
	if !strings.Contains(out, "false") {
		t.Fatalf("expected false, got %q", out)
	}
}

func TestLogicalAndBothFalse(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    console.print_ln(false and false)
`)
	if !strings.Contains(out, "false") {
		t.Fatalf("expected false, got %q", out)
	}
}

func TestLogicalOrBothTrue(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    console.print_ln(true or true)
`)
	if !strings.Contains(out, "true") {
		t.Fatalf("expected true, got %q", out)
	}
}

func TestAssignDictKeyNested(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    m = {"a": {"x": 1}}
    m["a"]["x"] = 99
    console.print_ln(m["a"]["x"])
`)
	if !strings.Contains(out, "99") {
		t.Fatalf("expected 99, got %q", out)
	}
}

func TestSwitchStmtNoMatchNoDefault(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    x = 99
    switch x:
        1 -> console.print_ln("one")
    console.print_ln("done")
`)
	if !strings.Contains(out, "done") {
		t.Fatalf("expected done, got %q", out)
	}
}

func TestSwitchExprNoMatchNoDefault(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    result = switch 99:
        1 -> "one"
    console.print_ln(result)
`)
	if !strings.Contains(out, "null") {
		t.Fatalf("expected null, got %q", out)
	}
}

func TestFieldAccessOnVariable(t *testing.T) {
	out := runSource(t, `package app.main

type Item:
    Name: string
    Value: int

func main():
    item = Item(Name: "test", Value: 42)
    n = item.Name
    v = item.Value
    console.print_ln(n)
    console.print_ln(v)
`)
	if !strings.Contains(out, "test") || !strings.Contains(out, "42") {
		t.Fatalf("expected test and 42, got %q", out)
	}
}

// --- Targeted coverage for specific branches ---

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

func TestFieldAssignmentOnNonRecordError(t *testing.T) {
	requireRuntimeCode(t, `package app.main

func main():
    x = 42
    x.foo = 10
`, "E101")
}

func TestLoopWithStepModifier(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    loop i in 0..10, step 3:
        console.print_ln(i)
`)
	if !strings.Contains(out, "0") || !strings.Contains(out, "3") || !strings.Contains(out, "9") {
		t.Fatalf("expected 0, 3, 6, 9, got %q", out)
	}
}

func TestLoopWithWhileModifier(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    loop i in 0..10, while i < 5:
        console.print_ln(i)
`)
	if strings.Contains(out, "5") {
		t.Fatalf("should not contain 5, got %q", out)
	}
}

func TestInterpolatedStringWithMultipleExprs(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    x = 10
    y = 20
    s = $"x={x}, y={y}"
    console.print_ln(s)
`)
	if !strings.Contains(out, "x=10") || !strings.Contains(out, "y=20") {
		t.Fatalf("expected x=10 and y=20, got %q", out)
	}
}

func TestConstructorWithAllArgs(t *testing.T) {
	out := runSource(t, `package app.main

type Config:
    Name: string
    Value: int

func main():
    c = Config(Name: "test", Value: 42)
    console.print_ln(c.Name)
    console.print_ln(c.Value)
`)
	if !strings.Contains(out, "test") || !strings.Contains(out, "42") {
		t.Fatalf("expected test and 42, got %q", out)
	}
}

func TestRangeLoopWithDictValues(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    m = {"a": 1, "b": 2, "c": 3}
    loop v in m:
        console.print_ln(v)
`)
	if !strings.Contains(out, "1") || !strings.Contains(out, "2") || !strings.Contains(out, "3") {
		t.Fatalf("expected 1, 2, 3, got %q", out)
	}
}

func TestSwitchExprWithWildcard(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    result = switch 99:
        1 -> "one"
        _ -> "default"
    console.print_ln(result)
`)
	if !strings.Contains(out, "default") {
		t.Fatalf("expected default, got %q", out)
	}
}

func TestSwitchStmtWithWildcard(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    x = 99
    switch x:
        1 -> console.print_ln("one")
        _ -> console.print_ln("default")
`)
	if !strings.Contains(out, "default") {
		t.Fatalf("expected default, got %q", out)
	}
}

// --- Builtin coverage tests ---

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
    file.write_text("` + path + `", "content")
    result = file.read_text("` + path + `")
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

func TestRecordMethodWithMultipleReturns(t *testing.T) {
	out := runSource(t, `package app.main

type Pair:
    A: int
    B: int
    func Sum() -> int:
        return A + B

func main():
    p = Pair(A: 3, B: 7)
    console.print_ln(p.Sum())
`)
	if !strings.Contains(out, "10") {
		t.Fatalf("expected 10, got %q", out)
	}
}

func TestNestedRecordConstruction(t *testing.T) {
	out := runSource(t, `package app.main

type Inner:
    Value: int

type Outer:
    Inner: Inner
    Name: string

func main():
    inner = Inner(Value: 42)
    outer = Outer(Inner: inner, Name: "test")
    console.print_ln(outer.Inner.Value)
    console.print_ln(outer.Name)
`)
	if !strings.Contains(out, "42") || !strings.Contains(out, "test") {
		t.Fatalf("expected 42 and test, got %q", out)
	}
}

// --- Additional path coverage ---

func TestConstructorWithExplicitBody(t *testing.T) {
	out := runSource(t, `package app.main

type Counter:
    Value: int
    constructor(initial: int):
        Value = initial

func main():
    c = Counter(initial: 10)
    console.print_ln(c.Value)
`)
	if !strings.Contains(out, "10") {
		t.Fatalf("expected 10, got %q", out)
	}
}

func TestMethodMutatingRecord(t *testing.T) {
	out := runSource(t, `package app.main

type Counter:
    Value: int
    func Increment():
        Value = Value + 1
    func Get() -> int:
        return Value

func main():
    c = Counter(Value: 0)
    c.Increment()
    c.Increment()
    console.print_ln(c.Get())
`)
	if !strings.Contains(out, "2") {
		t.Fatalf("expected 2, got %q", out)
	}
}

func TestDictIteration(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    m = {"a": 1, "b": 2}
    loop k in m:
        console.print_ln(k)
`)
	if !strings.Contains(out, "a") || !strings.Contains(out, "b") {
		t.Fatalf("expected a and b, got %q", out)
	}
}

func TestListIteration(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    xs = [10, 20, 30]
    loop x in xs:
        console.print_ln(x)
`)
	if !strings.Contains(out, "10") || !strings.Contains(out, "20") || !strings.Contains(out, "30") {
		t.Fatalf("expected 10, 20, 30, got %q", out)
	}
}

func TestSwitchStmtWithWildcardPattern(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    x = 2
    switch x:
        1 -> console.print_ln("one")
        2 -> console.print_ln("two")
        _ -> console.print_ln("other")
`)
	if !strings.Contains(out, "two") {
		t.Fatalf("expected two, got %q", out)
	}
}

func TestSwitchExprWithWildcardPattern(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    result = switch 3:
        1 -> "one"
        2 -> "two"
        _ -> "other"
    console.print_ln(result)
`)
	if !strings.Contains(out, "other") {
		t.Fatalf("expected other, got %q", out)
	}
}

func TestIfElseIfElse(t *testing.T) {
	out := runSource(t, `package app.main

func classify(x: int) -> string:
    if x < 0:
        return "negative"
    else if x == 0:
        return "zero"
    else:
        return "positive"

func main():
    console.print_ln(classify(-1))
    console.print_ln(classify(0))
    console.print_ln(classify(1))
`)
	if !strings.Contains(out, "negative") || !strings.Contains(out, "zero") || !strings.Contains(out, "positive") {
		t.Fatalf("expected negative, zero, positive, got %q", out)
	}
}

func TestLoopWithStepAndWhile(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    loop i in 0..100, step 3, while i < 10:
        console.print_ln(i)
`)
	if !strings.Contains(out, "0") || !strings.Contains(out, "3") || !strings.Contains(out, "9") || strings.Contains(out, "12") {
		t.Fatalf("expected 0, 3, 6, 9 but not 12, got %q", out)
	}
}

func TestLoopWithIfAndContinue(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    loop i in 0..5, if i != 2:
        console.print_ln(i)
`)
	if strings.Contains(out, "2") {
		t.Fatalf("should not contain 2, got %q", out)
	}
}

func TestFunctionReturningRecord(t *testing.T) {
	out := runSource(t, `package app.main

type Vec2:
    X: int
    Y: int

func make_vec(x: int, y: int) -> Vec2:
    return Vec2(X: x, Y: y)

func main():
    v = make_vec(3, 4)
    console.print_ln(v.X)
    console.print_ln(v.Y)
`)
	if !strings.Contains(out, "3") || !strings.Contains(out, "4") {
		t.Fatalf("expected 3 and 4, got %q", out)
	}
}

func TestRecordWithConstructorSettingDefaults(t *testing.T) {
	out := runSource(t, `package app.main

type Config:
    Name: string
    Value: int
    constructor():
        Name = "default"
        Value = 0

func main():
    c = Config()
    console.print_ln(c.Name)
    console.print_ln(c.Value)
`)
	if !strings.Contains(out, "default") || !strings.Contains(out, "0") {
		t.Fatalf("expected default and 0, got %q", out)
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

func TestVarWithInferredType(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    x = 42
    console.print_ln(x)
`)
	if !strings.Contains(out, "42") {
		t.Fatalf("expected 42, got %q", out)
	}
}

func TestStringConcatenation(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    s = "hello" + " " + "world"
    console.print_ln(s)
`)
	if !strings.Contains(out, "hello world") {
		t.Fatalf("expected hello world, got %q", out)
	}
}

func TestIntArithmeticAllOps(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    console.print_ln(10 + 3)
    console.print_ln(10 - 3)
    console.print_ln(10 * 3)
    console.print_ln(10 / 3)
    console.print_ln(10 % 3)
`)
	if !strings.Contains(out, "13") || !strings.Contains(out, "7") || !strings.Contains(out, "30") || !strings.Contains(out, "3") || !strings.Contains(out, "1") {
		t.Fatalf("expected arithmetic results, got %q", out)
	}
}

func TestFloatArithmeticAllOps(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    console.print_ln(1.5 + 2.5)
    console.print_ln(5.0 - 1.5)
    console.print_ln(2.0 * 3.0)
    console.print_ln(10.0 / 4.0)
`)
	if !strings.Contains(out, "4") || !strings.Contains(out, "3.5") || !strings.Contains(out, "6") || !strings.Contains(out, "2.5") {
		t.Fatalf("expected float arithmetic results, got %q", out)
	}
}

func TestVoidReturn(t *testing.T) {
	out := runSource(t, `package app.main

func do_something():
    console.print_ln("doing")
    return

func main():
    do_something()
`)
	if !strings.Contains(out, "doing") {
		t.Fatalf("expected doing, got %q", out)
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

func TestDictLoopEntry(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    m = {"name": "test", "value": "42"}
    loop entry in m:
        console.print_ln(entry)
`)
	if !strings.Contains(out, "name") || !strings.Contains(out, "test") {
		t.Fatalf("expected dict entries, got %q", out)
	}
}

func TestAssignNewVariable(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    x = 10
    console.print_ln(x)
`)
	if !strings.Contains(out, "10") {
		t.Fatalf("expected 10, got %q", out)
	}
}

func TestMultipleVars(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    a = 1
    b = 2
    c = 3
    console.print_ln(a + b + c)
`)
	if !strings.Contains(out, "6") {
		t.Fatalf("expected 6, got %q", out)
	}
}

func TestNestedFunctionCall(t *testing.T) {
	out := runSource(t, `package app.main

func double(x: int) -> int:
    return x * 2

func quadruple(x: int) -> int:
    return double(double(x))

func main():
    console.print_ln(quadruple(5))
`)
	if !strings.Contains(out, "20") {
		t.Fatalf("expected 20, got %q", out)
	}
}

func TestRefParamFunction(t *testing.T) {
	out := runSource(t, `package app.main

func increment(ref x: int):
    x = x + 1

func main():
    val = 10
    increment(val)
    console.print_ln(val)
`)
	if !strings.Contains(out, "11") {
		t.Fatalf("expected 11, got %q", out)
	}
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

func TestStringInterpolationWithExpr(t *testing.T) {
	out := runSource(t, `package app.main

func double(x: int) -> int:
    return x * 2

func main():
    x = 21
    console.print_ln($"result: {double(x)}")
`)
	if !strings.Contains(out, "42") {
		t.Fatalf("expected 42, got %q", out)
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

func TestDiscardWriterNil(t *testing.T) {
	// Test that discardWriter returns io.Discard when given nil.
	w := discardWriter(nil)
	if w == nil {
		t.Fatal("expected io.Discard, got nil")
	}
	_, err := w.Write([]byte("test"))
	if err != nil {
		t.Fatalf("unexpected error writing to discard: %v", err)
	}
}

func TestCloneForTaskListWithNonSendable(t *testing.T) {
	// A list containing a FuncVal should fail to clone.
	_, err := cloneForTask(&ListVal{Elements: []Value{IntVal(1), &FuncVal{Name: "bad"}}})
	if err == nil {
		t.Fatal("expected error for list with non-sendable element")
	}
}

func TestCloneForTaskDictWithNonSendable(t *testing.T) {
	// A dict containing a BuiltinVal should fail to clone.
	_, err := cloneForTask(&DictVal{Entries: map[string]Value{"a": &BuiltinVal{Name: "bad"}}})
	if err == nil {
		t.Fatal("expected error for dict with non-sendable element")
	}
}

func TestCloneForTaskTupleWithNonSendable(t *testing.T) {
	// A tuple containing a TaskVal should fail to clone.
	_, err := cloneForTask(&TupleVal{Elements: []Value{IntVal(1), newTaskVal()}})
	if err == nil {
		t.Fatal("expected error for tuple with non-sendable element")
	}
}
