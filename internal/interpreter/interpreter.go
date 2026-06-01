// Copyright (c) 2026 LealLang Contributors
// SPDX-License-Identifier: MIT

package interpreter

import (
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/LealLang/leallang/internal/ast"
	"github.com/LealLang/leallang/internal/diagnostics"
	"github.com/LealLang/leallang/internal/token"
	"github.com/LealLang/leallang/internal/uiir"
)

// syncWriter wraps an io.Writer with a mutex for safe concurrent use.
// This is needed because async tasks share the parent's stdout/stderr writers.
type syncWriter struct {
	mu sync.Mutex
	w  io.Writer
}

func (sw *syncWriter) Write(p []byte) (int, error) {
	sw.mu.Lock()
	defer sw.mu.Unlock()
	return sw.w.Write(p)
}

const maxCallDepth = 1000

type Interpreter struct {
	diag      *diagnostics.Diagnostics
	globals   *Env
	builtins  map[string]Value
	types     map[string]*ast.TypeDecl
	stdout    io.Writer
	stderr    io.Writer
	args      []string
	callDepth int
	exitCode  int
	uiBackend *uiir.Log // nil = stub mode, non-nil = record to log
}

func New(diag *diagnostics.Diagnostics) *Interpreter {
	return &Interpreter{
		diag:     diag,
		globals:  NewEnv(nil),
		builtins: make(map[string]Value),
		types:    make(map[string]*ast.TypeDecl),
		stdout:   os.Stdout,
		stderr:   os.Stderr,
	}
}

func (interp *Interpreter) SetOutput(stdout, stderr io.Writer) {
	interp.stdout = &syncWriter{w: discardWriter(stdout)}
	interp.stderr = &syncWriter{w: discardWriter(stderr)}
}

func (interp *Interpreter) SetArgs(args []string) {
	interp.args = append([]string(nil), args...)
}

// SetUIBackend sets the UI backend log for recording UI operations.
// When set, UI operations are recorded to the log instead of printing stubs.
func (interp *Interpreter) SetUIBackend(log *uiir.Log) {
	interp.uiBackend = log
}

func (interp *Interpreter) ExitCode() int {
	return interp.exitCode
}

// NewChild creates a child interpreter for an async task. The child has its own
// mutable execution state (globals, callDepth, diag, builtins) but shares
// read-only program metadata (types, args) and writers (stdout, stderr) with
// the parent. Builtins are re-registered on the child so they close over the
// child interpreter. FuncVal entries in taskEnv are cloned with closures
// re-bound to taskEnv so they do not reach into the caller's scope chain.
func (interp *Interpreter) NewChild(taskEnv *Env) *Interpreter {
	// Clone FuncVal entries and rebind closures to the task environment.
	for name, c := range taskEnv.vars {
		if fv, ok := c.value.(*FuncVal); ok {
			taskEnv.vars[name] = &cell{
				value: &FuncVal{
					Name:    fv.Name,
					Params:  fv.Params,
					Body:    fv.Body,
					Closure: taskEnv,
					Async:   fv.Async,
					UI:      fv.UI,
				},
				constBind: c.constBind,
			}
		}
	}
	child := &Interpreter{
		diag:      diagnostics.New(),
		globals:   taskEnv,
		types:     interp.types, // read-only after init
		stdout:    interp.stdout,
		stderr:    interp.stderr,
		args:      interp.args, // read-only after init
		uiBackend: interp.uiBackend, // shared with parent
	}
	child.registerBuiltins()
	return child
}

func (interp *Interpreter) Run(program *ast.Program) error {
	interp.globals = NewEnv(nil)
	interp.builtins = make(map[string]Value)
	interp.types = make(map[string]*ast.TypeDecl)
	interp.callDepth = 0
	interp.exitCode = 0
	interp.registerBuiltins()

	for _, decl := range program.Decls {
		if d, ok := decl.(*ast.TypeDecl); ok {
			interp.types[d.Name] = d
		}
	}
	for _, decl := range program.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			interp.globals.Set(d.Name, &FuncVal{Name: d.Name, Params: d.Params, Body: d.Body, Closure: interp.globals, Async: d.Async, UI: d.UI})
		case *ast.TypeDecl:
			interp.globals.Set(d.Name, &RecordTypeVal{Decl: d})
		}
	}
	for _, decl := range program.Decls {
		switch d := decl.(type) {
		case *ast.VarDecl:
			if _, err := interp.evalVarDecl(d, interp.globals, false); err != nil {
				return err
			}
		case *ast.ConstDecl:
			if _, err := interp.evalConstDecl(d, interp.globals); err != nil {
				return err
			}
		}
	}
	if mainVal, ok := interp.globals.Get("main"); ok {
		_, err := interp.callFunc(mainVal, nil, token.Position{})
		return interp.handleExit(err)
	}
	return nil
}

func (interp *Interpreter) handleExit(err error) error {
	if err == nil {
		return nil
	}
	var exit exitError
	if errors.As(err, &exit) {
		interp.exitCode = exit.code
		return nil
	}
	return err
}

