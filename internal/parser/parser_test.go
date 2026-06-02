// Copyright (c) 2026 LealLang Contributors
// SPDX-License-Identifier: MIT

package parser

import (
	"bytes"
	"strings"
	"testing"

	"github.com/LealLang/leallang/internal/ast"
	"github.com/LealLang/leallang/internal/diagnostics"
	"github.com/LealLang/leallang/internal/lexer"
	"github.com/LealLang/leallang/internal/token"
)

func parseSource(t *testing.T, src string) (*ast.Program, *diagnostics.Diagnostics) {
	t.Helper()
	diag := diagnostics.New()
	tokens := lexer.New("test.ll", src, diag).Tokenize()
	program := New(tokens, diag).Parse()
	return program, diag
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

func onlyFunc(t *testing.T, src string) *ast.FuncDecl {
	t.Helper()
	program, diag := parseSource(t, src)
	requireNoErrors(t, diag)
	if len(program.Decls) != 1 {
		t.Fatalf("decl count = %d, want 1", len(program.Decls))
	}
	fn, ok := program.Decls[0].(*ast.FuncDecl)
	if !ok {
		t.Fatalf("decl type = %T, want *ast.FuncDecl", program.Decls[0])
	}
	return fn
}

func onlyStmt(t *testing.T, src string) ast.Stmt {
	t.Helper()
	fn := onlyFunc(t, "package app.main\n\nfunc main():\n"+indent(src))
	if len(fn.Body) != 1 {
		t.Fatalf("body len = %d, want 1", len(fn.Body))
	}
	return fn.Body[0]
}

func indent(src string) string {
	lines := strings.Split(src, "\n")
	for i, line := range lines {
		if line != "" {
			lines[i] = "    " + line
		}
	}
	return strings.Join(lines, "\n")
}

func TestPackageDeclaration(t *testing.T) {
	program, diag := parseSource(t, "package app\n")
	requireNoErrors(t, diag)
	if got := strings.Join(program.Package.Path, "."); got != "app" {
		t.Fatalf("package path = %q", got)
	}

	program, diag = parseSource(t, "package app.main\n")
	requireNoErrors(t, diag)
	if got := strings.Join(program.Package.Path, "."); got != "app.main" {
		t.Fatalf("package path = %q", got)
	}
}

func TestImports(t *testing.T) {
	program, diag := parseSource(t, `package app.main

import app.settings
import app.util as util
`)
	requireNoErrors(t, diag)
	if len(program.Imports) != 2 {
		t.Fatalf("imports = %d", len(program.Imports))
	}
	if got := strings.Join(program.Imports[0].Path, "."); got != "app.settings" {
		t.Fatalf("first import = %q", got)
	}
	if program.Imports[1].Alias != "util" {
		t.Fatalf("alias = %q", program.Imports[1].Alias)
	}
}

func TestVariableDeclarations(t *testing.T) {
	program, diag := parseSource(t, `package app.main

name: string = "hello"
count = 3
title: string
`)
	requireNoErrors(t, diag)
	if len(program.Decls) != 3 {
		t.Fatalf("decl count = %d", len(program.Decls))
	}
	explicit := program.Decls[0].(*ast.VarDecl)
	if explicit.Name != "name" || explicit.Type == nil || explicit.Value == nil {
		t.Fatalf("bad explicit var: %#v", explicit)
	}
	inferred := program.Decls[1].(*ast.VarDecl)
	if inferred.Name != "count" || inferred.Type != nil || inferred.Value == nil {
		t.Fatalf("bad inferred var: %#v", inferred)
	}
	noValue := program.Decls[2].(*ast.VarDecl)
	if noValue.Name != "title" || noValue.Type == nil || noValue.Value != nil {
		t.Fatalf("bad no-value var: %#v", noValue)
	}
}

func TestConstDeclaration(t *testing.T) {
	program, diag := parseSource(t, `package app.main

const MAX: int = 100
`)
	requireNoErrors(t, diag)
	decl := program.Decls[0].(*ast.ConstDecl)
	if decl.Name != "MAX" || decl.Type == nil || decl.Value == nil {
		t.Fatalf("bad const decl: %#v", decl)
	}
}

func TestFunctionDeclaration(t *testing.T) {
	program, diag := parseSource(t, `package app.main

func empty():
    pass

pub func greet(ref name: string, age: int) -> string:
    return name
`)
	requireNoErrors(t, diag)
	if len(program.Decls) != 2 {
		t.Fatalf("decl count = %d", len(program.Decls))
	}
	empty := program.Decls[0].(*ast.FuncDecl)
	if empty.Name != "empty" || len(empty.Params) != 0 || len(empty.ReturnTypes) != 0 {
		t.Fatalf("bad empty func: %#v", empty)
	}
	greet := program.Decls[1].(*ast.FuncDecl)
	if !greet.Pub || greet.Name != "greet" || len(greet.Params) != 2 || len(greet.ReturnTypes) == 0 {
		t.Fatalf("bad greet func: %#v", greet)
	}
	if !greet.Params[0].Ref {
		t.Fatalf("first param should be ref")
	}
}

func TestFunctionCallArguments(t *testing.T) {
	stmt := onlyStmt(t, `greet("Alice", age: 30)`)
	call := stmt.(*ast.ExprStmt).X.(*ast.CallExpr)
	if len(call.Args) != 2 || call.Args[0].Name != "" || call.Args[1].Name != "age" {
		t.Fatalf("bad args: %#v", call.Args)
	}

	_, diag := parseSource(t, `package app.main

func main():
    greet(name: "Alice", 30)
`)
	requireErrorCode(t, diag, "E011")
}

func TestBinaryPrecedence(t *testing.T) {
	stmt := onlyStmt(t, `result = a or b and c == d + e * f`)
	assign := stmt.(*ast.AssignStmt)
	orExpr := assign.Value.(*ast.BinaryExpr)
	if orExpr.Op != token.OR {
		t.Fatalf("top op = %s, want or", orExpr.Op)
	}
	andExpr := orExpr.Right.(*ast.BinaryExpr)
	if andExpr.Op != token.AND {
		t.Fatalf("right op = %s, want and", andExpr.Op)
	}
	cmp := andExpr.Right.(*ast.BinaryExpr)
	if cmp.Op != token.EQ_EQ {
		t.Fatalf("compare op = %s", cmp.Op)
	}
	add := cmp.Right.(*ast.BinaryExpr)
	if add.Op != token.PLUS {
		t.Fatalf("add op = %s", add.Op)
	}
	mul := add.Right.(*ast.BinaryExpr)
	if mul.Op != token.STAR {
		t.Fatalf("mul op = %s", mul.Op)
	}
}

func TestUnaryExpressions(t *testing.T) {
	stmt := onlyStmt(t, `result = not ready or -count < 0`)
	assign := stmt.(*ast.AssignStmt)
	orExpr := assign.Value.(*ast.BinaryExpr)
	if _, ok := orExpr.Left.(*ast.UnaryExpr); !ok {
		t.Fatalf("left should be unary not: %T", orExpr.Left)
	}
	cmp := orExpr.Right.(*ast.BinaryExpr)
	if _, ok := cmp.Left.(*ast.UnaryExpr); !ok {
		t.Fatalf("comparison left should be unary minus: %T", cmp.Left)
	}
}

func TestFieldAccessChaining(t *testing.T) {
	stmt := onlyStmt(t, `result = a.b.c`)
	assign := stmt.(*ast.AssignStmt)
	field := assign.Value.(*ast.FieldExpr)
	if field.Field != "c" {
		t.Fatalf("field = %q", field.Field)
	}
	inner := field.X.(*ast.FieldExpr)
	if inner.Field != "b" {
		t.Fatalf("inner field = %q", inner.Field)
	}
}

func TestIndexExpression(t *testing.T) {
	stmt := onlyStmt(t, `result = items[0]`)
	assign := stmt.(*ast.AssignStmt)
	if _, ok := assign.Value.(*ast.IndexExpr); !ok {
		t.Fatalf("value = %T, want IndexExpr", assign.Value)
	}
}

func TestComponentRef(t *testing.T) {
	stmt := onlyStmt(t, `@Button[save_button].Text = "hello"`)
	assign := stmt.(*ast.AssignStmt)
	field := assign.Target.(*ast.FieldExpr)
	ref := field.X.(*ast.ComponentRefExpr)
	if ref.Component != "Button" || ref.ID != "save_button" || field.Field != "Text" {
		t.Fatalf("bad component ref: %#v %#v", ref, field)
	}
}

func TestRangeExpression(t *testing.T) {
	stmt := onlyStmt(t, `a = 0..10`)
	if stmt.(*ast.AssignStmt).Value.(*ast.RangeExpr).Exclusive {
		t.Fatalf("inclusive range marked exclusive")
	}
	stmt = onlyStmt(t, `a = 0..<10`)
	if !stmt.(*ast.AssignStmt).Value.(*ast.RangeExpr).Exclusive {
		t.Fatalf("exclusive range not marked exclusive")
	}
}

func TestListAndDictLiteral(t *testing.T) {
	stmt := onlyStmt(t, `xs = [1, 2, 3]`)
	list := stmt.(*ast.AssignStmt).Value.(*ast.ListLiteral)
	if len(list.Elements) != 3 {
		t.Fatalf("list elements = %d", len(list.Elements))
	}
	stmt = onlyStmt(t, `m = {"key": 1, "other": 2}`)
	dict := stmt.(*ast.AssignStmt).Value.(*ast.DictLiteral)
	if len(dict.Pairs) != 2 {
		t.Fatalf("dict pairs = %d", len(dict.Pairs))
	}
}

func TestInterpolatedString(t *testing.T) {
	stmt := onlyStmt(t, `console.print_ln($"Hello, {name}, age {age}")`)
	call := stmt.(*ast.ExprStmt).X.(*ast.CallExpr)
	interp := call.Args[0].Value.(*ast.InterpStringExpr)
	if len(interp.Segments) != 4 {
		t.Fatalf("segments = %d: %#v", len(interp.Segments), interp.Segments)
	}
	if interp.Segments[1].Expr.(*ast.Ident).Name != "name" {
		t.Fatalf("bad first expr segment")
	}
}

func TestIfElseIfElse(t *testing.T) {
	stmt := onlyStmt(t, `if ready:
    pass
else if waiting:
    pass
else:
    pass`)
	ifStmt := stmt.(*ast.IfStmt)
	if len(ifStmt.ElseIfs) != 1 || len(ifStmt.Else) != 1 {
		t.Fatalf("bad if branches: %#v", ifStmt)
	}
}

func TestStatementSwitch(t *testing.T) {
	stmt := onlyStmt(t, `switch color:
    "blue":
        pass
    _:
        pass`)
	sw := stmt.(*ast.SwitchStmt)
	if len(sw.Cases) != 2 || sw.Cases[1].Patterns != nil {
		t.Fatalf("bad switch cases: %#v", sw.Cases)
	}
}

func TestExpressionSwitch(t *testing.T) {
	stmt := onlyStmt(t, `result = switch color:
    "blue" -> colors.blue
    "red" | "rouge" -> colors.red
    _ -> colors.white`)
	sw := stmt.(*ast.AssignStmt).Value.(*ast.SwitchExpr)
	if len(sw.Arms) != 3 || len(sw.Arms[1].Patterns) != 2 || sw.Arms[2].Patterns != nil {
		t.Fatalf("bad switch arms: %#v", sw.Arms)
	}
}

func TestLoopModifiers(t *testing.T) {
	stmt := onlyStmt(t, `loop i in 0..10, step 2, while i < 8, if i != 4:
    pass`)
	loop := stmt.(*ast.LoopStmt)
	if len(loop.Iterators) != 1 || loop.Step == nil || loop.While == nil || loop.IfCond == nil {
		t.Fatalf("bad loop: %#v", loop)
	}
}

func TestParallelLoop(t *testing.T) {
	stmt := onlyStmt(t, `loop num in list1, char in list2:
    pass`)
	loop := stmt.(*ast.LoopStmt)
	if len(loop.Iterators) != 2 {
		t.Fatalf("iterators = %d", len(loop.Iterators))
	}
}

func TestInfiniteLoop(t *testing.T) {
	stmt := onlyStmt(t, `loop true:
    pass`)
	loop := stmt.(*ast.LoopStmt)
	if len(loop.Iterators) != 1 || loop.Iterators[0].Variable != "_" {
		t.Fatalf("bad infinite loop: %#v", loop)
	}
}

func TestBreakContinueInsideLoop(t *testing.T) {
	stmt := onlyStmt(t, `loop true:
    break
    continue`)
	loop := stmt.(*ast.LoopStmt)
	if len(loop.Body) != 2 {
		t.Fatalf("loop body = %d", len(loop.Body))
	}

	_, diag := parseSource(t, `package app.main

func main():
    break
`)
	requireErrorCode(t, diag, "E016")
}

func TestReturnWithAndWithoutValue(t *testing.T) {
	fn := onlyFunc(t, `package app.main

func main():
    return
    return 1
`)
	if fn.Body[0].(*ast.ReturnStmt).Value != nil {
		t.Fatalf("first return should be bare")
	}
	if fn.Body[1].(*ast.ReturnStmt).Value == nil {
		t.Fatalf("second return should have value")
	}
}

func TestPass(t *testing.T) {
	stmt := onlyStmt(t, `pass`)
	if _, ok := stmt.(*ast.PassStmt); !ok {
		t.Fatalf("stmt = %T", stmt)
	}
}

func TestTypeDeclaration(t *testing.T) {
	program, diag := parseSource(t, `package app.main

type Point:
    X: int
    Y: int
    constructor(x: int, y: int):
        X = x
        Y = y
    func Greet() -> string:
        return "hello"
`)
	requireNoErrors(t, diag)
	typ := program.Decls[0].(*ast.TypeDecl)
	if len(typ.Fields) != 2 || typ.Constructor == nil || len(typ.Methods) != 1 {
		t.Fatalf("bad type decl: %#v", typ)
	}
}

func TestFieldDefaultAndExtends(t *testing.T) {
	program, diag := parseSource(t, `package app.main

type Dog ext Animal:
    name: string = "dog"
`)
	requireNoErrors(t, diag)
	typ := program.Decls[0].(*ast.TypeDecl)
	if typ.Extends != "Animal" || typ.Fields[0].Default == nil {
		t.Fatalf("bad type: %#v", typ)
	}
}

func TestNullableAndGenericTypes(t *testing.T) {
	program, diag := parseSource(t, `package app.main

name: string?
items: List<int>
flags: Dict<string, bool>
`)
	requireNoErrors(t, diag)
	if _, ok := program.Decls[0].(*ast.VarDecl).Type.(*ast.NullableType); !ok {
		t.Fatalf("first type = %T", program.Decls[0].(*ast.VarDecl).Type)
	}
	if _, ok := program.Decls[1].(*ast.VarDecl).Type.(*ast.GenericType); !ok {
		t.Fatalf("second type = %T", program.Decls[1].(*ast.VarDecl).Type)
	}
	if got := len(program.Decls[2].(*ast.VarDecl).Type.(*ast.GenericType).Params); got != 2 {
		t.Fatalf("dict params = %d", got)
	}
}

func TestErrorRecovery(t *testing.T) {
	program, diag := parseSource(t, `package app.main

func broken():
    value =

func ok():
    pass
`)
	if !diag.HasErrors() {
		t.Fatalf("expected errors")
	}
	if len(program.Decls) != 2 {
		t.Fatalf("decl count = %d", len(program.Decls))
	}
	if program.Decls[1].(*ast.FuncDecl).Name != "ok" {
		t.Fatalf("parser did not recover to ok function")
	}
}

func TestErrorRecoveryTopLevelInvalidToken(t *testing.T) {
	_, diag := parseSource(t, `package app.main

func main():
    loop i in 0..10, step 1, step 2:
        pass

else:
    pass

func ok():
    pass
`)
	requireErrorCode(t, diag, "E017")
	if !diag.HasErrors() {
		t.Fatalf("expected parser diagnostics")
	}
}

func TestEndToEndSnippet(t *testing.T) {
	program, diag := parseSource(t, `package app.main

import app.util as util

const MAX: int = 100

type Point:
    X: int
    Y: int
    constructor(x: int, y: int):
        X = x
        Y = y

func main():
    p = Point(x: 10, y: 20)
    result: string = switch p.X:
        10 -> "ten"
        _  -> "other"
    loop i in 0..MAX, step 2, while i < 50:
        console.print_ln($"i is {i}")
`)
	requireNoErrors(t, diag)
	if program.Package == nil || len(program.Imports) != 1 || len(program.Decls) != 3 {
		t.Fatalf("bad program shape: %#v", program)
	}
	fn := program.Decls[2].(*ast.FuncDecl)
	if len(fn.Body) != 3 {
		t.Fatalf("main body len = %d", len(fn.Body))
	}
	if _, ok := fn.Body[1].(*ast.VarDecl).Value.(*ast.SwitchExpr); !ok {
		t.Fatalf("result value = %T", fn.Body[1].(*ast.VarDecl).Value)
	}
	if _, ok := fn.Body[2].(*ast.LoopStmt); !ok {
		t.Fatalf("third stmt = %T", fn.Body[2])
	}
}

func TestErrorConstRequiresValue(t *testing.T) {
	_, diag := parseSource(t, `package app.main

const X: int
`)
	requireErrorCode(t, diag, "E085")
}

func TestErrorInvalidAssignmentTarget(t *testing.T) {
	_, diag := parseSource(t, `package app.main

func main():
    1 = 2
`)
	requireErrorCode(t, diag, "E014")
}

func TestErrorElseWithoutMatchingIf(t *testing.T) {
	_, diag := parseSource(t, `package app.main

func main():
    else:
        pass
`)
	requireErrorCode(t, diag, "E015")
}

func TestErrorContinueOutsideLoop(t *testing.T) {
	_, diag := parseSource(t, `package app.main

func main():
    continue
`)
	requireErrorCode(t, diag, "E016")
}

func TestErrorDuplicateWhileModifier(t *testing.T) {
	_, diag := parseSource(t, `package app.main

func main():
    loop i in 0..10, while i < 8, while i < 5:
        pass
`)
	requireErrorCode(t, diag, "E017")
}

func TestErrorDuplicateIfModifier(t *testing.T) {
	_, diag := parseSource(t, `package app.main

func main():
    loop i in 0..10, if i != 4, if i != 2:
        pass
`)
	requireErrorCode(t, diag, "E017")
}

func TestErrorPackageMustBeFirst(t *testing.T) {
	_, diag := parseSource(t, `import app.util

package app.main
`)
	requireErrorCode(t, diag, "E019")
}

func TestErrorMultiplePackageDeclarations(t *testing.T) {
	_, diag := parseSource(t, `package app.main

package app.other
`)
	requireErrorCode(t, diag, "E020")
}

func TestErrorEmptySwitchStatement(t *testing.T) {
	// Build token stream directly: empty switch body
	toks := []token.Token{
		{Kind: token.SWITCH, Pos: token.Position{File: "test.ll", Line: 1, Col: 1}},
		{Kind: token.IDENT, Lexeme: "x", Pos: token.Position{File: "test.ll", Line: 1, Col: 8}},
		{Kind: token.COLON, Pos: token.Position{File: "test.ll", Line: 1, Col: 9}},
		{Kind: token.NEWLINE, Pos: token.Position{File: "test.ll", Line: 1, Col: 10}},
		{Kind: token.INDENT, Pos: token.Position{File: "test.ll", Line: 2, Col: 1}},
		{Kind: token.DEDENT, Pos: token.Position{File: "test.ll", Line: 2, Col: 1}},
	}
	diag := diagnostics.New()
	p := New(toks, diag)
	program := &ast.Program{
		Package: &ast.PackageDecl{Path: []string{"app"}},
		Decls: []ast.Decl{&ast.FuncDecl{
			Name: "main",
			Body: []ast.Stmt{p.parseSwitchStmt()},
		}},
	}
	_ = program
	requireErrorCode(t, diag, "E018")
}

func TestErrorEmptySwitchExpression(t *testing.T) {
	// Build token stream directly: empty switch expression body
	toks := []token.Token{
		{Kind: token.SWITCH, Pos: token.Position{File: "test.ll", Line: 1, Col: 1}},
		{Kind: token.IDENT, Lexeme: "x", Pos: token.Position{File: "test.ll", Line: 1, Col: 8}},
		{Kind: token.COLON, Pos: token.Position{File: "test.ll", Line: 1, Col: 9}},
		{Kind: token.NEWLINE, Pos: token.Position{File: "test.ll", Line: 1, Col: 10}},
		{Kind: token.INDENT, Pos: token.Position{File: "test.ll", Line: 2, Col: 1}},
		{Kind: token.DEDENT, Pos: token.Position{File: "test.ll", Line: 2, Col: 1}},
	}
	start := toks[0]
	diag := diagnostics.New()
	p := New(toks, diag)
	_ = p.parseSwitchExpr(start)
	requireErrorCode(t, diag, "E018")
}

func TestErrorUnterminatedInterpolationBrace(t *testing.T) {
	// $"Hello, {name" — the lexer produces an INTERP_STRING_LIT token;
	// the parser's interpolation splitter should emit E012 for the unmatched brace.
	_, diag := parseSource(t, `package app.main

func main():
    x = $"Hello, {name"
`)
	requireErrorCode(t, diag, "E012")
}

func TestPrettyPrinter(t *testing.T) {
	program, diag := parseSource(t, `package app.main

func greet(name: string):
    console.print_ln($"Hello, {name}")
`)
	requireNoErrors(t, diag)
	var buf bytes.Buffer
	ast.Print(&buf, program)
	out := buf.String()
	for _, want := range []string{"Program", "PackageDecl", "FuncDecl name=greet", "CallExpr", "InterpStringExpr"} {
		if !strings.Contains(out, want) {
			t.Fatalf("pretty print missing %q:\n%s", want, out)
		}
	}
}

func TestComponentDeclaration(t *testing.T) {
	program, diag := parseSource(t, `package app.main

Window[main]:
    title = "LealLang App"
    w = 800
    h = 600
`)
	requireNoErrors(t, diag)
	if len(program.Decls) != 1 {
		t.Fatalf("decl count = %d, want 1", len(program.Decls))
	}
	comp, ok := program.Decls[0].(*ast.ComponentDecl)
	if !ok {
		t.Fatalf("decl type = %T, want *ast.ComponentDecl", program.Decls[0])
	}
	if comp.Component != "Window" {
		t.Fatalf("component = %q, want Window", comp.Component)
	}
	if comp.ID != "main" {
		t.Fatalf("id = %q, want main", comp.ID)
	}
	if len(comp.Props) != 3 {
		t.Fatalf("props count = %d, want 3", len(comp.Props))
	}
	if comp.Props[0].Name != "title" {
		t.Fatalf("prop[0] name = %q, want title", comp.Props[0].Name)
	}
	if comp.Props[1].Name != "w" {
		t.Fatalf("prop[1] name = %q, want w", comp.Props[1].Name)
	}
	if comp.Props[2].Name != "h" {
		t.Fatalf("prop[2] name = %q, want h", comp.Props[2].Name)
	}
}

func TestComponentWithChildren(t *testing.T) {
	program, diag := parseSource(t, `package app.main

Window[main]:
    title = "App"

    Col[root]:
        gap = 12

        Label[title]:
            text = "Hello"

        Button[save_btn]:
            text = "Save"
`)
	requireNoErrors(t, diag)
	comp, ok := program.Decls[0].(*ast.ComponentDecl)
	if !ok {
		t.Fatalf("decl type = %T, want *ast.ComponentDecl", program.Decls[0])
	}
	if len(comp.Children) != 1 {
		t.Fatalf("children count = %d, want 1", len(comp.Children))
	}
	col := comp.Children[0]
	if col.Component != "Col" || col.ID != "root" {
		t.Fatalf("child = %s[%s], want Col[root]", col.Component, col.ID)
	}
	if len(col.Children) != 2 {
		t.Fatalf("col children count = %d, want 2", len(col.Children))
	}
	if col.Children[0].Component != "Label" || col.Children[0].ID != "title" {
		t.Fatalf("col child[0] = %s[%s], want Label[title]", col.Children[0].Component, col.Children[0].ID)
	}
	if col.Children[1].Component != "Button" || col.Children[1].ID != "save_btn" {
		t.Fatalf("col child[1] = %s[%s], want Button[save_btn]", col.Children[1].Component, col.Children[1].ID)
	}
}

func TestComponentEventBinding(t *testing.T) {
	program, diag := parseSource(t, `package app.main

Window[main]:
    title = "App"

    Button[save_btn]:
        text = "Save"
        on_click = handle_save
`)
	requireNoErrors(t, diag)
	comp, ok := program.Decls[0].(*ast.ComponentDecl)
	if !ok {
		t.Fatalf("decl type = %T, want *ast.ComponentDecl", program.Decls[0])
	}
	if len(comp.Children) != 1 {
		t.Fatalf("children count = %d, want 1", len(comp.Children))
	}
	btn := comp.Children[0]
	if len(btn.Events) != 1 {
		t.Fatalf("events count = %d, want 1", len(btn.Events))
	}
	ev := btn.Events[0]
	if ev.Event != "click" {
		t.Fatalf("event = %q, want click", ev.Event)
	}
	handler, ok := ev.Handler.(*ast.Ident)
	if !ok {
		t.Fatalf("handler type = %T, want *ast.Ident", ev.Handler)
	}
	if handler.Name != "handle_save" {
		t.Fatalf("handler name = %q, want handle_save", handler.Name)
	}
}

func TestOldComponentEventSyntaxDiagnostic(t *testing.T) {
	_, diag := parseSource(t, `package app.main

Window[main]:
    title = "App"

    Button[save_btn]:
        text = "Save"
        on click = handle_save
`)
	requireErrorCode(t, diag, "E089")
	foundHint := false
	for _, err := range diag.Errors() {
		if err.Code == "E089" && strings.Contains(err.Hint, "use on_click = handle_save") {
			foundHint = true
			break
		}
	}
	if !foundHint {
		t.Fatalf("missing migration hint, got:\n%s", diag.Format())
	}
}

func TestComponentOnPropertyStillParses(t *testing.T) {
	program, diag := parseSource(t, `package app.main

Window[main]:
    title = "App"

    Toggle[dark_mode]:
        on = true
`)
	requireNoErrors(t, diag)
	toggle := program.Decls[0].(*ast.ComponentDecl).Children[0]
	if len(toggle.Props) != 1 || toggle.Props[0].Name != "on" {
		t.Fatalf("props = %#v, want on property", toggle.Props)
	}
}

func TestEmptyComponentDeclParses(t *testing.T) {
	program, diag := parseSource(t, `package app.main

Window[main]:
    title = "App"

    MenuBar[menu]:
        Menu[file]:
            label = "File"
            MenuSeparator[sep]
`)
	requireNoErrors(t, diag)
	menu := program.Decls[0].(*ast.ComponentDecl).Children[0].Children[0]
	if len(menu.Children) != 1 {
		t.Fatalf("menu children count = %d, want 1", len(menu.Children))
	}
	sep := menu.Children[0]
	if sep.Component != "MenuSeparator" || sep.ID != "sep" {
		t.Fatalf("component = %s[%s], want MenuSeparator[sep]", sep.Component, sep.ID)
	}
	if len(sep.Props) != 0 || len(sep.Children) != 0 || len(sep.Events) != 0 {
		t.Fatalf("empty component has body content: %#v", sep)
	}
}

func TestUnknownOnUnderscoreEventParsesForChecker(t *testing.T) {
	program, diag := parseSource(t, `package app.main

Window[main]:
    title = "App"

    Button[save_btn]:
        text = "Save"
        on_unknown_event = handle_save
`)
	requireNoErrors(t, diag)
	btn := program.Decls[0].(*ast.ComponentDecl).Children[0]
	if len(btn.Events) != 1 {
		t.Fatalf("events count = %d, want 1", len(btn.Events))
	}
	if btn.Events[0].Event != "unknown_event" {
		t.Fatalf("event = %q, want unknown_event", btn.Events[0].Event)
	}
}

func TestComponentMultipleProps(t *testing.T) {
	program, diag := parseSource(t, `package app.main

Window[main]:
    title = "App"
    w = 800
    h = 600
    bg = "white"
`)
	requireNoErrors(t, diag)
	comp, ok := program.Decls[0].(*ast.ComponentDecl)
	if !ok {
		t.Fatalf("decl type = %T, want *ast.ComponentDecl", program.Decls[0])
	}
	if len(comp.Props) != 4 {
		t.Fatalf("props count = %d, want 4", len(comp.Props))
	}
	names := make([]string, len(comp.Props))
	for i, p := range comp.Props {
		names[i] = p.Name
	}
	expected := []string{"title", "w", "h", "bg"}
	for i, want := range expected {
		if names[i] != want {
			t.Fatalf("prop[%d] = %q, want %q", i, names[i], want)
		}
	}
}

func TestComponentRefExprStillWorks(t *testing.T) {
	program, diag := parseSource(t, `package app.main

func handle_save():
    @Label[title].text = "Saved"
`)
	requireNoErrors(t, diag)
	fn := program.Decls[0].(*ast.FuncDecl)
	if fn.UI {
		t.Fatal("expected non-ui func")
	}
	if len(fn.Body) != 1 {
		t.Fatalf("body len = %d, want 1", len(fn.Body))
	}
	assign, ok := fn.Body[0].(*ast.AssignStmt)
	if !ok {
		t.Fatalf("stmt type = %T, want *ast.AssignStmt", fn.Body[0])
	}
	field, ok := assign.Target.(*ast.FieldExpr)
	if !ok {
		t.Fatalf("target type = %T, want *ast.FieldExpr", assign.Target)
	}
	compRef, ok := field.X.(*ast.ComponentRefExpr)
	if !ok {
		t.Fatalf("field.X type = %T, want *ast.ComponentRefExpr", field.X)
	}
	if compRef.Component != "Label" || compRef.ID != "title" {
		t.Fatalf("component ref = @%s[%s], want @Label[title]", compRef.Component, compRef.ID)
	}
	if field.Field != "text" {
		t.Fatalf("field = %q, want text", field.Field)
	}
}

func TestOnOutsideComponentBlock(t *testing.T) {
	_, diag := parseSource(t, `package app.main

func main():
    on click = handle_save
`)
	requireErrorCode(t, diag, "E089")
}

func TestUIFuncInsideWindowBlock(t *testing.T) {
	program, diag := parseSource(t, `package app.main

Window[main]:
    title = "App"
    w = 800
    h = 600

    ui func handle_save():
        @Label[title].text = "Saved"

    Col[root]:
        gap = 12

        Label[title]:
            text = "Hello"

        Button[save_btn]:
            text = "Save"
            on_click = handle_save
`)
	requireNoErrors(t, diag)
	if len(program.Decls) != 1 {
		t.Fatalf("decl count = %d, want 1", len(program.Decls))
	}
	comp, ok := program.Decls[0].(*ast.ComponentDecl)
	if !ok {
		t.Fatalf("decl type = %T, want *ast.ComponentDecl", program.Decls[0])
	}
	if comp.Component != "Window" || comp.ID != "main" {
		t.Fatalf("component = %s[%s], want Window[main]", comp.Component, comp.ID)
	}
	if len(comp.Funcs) != 1 {
		t.Fatalf("funcs count = %d, want 1", len(comp.Funcs))
	}
	if comp.Funcs[0].Name != "handle_save" {
		t.Fatalf("func name = %q, want handle_save", comp.Funcs[0].Name)
	}
	if !comp.Funcs[0].UI {
		t.Fatal("expected ui func")
	}
	if len(comp.Children) != 1 {
		t.Fatalf("children count = %d, want 1", len(comp.Children))
	}
}

func TestUIFuncOutsideWindowBlock(t *testing.T) {
	_, diag := parseSource(t, `package app.main

ui func view():
    Col[root]:
        gap = 12
`)
	requireErrorCode(t, diag, "E081")
}

func TestUIFuncInsideNonWindowComponent(t *testing.T) {
	_, diag := parseSource(t, `package app.main

Window[main]:
    title = "App"

    Col[root]:
        gap = 12

        ui func bad():
            console.print_ln("nope")
`)
	requireErrorCode(t, diag, "E081")
}
