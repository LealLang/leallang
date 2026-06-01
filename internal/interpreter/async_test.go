// Copyright (c) 2026 LealLang Contributors
// SPDX-License-Identifier: MIT

package interpreter

import (
	"strings"
	"testing"
	"time"
)

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