func (interp *Interpreter) evalExpr(expr ast.Expr, env *Env) (Value, error) {
	switch e := expr.(type) {
	case *ast.Literal:
		return interp.evalLiteral(e)
	case *ast.Ident:
		if val, ok := env.Get(e.Name); ok {
			return val, nil
		}
		return nil, interp.runtimeError("E100", e.NamePos, 0, fmt.Sprintf("undefined variable '%s'", e.Name), "")
	case *ast.InterpStringExpr:
		var b strings.Builder
		for _, seg := range e.Segments {
			if !seg.IsExpr {
				b.WriteString(seg.Text)
				continue
			}
			val, err := interp.evalExpr(seg.Expr, env)
			if err != nil {
				return nil, err
			}
			b.WriteString(val.String())
		}
		return StringVal(b.String()), nil
	case *ast.BinaryExpr:
		return interp.evalBinary(e, env)
	case *ast.UnaryExpr:
		return interp.evalUnary(e, env)
	case *ast.AwaitExpr:
		val, err := interp.evalExpr(e.X, env)
		if err != nil {
			return nil, err
		}
		if task, ok := val.(*TaskVal); ok {
			result, taskErr := task.await()
			return &TupleVal{Elements: []Value{result, taskErr}}, nil
		}
		return val, nil
	case *ast.CallExpr:
		return interp.evalCall(e, env)
	case *ast.FieldExpr:
		return interp.evalField(e, env)
	case *ast.IndexExpr:
		return interp.evalIndex(e, env)
	case *ast.ComponentRefExpr:
		return &ComponentRefVal{Component: e.Component, ID: e.ID}, nil
	case *ast.RangeExpr:
		low, err := interp.evalExpr(e.Low, env)
		if err != nil {
			return nil, err
		}
		high, err := interp.evalExpr(e.High, env)
		if err != nil {
			return nil, err
		}
		lowInt, ok := low.(IntVal)
		if !ok {
			return nil, interp.runtimeError("E101", e.Low.Pos(), 0, "range lower bound must be int", "")
		}
		highInt, ok := high.(IntVal)
		if !ok {
			return nil, interp.runtimeError("E101", e.High.Pos(), 0, "range upper bound must be int", "")
		}
		return &RangeVal{Low: int64(lowInt), High: int64(highInt), Exclusive: e.Exclusive}, nil
	case *ast.ListLiteral:
		items := make([]Value, 0, len(e.Elements))
		for _, el := range e.Elements {
			val, err := interp.evalExpr(el, env)
			if err != nil {
				return nil, err
			}
			items = append(items, val)
		}
		return &ListVal{Elements: items}, nil
	case *ast.DictLiteral:
		items := make(map[string]Value, len(e.Pairs))
		for _, pair := range e.Pairs {
			keyVal, err := interp.evalExpr(pair.Key, env)
			if err != nil {
				return nil, err
			}
			key, ok := keyVal.(StringVal)
			if !ok {
				return nil, interp.runtimeError("E101", pair.Key.Pos(), 0, "dict keys must be strings", "")
			}
			val, err := interp.evalExpr(pair.Value, env)
			if err != nil {
				return nil, err
			}
			items[string(key)] = val
		}
		return &DictVal{Entries: items}, nil
	case *ast.SwitchExpr:
		return interp.evalSwitchExpr(e, env)
	case *ast.BadExpr:
		return nil, interp.runtimeError("E101", e.Start, 0, "cannot evaluate invalid expression", "")
	default:
		return nil, fmt.Errorf("unsupported expression %T", expr)
	}
}

func (interp *Interpreter) evalStmt(stmt ast.Stmt, env *Env) (*Signal, error) {
	switch s := stmt.(type) {
	case *ast.ExprStmt:
		_, err := interp.evalExpr(s.X, env)
		return nil, err
	case *ast.VarDecl:
		return interp.evalVarDecl(s, env, false)
	case *ast.ConstDecl:
		return interp.evalConstDecl(s, env)
	case *ast.AssignStmt:
		return nil, interp.evalAssign(s, env)
	case *ast.ReturnStmt:
		if s.Value == nil {
			return &Signal{Kind: signalReturn, Value: Null}, nil
		}
		val, err := interp.evalExpr(s.Value, env)
		if err != nil {
			return nil, err
		}
		return &Signal{Kind: signalReturn, Value: val}, nil
	case *ast.BreakStmt:
		return &Signal{Kind: signalBreak}, nil
	case *ast.ContinueStmt:
		return &Signal{Kind: signalContinue}, nil
	case *ast.PassStmt:
		return nil, nil
	case *ast.IfStmt:
		return interp.evalIf(s, env)
	case *ast.SwitchStmt:
		return interp.evalSwitchStmt(s, env)
	case *ast.LoopStmt:
		return interp.evalLoop(s, env)
	case *ast.ComponentDecl:
		return nil, interp.evalComponentDecl(s, env, "")
	default:
		return nil, fmt.Errorf("unsupported statement %T", stmt)
	}
}

func (interp *Interpreter) evalBlock(stmts []ast.Stmt, env *Env) (*Signal, error) {
	for _, stmt := range stmts {
		sig, err := interp.evalStmt(stmt, env)
		if err != nil || sig != nil {
			return sig, err
		}
	}
	return nil, nil
}

func (interp *Interpreter) callFunc(fn Value, args []Value, pos token.Position) (Value, error) {
	return interp.callFuncWithRefs(fn, args, nil, pos)
}

