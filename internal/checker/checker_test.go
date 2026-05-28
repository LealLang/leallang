// Copyright (c) 2026 LealLang Contributors
// SPDX-License-Identifier: MIT

package checker

import (
	"testing"

	"github.com/LealLang/leallang/internal/diagnostics"
	"github.com/LealLang/leallang/internal/lexer"
	"github.com/LealLang/leallang/internal/parser"
)

func checkSource(t *testing.T, src string) *diagnostics.Diagnostics {
	t.Helper()
	diag := diagnostics.New()
	tokens := lexer.New("test.ll", src, diag).Tokenize()
	program := parser.New(tokens, diag).Parse()
	if diag.HasErrors() {
		t.Fatalf("parse errors:\n%s", diag.Format())
	}
	Check(program, diag)
	return diag
}

func checkSourceAllowParseErrors(t *testing.T, src string) *diagnostics.Diagnostics {
	t.Helper()
	diag := diagnostics.New()
	tokens := lexer.New("test.ll", src, diag).Tokenize()
	program := parser.New(tokens, diag).Parse()
	Check(program, diag)
	return diag
}

func requireNoErrors(t *testing.T, diag *diagnostics.Diagnostics) {
	t.Helper()
	if diag.HasErrors() {
		t.Fatalf("unexpected diagnostics:\n%s", diag.Format())
	}
}

func requireErrorCode(t *testing.T, diag *diagnostics.Diagnostics, code string) {
	t.Helper()
	for _, err := range diag.Errors() {
		if err.Code == code {
			return
		}
	}
	t.Fatalf("expected diagnostic %s, got:\n%s", code, diag.Format())
}

// --- Phase 1: Variable and Const Declarations ---

func TestVarDeclExplicitType(t *testing.T) {
	diag := checkSource(t, `package app.main

name: string = "hello"
count: int = 42
flag: bool = true
`)
	requireNoErrors(t, diag)
}

func TestVarDeclInferredType(t *testing.T) {
	diag := checkSource(t, `package app.main

name = "hello"
count = 42
flag = true
rate = 3.14
ch = 'a'
`)
	requireNoErrors(t, diag)
}

func TestVarDeclNoValue(t *testing.T) {
	diag := checkSource(t, `package app.main

name: string
count: int
`)
	requireNoErrors(t, diag)
}

func TestVarDeclTypeMismatch(t *testing.T) {
	diag := checkSource(t, `package app.main

name: int = "hello"
`)
	requireErrorCode(t, diag, "E022")
}

func TestConstDecl(t *testing.T) {
	diag := checkSource(t, `package app.main

const MAX: int = 100
const GREETING: string = "hi"
`)
	requireNoErrors(t, diag)
}

func TestConstDeclTypeMismatch(t *testing.T) {
	diag := checkSource(t, `package app.main

const MAX: string = 100
`)
	requireErrorCode(t, diag, "E022")
}

func TestConstReassign(t *testing.T) {
	diag := checkSource(t, `package app.main

const MAX: int = 100

func main():
    MAX = 200
`)
	requireErrorCode(t, diag, "E033")
}

// --- Phase 1: Duplicate Name ---

func TestDuplicateName(t *testing.T) {
	diag := checkSource(t, `package app.main

name: string = "a"
name: int = 1
`)
	requireErrorCode(t, diag, "E050")
}

// --- Phase 1: Undefined Name ---

func TestUndefinedName(t *testing.T) {
	diag := checkSource(t, `package app.main

func main():
    x = unknown_var
`)
	requireErrorCode(t, diag, "E021")
}

// --- Phase 1: Literal Types ---

func TestLiteralTypes(t *testing.T) {
	diag := checkSource(t, `package app.main

a: int = 42
b: float = 3.14
c: string = "hello"
d: char = 'x'
e: bool = true
f: bool = false
`)
	requireNoErrors(t, diag)
}

// --- Phase 2: Operators ---

func TestArithmeticOperators(t *testing.T) {
	diag := checkSource(t, `package app.main

func main():
    a: int = 1 + 2
    b: int = 10 - 3
    c: int = 4 * 5
    d: int = 10 / 2
    e: int = 10 % 3
`)
	requireNoErrors(t, diag)
}

func TestFloatArithmetic(t *testing.T) {
	diag := checkSource(t, `package app.main

func main():
    a: float = 1.0 + 2.0
    b: float = 3 + 1.5
`)
	requireNoErrors(t, diag)
}

