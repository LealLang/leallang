// Copyright (c) 2026 LealLang Contributors
// SPDX-License-Identifier: MIT

package interpreter

import (
	"os"
	"strings"
	"testing"
)

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