func (interp *Interpreter) callFuncWithRefs(fn Value, args []Value, refArgs []*cell, pos token.Position) (Value, error) {
	if interp.callDepth >= maxCallDepth {
		return nil, interp.runtimeError("E107", pos, 0, "call depth exceeded", "")
	}
	interp.callDepth++
	defer func() { interp.callDepth-- }()

	switch f := fn.(type) {
	case *BuiltinVal:
		return f.Fn(args)
	case *FuncVal:
		if f.Async {
			// Clone arguments for sendable boundary.
			clonedArgs := make([]Value, len(args))
			for i, arg := range args {
				cloned, err := cloneForTask(arg)
				if err != nil {
					return nil, err
				}
				clonedArgs[i] = cloned
			}
			// Build isolated child environment from globals snapshot.
			taskEnv := interp.globals.SnapshotGlobals()
			child := interp.NewChild(taskEnv)
			// Bind cloned args against child globals (not caller's closure).
			callEnv, err := child.bindCallEnv(f.Params, clonedArgs, nil, taskEnv, pos)
			if err != nil {
				return nil, err
			}
			task := newTaskVal()
			go func() {
				sig, callErr := child.evalBlock(f.Body, callEnv)
				if callErr != nil {
					task.resolve(Null, errorRecord(callErr.Error(), "task_failed"))
				} else if sig != nil && sig.Kind == signalReturn {
					task.resolve(sig.Value, Null)
				} else {
					task.resolve(Null, Null)
				}
			}()
			return task, nil
		}
		callEnv, err := interp.bindCallEnv(f.Params, args, refArgs, f.Closure, pos)
		if err != nil {
			return nil, err
		}
		sig, err := interp.evalBlock(f.Body, callEnv)
		if err != nil {
			return nil, err
		}
		var result Value
		if sig != nil {
			if sig.Kind != signalReturn {
				return nil, interp.runtimeError("E101", pos, 0, sig.Error()+" outside loop", "")
			}
			result = sig.Value
		} else {
			result = Null
		}
		return result, nil
	case *BoundMethodVal:
		return interp.callMethod(f, args, refArgs, pos)
	case *RecordTypeVal:
		return interp.callConstructor(f.Decl, args, pos)
	default:
		return nil, interp.runtimeError("E104", pos, 0, fmt.Sprintf("cannot call %s", fn.Type()), "")
	}
}

func (interp *Interpreter) evalLiteral(lit *ast.Literal) (Value, error) {
	switch lit.Kind {
	case token.INT_LIT:
		val, err := strconv.ParseInt(lit.Value, 10, 64)
		if err != nil {
			return nil, interp.runtimeError("E101", lit.ValuePos, 0, err.Error(), "")
		}
		return IntVal(val), nil
	case token.FLOAT_LIT:
		val, err := strconv.ParseFloat(lit.Value, 64)
		if err != nil {
			return nil, interp.runtimeError("E101", lit.ValuePos, 0, err.Error(), "")
		}
		return FloatVal(val), nil
	case token.STRING_LIT:
		val, err := strconv.Unquote(lit.Value)
		if err != nil {
			return nil, interp.runtimeError("E101", lit.ValuePos, 0, err.Error(), "")
		}
		return StringVal(val), nil
	case token.CHAR_LIT:
		val, err := strconv.Unquote(lit.Value)
		if err != nil {
			return nil, interp.runtimeError("E101", lit.ValuePos, 0, err.Error(), "")
		}
		runes := []rune(val)
		if len(runes) == 0 {
			return CharVal(0), nil
		}
		return CharVal(runes[0]), nil
	case token.TRUE:
		return BoolVal(true), nil
	case token.FALSE:
		return BoolVal(false), nil
	case token.NULL:
		return Null, nil
	default:
		return nil, interp.runtimeError("E101", lit.ValuePos, 0, "unknown literal", "")
	}
}

func (interp *Interpreter) evalBinary(expr *ast.BinaryExpr, env *Env) (Value, error) {
	if expr.Op == token.AND || expr.Op == token.OR {
		left, err := interp.evalExpr(expr.Left, env)
		if err != nil {
			return nil, err
		}
		leftBool, ok := left.(BoolVal)
		if !ok {
			return nil, interp.runtimeError("E101", expr.Left.Pos(), 0, "logical operand must be bool", "")
		}
		if expr.Op == token.AND && !bool(leftBool) {
			return BoolVal(false), nil
		}
		if expr.Op == token.OR && bool(leftBool) {
			return BoolVal(true), nil
		}
		right, err := interp.evalExpr(expr.Right, env)
		if err != nil {
			return nil, err
		}
		rightBool, ok := right.(BoolVal)
		if !ok {
			return nil, interp.runtimeError("E101", expr.Right.Pos(), 0, "logical operand must be bool", "")
		}
		return rightBool, nil
	}

	left, err := interp.evalExpr(expr.Left, env)
	if err != nil {
		return nil, err
	}
	right, err := interp.evalExpr(expr.Right, env)
	if err != nil {
		return nil, err
	}

	switch expr.Op {
	case token.PLUS:
		if l, ok := left.(StringVal); ok {
			if r, ok := right.(StringVal); ok {
				return StringVal(string(l) + string(r)), nil
			}
		}
		return interp.numericBinary(left, right, expr.Op, expr.OpPos)
	case token.MINUS, token.STAR, token.SLASH, token.PERCENT:
		return interp.numericBinary(left, right, expr.Op, expr.OpPos)
	case token.EQ_EQ:
		return BoolVal(valuesEqual(left, right)), nil
	case token.BANG_EQ:
		return BoolVal(!valuesEqual(left, right)), nil
	case token.LESS, token.LESS_EQ, token.GREATER, token.GREATER_EQ:
		return interp.compareValues(left, right, expr.Op, expr.OpPos)
	default:
		return nil, interp.runtimeError("E101", expr.OpPos, 0, "unsupported binary operator", "")
	}
}