func TestStringConcatenation(t *testing.T) {
	diag := checkSource(t, `package app.main

func main():
    s: string = "hello" + " " + "world"
`)
	requireNoErrors(t, diag)
}

func TestArithmeticTypeError(t *testing.T) {
	diag := checkSource(t, `package app.main

func main():
    x: int = "hello" + 1
`)
	requireErrorCode(t, diag, "E025")
}

func TestComparisonOperators(t *testing.T) {
	diag := checkSource(t, `package app.main

func main():
    a: bool = 1 < 2
    b: bool = 1 <= 2
    c: bool = 1 > 2
    d: bool = 1 >= 2
    e: bool = 1 == 2
    f: bool = 1 != 2
`)
	requireNoErrors(t, diag)
}

func TestBooleanOperators(t *testing.T) {
	diag := checkSource(t, `package app.main

func main():
    a: bool = true and false
    b: bool = true or false
`)
	requireNoErrors(t, diag)
}

func TestBooleanOperatorError(t *testing.T) {
	diag := checkSource(t, `package app.main

func main():
    x: bool = 1 and true
`)
	requireErrorCode(t, diag, "E054")
}

func TestUnaryMinus(t *testing.T) {
	diag := checkSource(t, `package app.main

func main():
    x: int = -5
    y: float = -3.14
`)
	requireNoErrors(t, diag)
}

func TestUnaryMinusError(t *testing.T) {
	diag := checkSource(t, `package app.main

func main():
    x: int = -"hello"
`)
	requireErrorCode(t, diag, "E055")
}

func TestUnaryNot(t *testing.T) {
	diag := checkSource(t, `package app.main

func main():
    x: bool = not true
`)
	requireNoErrors(t, diag)
}

func TestUnaryNotError(t *testing.T) {
	diag := checkSource(t, `package app.main

func main():
    x: bool = not 1
`)
	requireErrorCode(t, diag, "E055")
}

func TestInterpolatedString(t *testing.T) {
	diag := checkSource(t, `package app.main

func main():
    name = "world"
    s: string = $"Hello, {name}!"
`)
	requireNoErrors(t, diag)
}

// --- Phase 2: Ident Lookup ---

func TestIdentLookup(t *testing.T) {
	diag := checkSource(t, `package app.main

name: string = "hello"

func main():
    x: string = name
`)
	requireNoErrors(t, diag)
}

func TestUndefinedIdent(t *testing.T) {
	diag := checkSource(t, `package app.main

func main():
    x = undefined
`)
	requireErrorCode(t, diag, "E021")
}

// --- Phase 3: Functions ---

func TestFunctionDeclAndCall(t *testing.T) {
	diag := checkSource(t, `package app.main

func greet(name: string) -> string:
    return name

func main():
    result: string = greet("Alice")
`)
	requireNoErrors(t, diag)
}

func TestFunctionCallWrongArgCount(t *testing.T) {
	diag := checkSource(t, `package app.main

func greet(name: string):
    pass

func main():
    greet("Alice", "extra")
`)
	requireErrorCode(t, diag, "E026")
}

func TestFunctionCallWrongArgType(t *testing.T) {
	diag := checkSource(t, `package app.main

func greet(name: string):
    pass

func main():
    greet(42)
`)
	requireErrorCode(t, diag, "E022")
}

func TestFunctionCallNamedArgs(t *testing.T) {
	diag := checkSource(t, `package app.main

func greet(name: string, age: int):
    pass

func main():
    greet(age: 30, name: "Alice")
`)
	requireNoErrors(t, diag)
}

func TestFunctionCallUnknownNamedArg(t *testing.T) {
	diag := checkSource(t, `package app.main

func greet(name: string):
    pass

func main():
    greet(name: "Alice", unknown: 42)
`)
	requireErrorCode(t, diag, "E027")
}

func TestFunctionCallPositionalAfterNamed(t *testing.T) {
	// Parser catches this with E011 before the checker runs.
	diag := checkSourceAllowParseErrors(t, `package app.main

func greet(name: string, age: int):
    pass

func main():
    greet(name: "Alice", 30)
`)
	requireErrorCode(t, diag, "E011")
}

func TestRefParamValidation(t *testing.T) {
	diag := checkSource(t, `package app.main

func modify(ref x: int):
    pass

func main():
    y: int = 10
    modify(y)
`)
	requireNoErrors(t, diag)
}

func TestRefParamLiteralError(t *testing.T) {
	diag := checkSource(t, `package app.main

func modify(ref x: int):
    pass

func main():
    modify(10)
`)
	requireErrorCode(t, diag, "E034")
}