func (interp *Interpreter) evalUnary(expr *ast.UnaryExpr, env *Env) (Value, error) {
	val, err := interp.evalExpr(expr.X, env)
	if err != nil {
		return nil, err
	}
	switch expr.Op {
	case token.MINUS:
		switch v := val.(type) {
		case IntVal:
			return IntVal(-int64(v)), nil
		case FloatVal:
			return FloatVal(-float64(v)), nil
		default:
			return nil, interp.runtimeError("E101", expr.OpPos, 0, "unary '-' expects number", "")
		}
	case token.NOT:
		v, ok := val.(BoolVal)
		if !ok {
			return nil, interp.runtimeError("E101", expr.OpPos, 0, "not expects bool", "")
		}
		return BoolVal(!bool(v)), nil
	default:
		return nil, interp.runtimeError("E101", expr.OpPos, 0, "unsupported unary operator", "")
	}
}

func (interp *Interpreter) evalCall(call *ast.CallExpr, env *Env) (Value, error) {
	callee, err := interp.evalExpr(call.Func, env)
	if err != nil {
		return nil, err
	}
	params := paramsForCallable(callee)
	args, refs, err := interp.evalCallArgs(call, env, params)
	if err != nil {
		return nil, err
	}
	val, err := interp.callFuncWithRefs(callee, args, refs, call.LParen)
	if err != nil {
		return nil, err
	}
	return val, nil
}

func paramsForCallable(v Value) []*ast.Param {
	switch fn := v.(type) {
	case *FuncVal:
		return fn.Params
	case *BoundMethodVal:
		return fn.Decl.Params
	case *RecordTypeVal:
		if fn.Decl.Constructor != nil {
			return fn.Decl.Constructor.Params
		}
		params := make([]*ast.Param, 0, len(fn.Decl.Fields))
		for _, field := range fn.Decl.Fields {
			params = append(params, &ast.Param{Name: field.Name, Type: field.Type, Posn: field.FieldPos})
		}
		return params
	default:
		return nil
	}
}

func (interp *Interpreter) evalCallArgs(call *ast.CallExpr, env *Env, params []*ast.Param) ([]Value, []*cell, error) {
	if params == nil {
		args := make([]Value, 0, len(call.Args))
		for _, arg := range call.Args {
			val, err := interp.evalExpr(arg.Value, env)
			if err != nil {
				return nil, nil, err
			}
			args = append(args, val)
		}
		return args, make([]*cell, len(args)), nil
	}

	values := make([]Value, len(params))
	refs := make([]*cell, len(params))
	provided := make([]bool, len(params))
	posIndex := 0
	for _, arg := range call.Args {
		idx := posIndex
		if arg.Name != "" {
			idx = -1
			for i, param := range params {
				if param.Name == arg.Name {
					idx = i
					break
				}
			}
			if idx < 0 {
				return nil, nil, interp.runtimeError("E105", arg.Posn, 0, fmt.Sprintf("unknown named argument '%s'", arg.Name), "")
			}
		} else {
			posIndex++
		}
		if idx >= len(params) {
			return nil, nil, interp.runtimeError("E105", call.RParen, 0, fmt.Sprintf("expected %d arguments, got %d", len(params), len(call.Args)), "")
		}
		if provided[idx] {
			return nil, nil, interp.runtimeError("E105", arg.Posn, 0, fmt.Sprintf("duplicate argument for '%s'", params[idx].Name), "")
		}
		val, err := interp.evalExpr(arg.Value, env)
		if err != nil {
			return nil, nil, err
		}
		values[idx] = val
		provided[idx] = true
		if params[idx].Ref {
			ident, ok := arg.Value.(*ast.Ident)
			if !ok {
				return nil, nil, interp.runtimeError("E101", arg.Value.Pos(), 0, fmt.Sprintf("ref parameter '%s' requires a variable", params[idx].Name), "")
			}
			c, ok := env.lookupCell(ident.Name)
			if !ok {
				return nil, nil, interp.runtimeError("E100", ident.NamePos, 0, fmt.Sprintf("undefined variable '%s'", ident.Name), "")
			}
			refs[idx] = c
		}
	}
	for i, ok := range provided {
		if !ok {
			return nil, nil, interp.runtimeError("E105", call.LParen, 0, fmt.Sprintf("missing argument '%s'", params[i].Name), "")
		}
	}
	return values, refs, nil
}

func (interp *Interpreter) evalField(expr *ast.FieldExpr, env *Env) (Value, error) {
	base, err := interp.evalExpr(expr.X, env)
	if err != nil {
		return nil, err
	}
	switch v := base.(type) {
	case *RecordVal:
		if val, ok := v.Fields[expr.Field]; ok {
			return val, nil
		}
		if decl := interp.types[v.TypeName]; decl != nil {
			for _, method := range decl.Methods {
				if method.Name == expr.Field {
					return &BoundMethodVal{Receiver: v, Decl: method, Closure: env}, nil
				}
			}
		}
		return nil, interp.runtimeError("E101", expr.Dot, 0, fmt.Sprintf("type %s has no field '%s'", v.TypeName, expr.Field), "")
	case *NamespaceVal:
		if val, ok := v.Members[expr.Field]; ok {
			return val, nil
		}
		return nil, interp.runtimeError("E100", expr.Dot, 0, fmt.Sprintf("namespace %s has no member '%s'", v.Name, expr.Field), "")
	case *ConstGroupVal:
		if val, ok := v.Members[expr.Field]; ok {
			return val, nil
		}
		return nil, interp.runtimeError("E100", expr.Dot, 0, fmt.Sprintf("const group %s has no member '%s'", v.Name, expr.Field), "")
	case *ComponentRefVal:
		fmt.Fprintf(interp.stderr, "[stub] UI property %s.%s is not implemented\n", v.String(), expr.Field)
		return Null, nil
	default:
		return nil, interp.runtimeError("E101", expr.Dot, 0, fmt.Sprintf("%s has no fields", base.Type()), "")
	}
}

func (interp *Interpreter) evalIndex(expr *ast.IndexExpr, env *Env) (Value, error) {
	base, err := interp.evalExpr(expr.X, env)
	if err != nil {
		return nil, err
	}
	idx, err := interp.evalExpr(expr.Index, env)
	if err != nil {
		return nil, err
	}
	switch v := base.(type) {
	case *ListVal:
		i, ok := idx.(IntVal)
		if !ok {
			return nil, interp.runtimeError("E101", expr.Index.Pos(), 0, "list index must be int", "")
		}
		if i < 0 || int(i) >= len(v.Elements) {
			return nil, interp.runtimeError("E102", expr.LBrack, 0, "list index out of bounds", "")
		}
		return v.Elements[int(i)], nil
	case *DictVal:
		key, ok := idx.(StringVal)
		if !ok {
			return nil, interp.runtimeError("E101", expr.Index.Pos(), 0, "dict index must be string", "")
		}
		if val, ok := v.Entries[string(key)]; ok {
			return val, nil
		}
		return nil, interp.runtimeError("E106", expr.LBrack, 0, fmt.Sprintf("key %q not found in dict", string(key)), "")
	case StringVal:
		i, ok := idx.(IntVal)
		if !ok {
			return nil, interp.runtimeError("E101", expr.Index.Pos(), 0, "string index must be int", "")
		}
		runes := []rune(string(v))
		if i < 0 || int(i) >= len(runes) {
			return nil, interp.runtimeError("E102", expr.LBrack, 0, "string index out of bounds", "")
		}
		return CharVal(runes[int(i)]), nil
	default:
		return nil, interp.runtimeError("E101", expr.LBrack, 0, fmt.Sprintf("%s is not indexable", base.Type()), "")
	}
}

func (interp *Interpreter) evalSwitchExpr(expr *ast.SwitchExpr, env *Env) (Value, error) {
	subject, err := interp.evalExpr(expr.Subject, env)
	if err != nil {
		return nil, err
	}
	for _, arm := range expr.Arms {
		match, err := interp.patternsMatch(subject, arm.Patterns, env)
		if err != nil {
			return nil, err
		}
		if match {
			return interp.evalExpr(arm.Value, env)
		}
	}
	return Null, nil
}

func (interp *Interpreter) evalVarDecl(decl *ast.VarDecl, env *Env, constBind bool) (*Signal, error) {
	var val Value
	var err error
	if decl.Value != nil {
		val, err = interp.evalExpr(decl.Value, env)
		if err != nil {
			return nil, err
		}
	} else {
		val = zeroValue(decl.Type)
	}
	if constBind {
		env.setConst(decl.Name, val)
	} else {
		env.Set(decl.Name, val)
	}
	return nil, nil
}

func (interp *Interpreter) evalConstDecl(decl *ast.ConstDecl, env *Env) (*Signal, error) {
	val := Value(Null)
	if decl.Value != nil {
		var err error
		val, err = interp.evalExpr(decl.Value, env)
		if err != nil {
			return nil, err
		}
	}
	env.setConst(decl.Name, val)
	return nil, nil
}

func (interp *Interpreter) evalAssign(stmt *ast.AssignStmt, env *Env) error {
	val, err := interp.evalExpr(stmt.Value, env)
	if err != nil {
		return err
	}
	switch target := stmt.Target.(type) {
	case *ast.Ident:
		found, isConst := env.Assign(target.Name, val)
		if isConst {
			return interp.runtimeError("E101", target.NamePos, 0, fmt.Sprintf("cannot reassign const '%s'", target.Name), "")
		}
		if found {
			return nil
		}
		env.Set(target.Name, val)
		return nil
	case *ast.FieldExpr:
		base, err := interp.evalExpr(target.X, env)
		if err != nil {
			return err
		}
		rec, ok := base.(*RecordVal)
		if !ok {
			if compRef, ok := base.(*ComponentRefVal); ok {
				if interp.uiBackend != nil {
					interp.uiBackend.Record(uiir.Op{
						Kind:      uiir.OpPropSet,
						ID:        compRef.ID,
						PropName:  target.Field,
						PropValue: valueToUI(val),
					})
					return nil
				}
				fmt.Fprintf(interp.stderr, "[stub] UI assignment %s.%s is not implemented\n", base.String(), target.Field)
				return nil
			}
			return interp.runtimeError("E101", target.Dot, 0, "field assignment requires a record", "")
		}
		rec.Fields[target.Field] = val
		return nil
	case *ast.IndexExpr:
		return interp.assignIndex(target, val, env)
	case *ast.ComponentRefExpr:
		fmt.Fprintf(interp.stderr, "[stub] UI assignment %s[%s] is not implemented\n", target.Component, target.ID)
		return nil
	default:
		return interp.runtimeError("E101", stmt.EqPos, 0, "invalid assignment target", "")
	}
}