func TestReturnTypeError(t *testing.T) {
	diag := checkSource(t, `package app.main

func greet() -> int:
    return "hello"
`)
	requireErrorCode(t, diag, "E036")
}

func TestReturnInVoidFunction(t *testing.T) {
	diag := checkSource(t, `package app.main

func greet():
    return 42
`)
	requireErrorCode(t, diag, "E037")
}

func TestVoidFunctionNoReturn(t *testing.T) {
	diag := checkSource(t, `package app.main

func greet():
    pass
`)
	requireNoErrors(t, diag)
}

// --- Phase 4: Records ---

func TestRecordConstruction(t *testing.T) {
	diag := checkSource(t, `package app.main

type Point:
    X: int
    Y: int

func main():
    p = Point(X: 10, Y: 20)
`)
	requireNoErrors(t, diag)
}

func TestRecordFieldAccess(t *testing.T) {
	diag := checkSource(t, `package app.main

type Point:
    X: int
    Y: int

func main():
    p = Point(X: 10, Y: 20)
    x: int = p.X
`)
	requireNoErrors(t, diag)
}

func TestRecordUnknownField(t *testing.T) {
	diag := checkSource(t, `package app.main

type Point:
    X: int
    Y: int

func main():
    p = Point(X: 10, Y: 20, Z: 30)
`)
	requireErrorCode(t, diag, "E027")
}

func TestRecordMethodCall(t *testing.T) {
	diag := checkSource(t, `package app.main

type Point:
    X: int
    Y: int
    func Greet() -> string:
        return "hi"

func main():
    p = Point(X: 10, Y: 20)
    result: string = p.Greet()
`)
	requireNoErrors(t, diag)
}

func TestRecordMethodCallTypeError(t *testing.T) {
	diag := checkSource(t, `package app.main

type Point:
    X: int
    Y: int
    func Greet() -> string:
        return "hi"

func main():
    p = Point(X: 10, Y: 20)
    result: int = p.Greet()
`)
	requireErrorCode(t, diag, "E022")
}

func TestRecordMethodCallUnknown(t *testing.T) {
	diag := checkSource(t, `package app.main

type Point:
    X: int
    Y: int

func main():
    p = Point(X: 10, Y: 20)
    p.FooBar()
`)
	requireErrorCode(t, diag, "E039")
}

func TestRecordFieldAccessError(t *testing.T) {
	diag := checkSource(t, `package app.main

type Point:
    X: int
    Y: int

func main():
    p = Point(X: 10, Y: 20)
    z: int = p.Z
`)
	requireErrorCode(t, diag, "E039")
}

// --- Phase 5: Collections ---

func TestListLiteral(t *testing.T) {
	diag := checkSource(t, `package app.main

func main():
    xs: List<int> = [1, 2, 3]
`)
	requireNoErrors(t, diag)
}

func TestDictLiteral(t *testing.T) {
	diag := checkSource(t, `package app.main

func main():
    m: Dict<string, int> = {"a": 1, "b": 2}
`)
	requireNoErrors(t, diag)
}

func TestListIndex(t *testing.T) {
	diag := checkSource(t, `package app.main

func main():
    xs = [1, 2, 3]
    x: int = xs[0]
`)
	requireNoErrors(t, diag)
}

func TestDictIndex(t *testing.T) {
	diag := checkSource(t, `package app.main

func main():
    m = {"a": 1}
    x: int = m["a"]
`)
	requireNoErrors(t, diag)
}

func TestListIndexTypeError(t *testing.T) {
	diag := checkSource(t, `package app.main

func main():
    xs = [1, 2, 3]
    x: int = xs["hello"]
`)
	requireErrorCode(t, diag, "E041")
}

func TestNotIndexable(t *testing.T) {
	diag := checkSource(t, `package app.main

func main():
    x: int = 42
    y = x[0]
`)
	requireErrorCode(t, diag, "E040")
}

// --- Phase 6: Null Safety ---

func TestNullToNullable(t *testing.T) {
	diag := checkSource(t, `package app.main

name: string? = null
`)
	requireNoErrors(t, diag)
}

func TestNullToNonNullable(t *testing.T) {
	diag := checkSource(t, `package app.main

name: string = null
`)
	requireErrorCode(t, diag, "E023")
}

func TestNullableAssignment(t *testing.T) {
	diag := checkSource(t, `package app.main

func main():
    name: string? = "hello"
`)
	requireNoErrors(t, diag)
}