func (interp *Interpreter) assignIndex(target *ast.IndexExpr, val Value, env *Env) error {
	base, err := interp.evalExpr(target.X, env)
	if err != nil {
		return err
	}
	idx, err := interp.evalExpr(target.Index, env)
	if err != nil {
		return err
	}
	switch v := base.(type) {
	case *ListVal:
		i, ok := idx.(IntVal)
		if !ok {
			return interp.runtimeError("E101", target.Index.Pos(), 0, "list index must be int", "")
		}
		if i < 0 || int(i) >= len(v.Elements) {
			return interp.runtimeError("E102", target.LBrack, 0, "list index out of bounds", "")
		}
		v.Elements[int(i)] = val
		return nil
	case *DictVal:
		key, ok := idx.(StringVal)
		if !ok {
			return interp.runtimeError("E101", target.Index.Pos(), 0, "dict index must be string", "")
		}
		v.Entries[string(key)] = val
		return nil
	default:
		return interp.runtimeError("E101", target.LBrack, 0, "index assignment requires list or dict", "")
	}
}

// evalComponentDecl evaluates a UI component declaration.
// If a UI backend is set, it records mount, prop-set, and event-bind operations.
// Otherwise, it prints a stub message.
func (interp *Interpreter) evalComponentDecl(decl *ast.ComponentDecl, env *Env, parentID string) error {
	if interp.uiBackend == nil {
		fmt.Fprintf(interp.stderr, "[stub] UI component %s[%s] is not implemented\n", decl.Component, decl.ID)
		return nil
	}

	// Mount the component.
	interp.uiBackend.Record(uiir.Op{
		Kind:      uiir.OpMount,
		Component: decl.Component,
		ID:        decl.ID,
		ParentID:  parentID,
	})

	// Set properties.
	for _, prop := range decl.Props {
		val, err := interp.evalExpr(prop.Value, env)
		if err != nil {
			return err
		}
		interp.uiBackend.Record(uiir.Op{
			Kind:      uiir.OpPropSet,
			ID:        decl.ID,
			PropName:  prop.Name,
			PropValue: valueToUI(val),
		})
	}

	// Bind events.
	for _, ev := range decl.Events {
		handlerName := ""
		if ident, ok := ev.Handler.(*ast.Ident); ok {
			handlerName = ident.Name
		}
		interp.uiBackend.Record(uiir.Op{
			Kind:      uiir.OpEventBind,
			ID:        decl.ID,
			EventName: ev.Event,
			HandlerID: handlerName,
		})
	}

	// Recurse into children.
	for _, child := range decl.Children {
		if err := interp.evalComponentDecl(child, env, decl.ID); err != nil {
			return err
		}
	}

	return nil
}

func (interp *Interpreter) evalIf(stmt *ast.IfStmt, env *Env) (*Signal, error) {
	cond, err := interp.evalBool(stmt.Condition, env)
	if err != nil {
		return nil, err
	}
	if cond {
		return interp.evalBlock(stmt.Body, NewEnv(env))
	}
	for _, elseIf := range stmt.ElseIfs {
		cond, err := interp.evalBool(elseIf.Condition, env)
		if err != nil {
			return nil, err
		}
		if cond {
			return interp.evalBlock(elseIf.Body, NewEnv(env))
		}
	}
	if stmt.Else != nil {
		return interp.evalBlock(stmt.Else, NewEnv(env))
	}
	return nil, nil
}

func (interp *Interpreter) evalSwitchStmt(stmt *ast.SwitchStmt, env *Env) (*Signal, error) {
	subject, err := interp.evalExpr(stmt.Subject, env)
	if err != nil {
		return nil, err
	}
	for _, c := range stmt.Cases {
		match, err := interp.patternsMatch(subject, c.Patterns, env)
		if err != nil {
			return nil, err
		}
		if match {
			return interp.evalBlock(c.Body, NewEnv(env))
		}
	}
	return nil, nil
}

func (interp *Interpreter) evalLoop(stmt *ast.LoopStmt, env *Env) (*Signal, error) {
	if len(stmt.Iterators) == 0 {
		return nil, nil
	}
	iters, infinite, err := interp.buildIterators(stmt, env)
	if err != nil {
		return nil, err
	}
	step := int64(1)
	if stmt.Step != nil {
		val, err := interp.evalExpr(stmt.Step, env)
		if err != nil {
			return nil, err
		}
		intVal, ok := val.(IntVal)
		if !ok || intVal <= 0 {
			return nil, interp.runtimeError("E101", stmt.Step.Pos(), 0, "loop step must be positive int", "")
		}
		step = int64(intVal)
	}

	for index := int64(0); ; index += step {
		loopEnv := NewEnv(env)
		if infinite {
			if index > math.MaxInt32 {
				return nil, interp.runtimeError("E107", stmt.LoopPos, 0, "infinite loop exceeded safety limit", "")
			}
		} else {
			stop := false
			for _, iter := range iters {
				if !iter.bind(loopEnv, index) {
					stop = true
					break
				}
			}
			if stop {
				break
			}
		}
		if infinite {
			loopEnv.Set(stmt.Iterators[0].Variable, BoolVal(true))
		}
		if stmt.While != nil {
			ok, err := interp.evalBool(stmt.While, loopEnv)
			if err != nil {
				return nil, err
			}
			if !ok {
				break
			}
		}
		if stmt.IfCond != nil {
			ok, err := interp.evalBool(stmt.IfCond, loopEnv)
			if err != nil {
				return nil, err
			}
			if !ok {
				continue
			}
		}
		sig, err := interp.evalBlock(stmt.Body, loopEnv)
		if err != nil {
			return nil, err
		}
		if sig == nil {
			continue
		}
		switch sig.Kind {
		case signalBreak:
			return nil, nil
		case signalContinue:
			continue
		case signalReturn:
			return sig, nil
		}
	}
	return nil, nil
}

type loopIter struct {
	name   string
	values []Value
}

func (it loopIter) bind(env *Env, index int64) bool {
	if index < 0 || int(index) >= len(it.values) {
		return false
	}
	if it.name != "_" {
		env.Set(it.name, it.values[int(index)])
	}
	return true
}

func (interp *Interpreter) buildIterators(stmt *ast.LoopStmt, env *Env) ([]loopIter, bool, error) {
	if len(stmt.Iterators) == 1 {
		val, err := interp.evalExpr(stmt.Iterators[0].Iterable, env)
		if err != nil {
			return nil, false, err
		}
		if b, ok := val.(BoolVal); ok && bool(b) {
			return nil, true, nil
		}
	}
	iters := make([]loopIter, 0, len(stmt.Iterators))
	for _, item := range stmt.Iterators {
		val, err := interp.evalExpr(item.Iterable, env)
		if err != nil {
			return nil, false, err
		}
		values, err := interp.iterableValues(val, item.IterPos)
		if err != nil {
			return nil, false, err
		}
		iters = append(iters, loopIter{name: item.Variable, values: values})
	}
	return iters, false, nil
}

func (interp *Interpreter) iterableValues(val Value, pos token.Position) ([]Value, error) {
	switch v := val.(type) {
	case *RangeVal:
		if v.Exclusive && v.Low == v.High {
			return nil, nil
		}
		var out []Value
		end := v.High
		if v.Exclusive && v.Low <= v.High {
			end--
		} else if v.Exclusive && v.Low > v.High {
			end++
		}
		if v.Low <= end {
			for i := v.Low; i <= end; i++ {
				out = append(out, IntVal(i))
			}
		} else {
			for i := v.Low; i >= end; i-- {
				out = append(out, IntVal(i))
			}
		}
		return out, nil
	case *ListVal:
		return v.Elements, nil
	case *DictVal:
		out := make([]Value, 0, len(v.Entries))
		for key, value := range v.Entries {
			out = append(out, &RecordVal{TypeName: "DictEntry", Fields: map[string]Value{"key": StringVal(key), "value": value}})
		}
		return out, nil
	default:
		return nil, interp.runtimeError("E101", pos, 0, fmt.Sprintf("%s is not iterable", val.Type()), "")
	}
}

func (interp *Interpreter) patternsMatch(subject Value, patterns []ast.Expr, env *Env) (bool, error) {
	if patterns == nil {
		return true, nil
	}
	for _, pattern := range patterns {
		if ident, ok := pattern.(*ast.Ident); ok && ident.Name == "_" {
			return true, nil
		}
		val, err := interp.evalExpr(pattern, env)
		if err != nil {
			return false, err
		}
		if valuesEqual(subject, val) {
			return true, nil
		}
	}
	return false, nil
}

func (interp *Interpreter) bindCallEnv(params []*ast.Param, args []Value, refs []*cell, parent *Env, pos token.Position) (*Env, error) {
	if len(args) != len(params) {
		return nil, interp.runtimeError("E105", pos, 0, fmt.Sprintf("expected %d arguments, got %d", len(params), len(args)), "")
	}
	callEnv := NewEnv(parent)
	for i, param := range params {
		if i < len(refs) && refs[i] != nil {
			callEnv.bindAlias(param.Name, refs[i])
		} else {
			callEnv.Set(param.Name, args[i])
		}
	}
	return callEnv, nil
}

func (interp *Interpreter) callConstructor(decl *ast.TypeDecl, args []Value, pos token.Position) (Value, error) {
	rec := &RecordVal{TypeName: decl.Name, Fields: make(map[string]Value, len(decl.Fields))}
	for _, field := range decl.Fields {
		if field.Default != nil {
			val, err := interp.evalExpr(field.Default, interp.globals)
			if err != nil {
				return nil, err
			}
			rec.Fields[field.Name] = val
		} else {
			rec.Fields[field.Name] = zeroValue(field.Type)
		}
	}
	if decl.Constructor == nil {
		if len(args) != len(decl.Fields) {
			return nil, interp.runtimeError("E105", pos, 0, fmt.Sprintf("expected %d arguments, got %d", len(decl.Fields), len(args)), "")
		}
		for i, field := range decl.Fields {
			rec.Fields[field.Name] = args[i]
		}
		return rec, nil
	}
	ctorEnv, err := interp.bindCallEnv(decl.Constructor.Params, args, nil, interp.globals, pos)
	if err != nil {
		return nil, err
	}
	for _, field := range decl.Fields {
		fieldCell := &cell{value: rec.Fields[field.Name]}
		ctorEnv.setCell(field.Name, fieldCell)
	}
	sig, err := interp.evalBlock(decl.Constructor.Body, ctorEnv)
	if err != nil {
		return nil, err
	}
	if sig != nil && sig.Kind == signalReturn {
		return nil, interp.runtimeError("E101", pos, 0, "constructor cannot return a value", "")
	}
	for _, field := range decl.Fields {
		if val, ok := ctorEnv.Get(field.Name); ok {
			rec.Fields[field.Name] = val
		}
	}
	return rec, nil
}