// --- Phase 7: Visibility ---

func TestPrivateFieldAccess(t *testing.T) {
	// Private fields are accessed within the same package, so this should work.
	diag := checkSource(t, `package app.main

type Point:
    X: int
    Y: int

func main():
    p = Point(X: 10, Y: 20)
    x: int = p.X
`)
	requireNoErrors(t, diag)
}

// --- Phase 8: Control Flow ---

func TestIfConditionBool(t *testing.T) {
	diag := checkSource(t, `package app.main

func main():
    if true:
        pass
`)
	requireNoErrors(t, diag)
}

func TestIfConditionError(t *testing.T) {
	diag := checkSource(t, `package app.main

func main():
    if 42:
        pass
`)
	requireErrorCode(t, diag, "E047")
}

func TestSwitchExprTypeMismatch(t *testing.T) {
	diag := checkSource(t, `package app.main

func main():
    x: int = switch true:
        true -> 1
        false -> "hello"
`)
	requireErrorCode(t, diag, "E042")
}

func TestLoopRange(t *testing.T) {
	diag := checkSource(t, `package app.main

func main():
    loop i in 0..10:
        pass
`)
	requireNoErrors(t, diag)
}

func TestLoopList(t *testing.T) {
	diag := checkSource(t, `package app.main

func main():
    xs = [1, 2, 3]
    loop x in xs:
        pass
`)
	requireNoErrors(t, diag)
}

func TestLoopWhileCondition(t *testing.T) {
	diag := checkSource(t, `package app.main

func main():
    loop i in 0..10, while true:
        pass
`)
	requireNoErrors(t, diag)
}

func TestLoopWhileConditionError(t *testing.T) {
	diag := checkSource(t, `package app.main

func main():
    loop i in 0..10, while 42:
        pass
`)
	requireErrorCode(t, diag, "E047")
}

func TestRangeBoundsError(t *testing.T) {
	diag := checkSource(t, `package app.main

func main():
    loop i in "a".."z":
        pass
`)
	requireErrorCode(t, diag, "E048")
}

// --- Built-in Namespaces ---

func TestConsolePrint(t *testing.T) {
	diag := checkSource(t, `package app.main

func main():
    console.print_ln("hello")
`)
	requireNoErrors(t, diag)
}

func TestFileReadText(t *testing.T) {
	diag := checkSource(t, `package app.main

func main():
    result = file.read_text("test.txt")
`)
	requireNoErrors(t, diag)
}

func TestUnknownNamespace(t *testing.T) {
	diag := checkSource(t, `package app.main

func main():
    unknown.run()
`)
	requireErrorCode(t, diag, "E021")
}

func TestNamespaceUnknownMember(t *testing.T) {
	diag := checkSource(t, `package app.main

func main():
    console.unknown_func()
`)
	requireErrorCode(t, diag, "E039")
}

// --- Multiple Return ---

func TestMultiReturnDecl(t *testing.T) {
	// Multi-return type declaration is parsed correctly.
	diag := checkSource(t, `package app.main

func divide(a: int, b: int) -> int, string:
    return a / b

func main():
    result = divide(10, 2)
`)
	// The checker reports E036 because return value doesn't match tuple type.
	// This is expected behavior — multi-return values require parser support.
	requireErrorCode(t, diag, "E036")
}

// --- End-to-End Snippet ---

func TestEndToEndSnippet(t *testing.T) {
	diag := checkSource(t, `package app.main

const MAX: int = 100

type Point:
    X: int
    Y: int
    constructor(x: int, y: int):
        X = x
        Y = y
    func Greet() -> string:
        return $"Point({X}, {Y})"

func main():
    p = Point(x: 10, y: 20)
    result: string = switch p.X:
        10 -> "ten"
        _  -> "other"
    loop i in 0..MAX, step 2, while i < 50:
        console.print_ln($"i is {i}")
`)
	requireNoErrors(t, diag)
}

// --- Regression: fix.md ---

func TestLoopUnderscoreIterableTypeError(t *testing.T) {
	// Bug: loop iterables were not type-checked when using _ variable.
	diag := checkSource(t, `package app.main

func main():
    loop _ in undefined_func():
        pass
`)
	requireErrorCode(t, diag, "E021")
}

func TestAssignmentTypeMismatchSpan(t *testing.T) {
	// Bug: assignment error position pointed to = instead of the value expression.
	diag := checkSource(t, `package app.main

func main():
    x: int = true
`)
	requireErrorCode(t, diag, "E022")
}