func (interp *Interpreter) callMethod(method *BoundMethodVal, args []Value, refs []*cell, pos token.Position) (Value, error) {
	methodEnv, err := interp.bindCallEnv(method.Decl.Params, args, refs, method.Closure, pos)
	if err != nil {
		return nil, err
	}
	fieldCells := make(map[string]*cell, len(method.Receiver.Fields))
	for name, value := range method.Receiver.Fields {
		c := &cell{value: value}
		fieldCells[name] = c
		methodEnv.setCell(name, c)
	}
	sig, err := interp.evalBlock(method.Decl.Body, methodEnv)
	if err != nil {
		return nil, err
	}
	for name, c := range fieldCells {
		method.Receiver.Fields[name] = c.value
	}
	if sig != nil {
		if sig.Kind != signalReturn {
			return nil, interp.runtimeError("E101", pos, 0, sig.Error()+" outside loop", "")
		}
		return sig.Value, nil
	}
	return Null, nil
}

func (interp *Interpreter) numericBinary(left, right Value, op token.TokenKind, pos token.Position) (Value, error) {
	li, lok := left.(IntVal)
	ri, rok := right.(IntVal)
	if lok && rok {
		switch op {
		case token.PLUS:
			return IntVal(int64(li) + int64(ri)), nil
		case token.MINUS:
			return IntVal(int64(li) - int64(ri)), nil
		case token.STAR:
			return IntVal(int64(li) * int64(ri)), nil
		case token.SLASH:
			if ri == 0 {
				return nil, interp.runtimeError("E103", pos, 0, "division by zero", "")
			}
			return IntVal(int64(li) / int64(ri)), nil
		case token.PERCENT:
			if ri == 0 {
				return nil, interp.runtimeError("E103", pos, 0, "modulo by zero", "")
			}
			return IntVal(int64(li) % int64(ri)), nil
		}
	}
	lf, lok := asFloat(left)
	rf, rok := asFloat(right)
	if !lok || !rok {
		return nil, interp.runtimeError("E101", pos, 0, "numeric operator requires numbers", "")
	}
	switch op {
	case token.PLUS:
		return FloatVal(lf + rf), nil
	case token.MINUS:
		return FloatVal(lf - rf), nil
	case token.STAR:
		return FloatVal(lf * rf), nil
	case token.SLASH:
		if rf == 0 {
			return nil, interp.runtimeError("E103", pos, 0, "division by zero", "")
		}
		return FloatVal(lf / rf), nil
	case token.PERCENT:
		if rf == 0 {
			return nil, interp.runtimeError("E103", pos, 0, "modulo by zero", "")
		}
		return FloatVal(math.Mod(lf, rf)), nil
	}
	return nil, interp.runtimeError("E101", pos, 0, "unsupported numeric operator", "")
}

func (interp *Interpreter) compareValues(left, right Value, op token.TokenKind, pos token.Position) (Value, error) {
	if l, ok := left.(StringVal); ok {
		r, ok := right.(StringVal)
		if !ok {
			return nil, interp.runtimeError("E101", pos, 0, "string comparison requires strings", "")
		}
		return BoolVal(compareOrdered(string(l), string(r), op)), nil
	}
	lf, lok := asFloat(left)
	rf, rok := asFloat(right)
	if !lok || !rok {
		return nil, interp.runtimeError("E101", pos, 0, "comparison requires numbers or strings", "")
	}
	return BoolVal(compareOrdered(lf, rf, op)), nil
}

func asFloat(v Value) (float64, bool) {
	switch val := v.(type) {
	case IntVal:
		return float64(val), true
	case FloatVal:
		return float64(val), true
	default:
		return 0, false
	}
}

func compareOrdered[T ~string | ~float64](left, right T, op token.TokenKind) bool {
	switch op {
	case token.LESS:
		return left < right
	case token.LESS_EQ:
		return left <= right
	case token.GREATER:
		return left > right
	case token.GREATER_EQ:
		return left >= right
	default:
		return false
	}
}

func (interp *Interpreter) evalBool(expr ast.Expr, env *Env) (bool, error) {
	val, err := interp.evalExpr(expr, env)
	if err != nil {
		return false, err
	}
	b, ok := val.(BoolVal)
	if !ok {
		return false, interp.runtimeError("E101", expr.Pos(), 0, "condition must be bool", "")
	}
	return bool(b), nil
}

func zeroValue(t ast.TypeExpr) Value {
	switch typ := t.(type) {
	case *ast.SimpleType:
		switch typ.Name {
		case "int":
			return IntVal(0)
		case "float":
			return FloatVal(0)
		case "string":
			return StringVal("")
		case "bool":
			return BoolVal(false)
		case "char":
			return CharVal(0)
		default:
			return Null
		}
	case *ast.NullableType:
		return Null
	case *ast.GenericType:
		switch typ.Name {
		case "List":
			return &ListVal{}
		case "Dict":
			return &DictVal{Entries: make(map[string]Value)}
		default:
			return Null
		}
	default:
		return Null
	}
}

func (interp *Interpreter) runtimeError(code string, pos token.Position, span int, msg, hint string) error {
	if interp.diag != nil {
		interp.diag.ReportError(code, msg, pos, span, hint, "")
	}
	return fmt.Errorf("%s: %s", code, msg)
}
