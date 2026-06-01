// Copyright (c) 2026 LealLang Contributors
// SPDX-License-Identifier: MIT

// Package checker implements semantic analysis for the LealLang compiler.
package checker

import (
	"fmt"
	"strings"

	"github.com/LealLang/leallang/internal/ast"
	"github.com/LealLang/leallang/internal/diagnostics"
	"github.com/LealLang/leallang/internal/token"
)

// Checker performs semantic analysis on an AST.
type Checker struct {
	program       *ast.Program
	diag          *diagnostics.Diagnostics
	scope         *Scope
	global        *Scope
	pkgPath       string
	inAsyncFunc   bool
	inUIFunc      bool
	componentIDs  map[string]token.Position // scoped per ui func
	windowMethods map[string]map[string]bool // windowID -> set of ui func names
}

// Check performs semantic analysis on the given program.
// Errors are reported to diag.
func Check(program *ast.Program, diag *diagnostics.Diagnostics) {
	c := &Checker{
		program:       program,
		diag:          diag,
		global:        NewScope(nil),
		scope:         nil, // set after builtins
		windowMethods: make(map[string]map[string]bool),
	}
	c.scope = c.global

	// Register built-in types, functions, and namespaces.
	RegisterBuiltins(c.global)

	// Extract package path.
	if program.Package != nil {
		c.pkgPath = joinPath(program.Package.Path)
	}

	// Pass 1: collect top-level declarations.
	c.collectDecls(program)

	// Pass 2: check bodies.
	c.checkDecls(program)
}

// collectDecls is Pass 1: register all top-level names without checking bodies.
func (c *Checker) collectDecls(program *ast.Program) {
	for _, decl := range program.Decls {
		switch d := decl.(type) {
		case *ast.TypeDecl:
			c.collectTypeDecl(d)
		case *ast.FuncDecl:
			c.collectFuncDecl(d)
		case *ast.VarDecl:
			c.collectVarDecl(d)
		case *ast.ConstDecl:
			c.collectConstDecl(d)
		case *ast.ComponentDecl:
			c.collectComponentDecls(d)
		}
	}
}

// collectComponentDecls collects ui func declarations from a component tree.
func (c *Checker) collectComponentDecls(decl *ast.ComponentDecl) {
	if decl.Component == "Window" && len(decl.Funcs) > 0 {
		methods := make(map[string]bool)
		for _, fn := range decl.Funcs {
			c.collectFuncDecl(fn)
			methods[fn.Name] = true
		}
		c.windowMethods[decl.ID] = methods
	} else {
		for _, fn := range decl.Funcs {
			c.collectFuncDecl(fn)
		}
	}
	for _, child := range decl.Children {
		c.collectComponentDecls(child)
	}
}

// checkDecls is Pass 2: check bodies of all declarations.
func (c *Checker) checkDecls(program *ast.Program) {
	for _, decl := range program.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			c.checkFuncDecl(d)
		case *ast.VarDecl:
			c.checkVarDecl(d)
		case *ast.ConstDecl:
			c.checkConstDecl(d)
		case *ast.TypeDecl:
			c.checkTypeDecl(d)
		case *ast.ComponentDecl:
			c.checkComponentDeclTopLevel(d)
		}
	}
}

// checkComponentDeclTopLevel checks a top-level component declaration (e.g., Window).
func (c *Checker) checkComponentDeclTopLevel(decl *ast.ComponentDecl) {
	// Only Window is allowed at the top level.
	if decl.Component != "Window" {
		c.error(decl.CompPos, 0, "E076", "component declaration outside ui function", "only Window[...] is allowed at the top level; other components must be inside ui func bodies")
		return
	}

	// Initialize component ID tracking for the Window's tree.
	prevIDs := c.componentIDs
	c.componentIDs = make(map[string]token.Position)
	defer func() { c.componentIDs = prevIDs }()

	// Register the Window's own ID.
	c.componentIDs[decl.ID] = decl.CompPos

	// Look up the component type.
	sym := c.global.Lookup(decl.Component)
	if sym == nil {
		c.error(decl.CompPos, 0, "E043", fmt.Sprintf("unknown component type '%s'", decl.Component), "")
		return
	}
	ct, ok := sym.Type.(*ComponentType)
	if !ok {
		c.error(decl.CompPos, 0, "E043", fmt.Sprintf("'%s' is not a component type", decl.Component), "")
		return
	}

	// Validate properties.
	for _, prop := range decl.Props {
		propInfo, exists := ct.Properties[prop.Name]
		if !exists {
			if ct.Primary != nil && ct.Primary.Name == prop.Name {
				propInfo = ct.Primary
			} else {
				c.error(prop.PropPos, 0, "E072",
					fmt.Sprintf("unknown property '%s' for component %s", prop.Name, decl.Component),
					fmt.Sprintf("valid properties: %s", propertyNames(ct.Properties)))
				continue
			}
		}
		valType := c.checkExpr(prop.Value)
		if valType != nil && !IsAssignable(valType, propInfo.Type) {
			c.error(prop.Value.Pos(), 0, "E022",
				fmt.Sprintf("cannot assign %s to property '%s' of type %s", FormatType(valType), prop.Name, FormatType(propInfo.Type)),
				"")
		}
	}

	// Validate event bindings.
	for _, ev := range decl.Events {
		eventType, exists := ct.Events[ev.Event]
		if !exists {
			c.error(ev.OnPos, 0, "E073",
				fmt.Sprintf("unknown event '%s' for component %s", ev.Event, decl.Component),
				fmt.Sprintf("valid events: %s", eventNames(ct.Events)))
			continue
		}
		handlerIdent, ok := ev.Handler.(*ast.Ident)
		if !ok {
			c.error(ev.Handler.Pos(), 0, "E074", "event handler must be a function name", "")
			continue
		}
		handlerSym := c.scope.Lookup(handlerIdent.Name)
		if handlerSym == nil {
			c.error(ev.Handler.Pos(), 0, "E021", fmt.Sprintf("undefined name '%s'", handlerIdent.Name), "")
			continue
		}
		handlerSig, ok := handlerSym.Type.(*FuncSignature)
		if !ok {
			c.error(ev.Handler.Pos(), 0, "E074",
				fmt.Sprintf("'%s' is not a function", handlerIdent.Name), "")
			continue
		}
		if len(handlerSig.Params) > 1 {
			c.error(ev.Handler.Pos(), 0, "E074",
				fmt.Sprintf("event handler '%s' has too many parameters (expected 0 or 1)", handlerIdent.Name), "")
			continue
		}
		if len(handlerSig.Params) == 1 {
			paramType := handlerSig.Params[0].Type
			if !IsAssignable(eventType, paramType) {
				c.error(ev.Handler.Pos(), 0, "E074",
					fmt.Sprintf("event handler '%s' parameter type %s does not match event type %s",
						handlerIdent.Name, FormatType(paramType), FormatType(eventType)), "")
			}
		}
	}

	// Check ui func declarations inside the Window.
	for _, fn := range decl.Funcs {
		c.checkFuncDecl(fn)
	}

	// Recurse into children (they are inside the Window's component tree, which is a UI context).
	for _, child := range decl.Children {
		c.checkComponentDeclInWindow(child)
	}
}

// checkComponentDeclInWindow checks a component declaration inside a Window's tree.
// These components don't need to be inside a ui func — they are part of the Window's declarative tree.
func (c *Checker) checkComponentDeclInWindow(decl *ast.ComponentDecl) {
	// Check for duplicate component ID.
	if existing, exists := c.componentIDs[decl.ID]; exists {
		c.error(decl.CompPos, 0, "E070",
			fmt.Sprintf("duplicate component id '%s' in this window (first used at line %d)", decl.ID, existing.Line),
			"component ids must be unique within a window")
		return
	}
	c.componentIDs[decl.ID] = decl.CompPos

	sym := c.global.Lookup(decl.Component)
	if sym == nil {
		c.error(decl.CompPos, 0, "E043", fmt.Sprintf("unknown component type '%s'", decl.Component), "")
		return
	}
	ct, ok := sym.Type.(*ComponentType)
	if !ok {
		c.error(decl.CompPos, 0, "E043", fmt.Sprintf("'%s' is not a component type", decl.Component), "")
		return
	}

	for _, prop := range decl.Props {
		propInfo, exists := ct.Properties[prop.Name]
		if !exists {
			if ct.Primary != nil && ct.Primary.Name == prop.Name {
				propInfo = ct.Primary
			} else {
				c.error(prop.PropPos, 0, "E072",
					fmt.Sprintf("unknown property '%s' for component %s", prop.Name, decl.Component),
					fmt.Sprintf("valid properties: %s", propertyNames(ct.Properties)))
				continue
			}
		}
		valType := c.checkExpr(prop.Value)
		if valType != nil && !IsAssignable(valType, propInfo.Type) {
			c.error(prop.Value.Pos(), 0, "E022",
				fmt.Sprintf("cannot assign %s to property '%s' of type %s", FormatType(valType), prop.Name, FormatType(propInfo.Type)),
				"")
		}
	}

	for _, ev := range decl.Events {
		eventType, exists := ct.Events[ev.Event]
		if !exists {
			c.error(ev.OnPos, 0, "E073",
				fmt.Sprintf("unknown event '%s' for component %s", ev.Event, decl.Component),
				fmt.Sprintf("valid events: %s", eventNames(ct.Events)))
			continue
		}
		handlerIdent, ok := ev.Handler.(*ast.Ident)
		if !ok {
			c.error(ev.Handler.Pos(), 0, "E074", "event handler must be a function name", "")
			continue
		}
		handlerSym := c.scope.Lookup(handlerIdent.Name)
		if handlerSym == nil {
			c.error(ev.Handler.Pos(), 0, "E021", fmt.Sprintf("undefined name '%s'", handlerIdent.Name), "")
			continue
		}
		handlerSig, ok := handlerSym.Type.(*FuncSignature)
		if !ok {
			c.error(ev.Handler.Pos(), 0, "E074",
				fmt.Sprintf("'%s' is not a function", handlerIdent.Name), "")
			continue
		}
		if len(handlerSig.Params) > 1 {
			c.error(ev.Handler.Pos(), 0, "E074",
				fmt.Sprintf("event handler '%s' has too many parameters (expected 0 or 1)", handlerIdent.Name), "")
			continue
		}
		if len(handlerSig.Params) == 1 {
			paramType := handlerSig.Params[0].Type
			if !IsAssignable(eventType, paramType) {
				c.error(ev.Handler.Pos(), 0, "E074",
					fmt.Sprintf("event handler '%s' parameter type %s does not match event type %s",
						handlerIdent.Name, FormatType(paramType), FormatType(eventType)), "")
			}
		}
	}

	for _, child := range decl.Children {
		c.checkComponentDeclInWindow(child)
	}
}

// collectTypeDecl registers a type name in the global scope (Pass 1).
func (c *Checker) collectTypeDecl(d *ast.TypeDecl) {
	record := &RecordType{
		Name:    d.Name,
		Extends: d.Extends,
		Fields:  make([]*FieldInfo, len(d.Fields)),
		Methods: make(map[string]*FuncSignature),
	}

	// Resolve field types.
	for i, f := range d.Fields {
		fieldType := c.resolveTypeExpr(f.Type)
		record.Fields[i] = &FieldInfo{
			Name:    f.Name,
			Type:    fieldType,
			Default: f.Default != nil,
			Pub:     f.Pub,
		}
	}

	// Collect methods.
	for _, m := range d.Methods {
		sig := c.funcSignature(m)
		record.Methods[m.Name] = sig
	}

	// Collect constructor.
	if d.Constructor != nil {
		ctor := &FuncSignature{
			Name:       d.Name,
			Params:     c.paramInfos(d.Constructor.Params),
			ReturnType: nil, // constructors return void
			Pub:        true,
		}
		record.Ctor = ctor
	} else {
		// Generate default constructor from fields.
		defaultParams := make([]*ParamInfo, len(record.Fields))
		for i, f := range record.Fields {
			fieldType := f.Type
			if f.Default {
				fieldType = &NullableType{Inner: fieldType}
			}
			defaultParams[i] = &ParamInfo{
				Name: f.Name,
				Type: fieldType,
			}
		}
		record.Ctor = &FuncSignature{
			Name:       d.Name,
			Params:     defaultParams,
			ReturnType: nil,
			Pub:        true,
		}
	}

	sym := &Symbol{
		Name: d.Name,
		Type: record,
		Kind: SymType,
		Pos:  d.TypePos,
		Pub:  true, // types are always pub for now
	}
	if dup := c.global.Define(sym); dup != "" {
		c.error(d.TypePos, 0, "E050", fmt.Sprintf("duplicate name '%s' in this scope", dup), "")
	}
}

// collectFuncDecl registers a function name in the global scope (Pass 1).
func (c *Checker) collectFuncDecl(d *ast.FuncDecl) {
	if d.Async && d.UI {
		c.error(d.FuncPos, 0, "E060", "function cannot be both async and ui", "choose either async or ui")
	}
	sig := c.funcSignature(d)
	sym := &Symbol{
		Name: d.Name,
		Type: sig,
		Kind: SymFunc,
		Pos:  d.FuncPos,
		Pub:  d.Pub,
	}
	if dup := c.global.Define(sym); dup != "" {
		c.error(d.FuncPos, 0, "E050", fmt.Sprintf("duplicate name '%s' in this scope", dup), "")
	}
}

// collectVarDecl registers a variable name in the global scope (Pass 1).
func (c *Checker) collectVarDecl(d *ast.VarDecl) {
	varType := c.resolveTypeExpr(d.Type)
	sym := &Symbol{
		Name: d.Name,
		Type: varType,
		Kind: SymVar,
		Pos:  d.NamePos,
	}
	if dup := c.global.Define(sym); dup != "" {
		c.error(d.NamePos, 0, "E050", fmt.Sprintf("duplicate name '%s' in this scope", dup), "")
	}
}

// collectConstDecl registers a const name in the global scope (Pass 1).
func (c *Checker) collectConstDecl(d *ast.ConstDecl) {
	constType := c.resolveTypeExpr(d.Type)
	sym := &Symbol{
		Name: d.Name,
		Type: constType,
		Kind: SymConst,
		Pos:  d.ConstPos,
	}
	if dup := c.global.Define(sym); dup != "" {
		c.error(d.ConstPos, 0, "E050", fmt.Sprintf("duplicate name '%s' in this scope", dup), "")
	}
}

// checkFuncDecl checks a function's body (Pass 2).
func (c *Checker) checkFuncDecl(d *ast.FuncDecl) {
	sym := c.global.Lookup(d.Name)
	if sym == nil {
		return
	}
	sig, ok := sym.Type.(*FuncSignature)
	if !ok {
		return
	}

	// Create function scope.
	fnScope := NewScope(c.scope)
	prevScope := c.scope
	c.scope = fnScope

	// Check async-safety: no ref params in async func.
	if d.Async {
		for _, p := range d.Params {
			if p.Ref {
				c.error(p.Posn, 0, "E063", "ref parameter not allowed in async function", "async functions cannot have ref parameters")
			}
		}
	}

	// Define parameters in function scope.
	for _, p := range d.Params {
		paramType := c.resolveTypeExpr(p.Type)
		paramSym := &Symbol{
			Name: p.Name,
			Type: paramType,
			Kind: SymParam,
			Pos:  p.Posn,
			Ref:  p.Ref,
		}
		fnScope.Define(paramSym)
	}

	// Set async context for body checking.
	prevAsync := c.inAsyncFunc
	c.inAsyncFunc = d.Async

	// Set UI context for body checking.
	prevUI := c.inUIFunc
	prevIDs := c.componentIDs
	if d.UI {
		c.inUIFunc = true
		c.componentIDs = make(map[string]token.Position)
	}

	// Check body.
	for _, stmt := range d.Body {
		c.checkStmt(stmt, sig.ReturnType)
	}

	c.inAsyncFunc = prevAsync
	c.inUIFunc = prevUI
	c.componentIDs = prevIDs
	c.scope = prevScope
}

// checkVarDecl checks a variable's initializer (Pass 2).
func (c *Checker) checkVarDecl(d *ast.VarDecl) {
	varType := c.resolveTypeExpr(d.Type)

	// Define in current scope if not already defined (e.g., local variables in function bodies).
	sym := c.scope.LookupLocal(d.Name)
	if sym == nil {
		sym = &Symbol{
			Name: d.Name,
			Type: varType,
			Kind: SymVar,
			Pos:  d.NamePos,
		}
		if dup := c.scope.Define(sym); dup != "" {
			c.error(d.NamePos, 0, "E050", fmt.Sprintf("duplicate name '%s' in this scope", dup), "")
			return
		}
	}

	if d.Value != nil {
		valType := c.checkExpr(d.Value)
		c.checkAssignment(valType, sym.Type, d.Value.Pos(), d.Value.End().Col-d.Value.Pos().Col)
		if sym.Type == nil && valType != nil {
			sym.Type = valType
		}
	}
}

// checkConstDecl checks a const's initializer (Pass 2).
func (c *Checker) checkConstDecl(d *ast.ConstDecl) {
	constType := c.resolveTypeExpr(d.Type)

	// Define in current scope if not already defined (e.g., local constants in function bodies).
	sym := c.scope.LookupLocal(d.Name)
	if sym == nil {
		sym = &Symbol{
			Name: d.Name,
			Type: constType,
			Kind: SymConst,
			Pos:  d.ConstPos,
		}
		if dup := c.scope.Define(sym); dup != "" {
			c.error(d.ConstPos, 0, "E050", fmt.Sprintf("duplicate name '%s' in this scope", dup), "")
			return
		}
	}

	if d.Value != nil {
		valType := c.checkExpr(d.Value)
		c.checkAssignment(valType, sym.Type, d.Value.Pos(), d.Value.End().Col-d.Value.Pos().Col)
		if sym.Type == nil && valType != nil {
			sym.Type = valType
		}
	}
}

// checkTypeDecl checks a type's constructor and method bodies (Pass 2).
func (c *Checker) checkTypeDecl(d *ast.TypeDecl) {
	sym := c.global.Lookup(d.Name)
	if sym == nil {
		return
	}
	record, ok := sym.Type.(*RecordType)
	if !ok {
		return
	}

	// Check constructor body.
	if d.Constructor != nil {
		ctorScope := NewScope(c.scope)
		prevScope := c.scope
		c.scope = ctorScope

		// Define constructor params.
		for _, p := range d.Constructor.Params {
			paramType := c.resolveTypeExpr(p.Type)
			ctorScope.Define(&Symbol{
				Name: p.Name,
				Type: paramType,
				Kind: SymParam,
				Pos:  p.Posn,
				Ref:  p.Ref,
			})
		}

		// Define record fields so they can be assigned in the constructor body.
		for _, field := range record.Fields {
			ctorScope.Define(&Symbol{
				Name: field.Name,
				Type: field.Type,
				Kind: SymVar,
				Pos:  d.TypePos,
			})
		}

		for _, stmt := range d.Constructor.Body {
			c.checkStmt(stmt, nil)
		}

		c.scope = prevScope
	}

	// Check method bodies.
	for _, m := range d.Methods {
		sig := record.Methods[m.Name]
		if sig == nil {
			continue
		}
		methodScope := NewScope(c.scope)
		prevScope := c.scope
		c.scope = methodScope

		// Define method params.
		for _, p := range m.Params {
			paramType := c.resolveTypeExpr(p.Type)
			methodScope.Define(&Symbol{
				Name: p.Name,
				Type: paramType,
				Kind: SymParam,
				Pos:  p.Posn,
				Ref:  p.Ref,
			})
		}

		// Define record fields so they can be accessed in methods.
		for _, field := range record.Fields {
			methodScope.Define(&Symbol{
				Name: field.Name,
				Type: field.Type,
				Kind: SymVar,
				Pos:  d.TypePos,
			})
		}

		for _, stmt := range m.Body {
			c.checkStmt(stmt, sig.ReturnType)
		}

		c.scope = prevScope
	}
}

// funcSignature builds a FuncSignature from an AST FuncDecl.
func (c *Checker) funcSignature(d *ast.FuncDecl) *FuncSignature {
	sig := &FuncSignature{
		Name:   d.Name,
		Params: c.paramInfos(d.Params),
		Pub:    d.Pub,
		Async:  d.Async,
		UI:     d.UI,
	}
	if len(d.ReturnTypes) == 1 {
		sig.ReturnType = c.resolveTypeExpr(d.ReturnTypes[0])
	} else if len(d.ReturnTypes) > 1 {
		elems := make([]Type, len(d.ReturnTypes))
		for i, rt := range d.ReturnTypes {
			elems[i] = c.resolveTypeExpr(rt)
		}
		sig.ReturnType = &TupleType{Elements: elems}
	}
	return sig
}

// paramInfos converts AST params to ParamInfo slice.
func (c *Checker) paramInfos(params []*ast.Param) []*ParamInfo {
	infos := make([]*ParamInfo, len(params))
	for i, p := range params {
		infos[i] = &ParamInfo{
			Name: p.Name,
			Type: c.resolveTypeExpr(p.Type),
			Ref:  p.Ref,
		}
	}
	return infos
}

// resolveTypeExpr converts an ast.TypeExpr to a Type.
func (c *Checker) resolveTypeExpr(t ast.TypeExpr) Type {
	if t == nil {
		return nil
	}
	switch n := t.(type) {
	case *ast.SimpleType:
		return c.resolveSimpleType(n.Name)
	case *ast.NullableType:
		inner := c.resolveTypeExpr(n.Inner)
		if inner == nil {
			return nil
		}
		return &NullableType{Inner: inner}
	case *ast.GenericType:
		params := make([]Type, len(n.Params))
		for i, p := range n.Params {
			params[i] = c.resolveTypeExpr(p)
		}
		return &GenericType{Name: n.Name, Params: params}
	default:
		return nil
	}
}

// resolveSimpleType resolves a simple type name to a Type.
func (c *Checker) resolveSimpleType(name string) Type {
	if prim := primitiveByName(name); prim != nil {
		return prim
	}
	if sym := c.scope.Lookup(name); sym != nil && sym.Kind == SymType {
		return sym.Type
	}
	// Try global scope directly for builtins.
	if sym := c.global.Lookup(name); sym != nil && sym.Kind == SymType {
		return sym.Type
	}
	return nil
}

// checkExpr infers the type of an expression. Returns nil if unknown.
func (c *Checker) checkExpr(expr ast.Expr) Type {
	if expr == nil {
		return nil
	}
	c.checkAsyncSafety(expr)
	switch e := expr.(type) {
	case *ast.Literal:
		return c.checkLiteral(e)
	case *ast.Ident:
		return c.checkIdent(e)
	case *ast.BinaryExpr:
		return c.checkBinaryExpr(e)
	case *ast.UnaryExpr:
		return c.checkUnaryExpr(e)
	case *ast.AwaitExpr:
		return c.checkAwaitExpr(e)
	case *ast.CallExpr:
		return c.checkCallExpr(e)
	case *ast.FieldExpr:
		return c.checkFieldExpr(e)
	case *ast.IndexExpr:
		return c.checkIndexExpr(e)
	case *ast.ListLiteral:
		return c.checkListLiteral(e)
	case *ast.DictLiteral:
		return c.checkDictLiteral(e)
	case *ast.RangeExpr:
		return c.checkRangeExpr(e)
	case *ast.InterpStringExpr:
		return c.checkInterpString(e)
	case *ast.SwitchExpr:
		return c.checkSwitchExpr(e)
	case *ast.ComponentRefExpr:
		return c.checkComponentRef(e)
	case *ast.BadExpr:
		return nil
	default:
		return nil
	}
}

// checkAsyncSafety checks async-safety rules for expressions inside async functions.
func (c *Checker) checkAsyncSafety(expr ast.Expr) {
	if !c.inAsyncFunc {
		return
	}
	switch e := expr.(type) {
	case *ast.ComponentRefExpr:
		c.error(e.AtPos, 0, "E061", "UI component reference not allowed in async function", "async functions cannot access UI components")
	case *ast.FieldExpr:
		if ident, ok := e.X.(*ast.Ident); ok {
			if isUIAffineNamespace(ident.Name) {
				c.error(e.Dot, 0, "E062", fmt.Sprintf("namespace '%s' not allowed in async function", ident.Name), "async functions cannot use UI-affine namespaces")
			}
		}
	}
}

func isUIAffineNamespace(name string) bool {
	switch name {
	case "window", "msg", "modal", "toast":
		return true
	default:
		return false
	}
}

// checkAwaitExpr checks an await expression and returns the unwrapped type.
// await yields (T, Error?) for Task<T>, or Error? for void Task.
func (c *Checker) checkAwaitExpr(a *ast.AwaitExpr) Type {
	innerType := c.checkExpr(a.X)
	if innerType == nil {
		return nil
	}
	if gt, ok := innerType.(*GenericType); ok && gt.Name == "Task" {
		if len(gt.Params) == 1 {
			payload := gt.Params[0]
			errType := &NullableType{Inner: c.global.Lookup("Error").Type}
			if _, ok := payload.(*Void); ok {
				// async func -> Error? only: await yields Error?
				return errType
			}
			// async func -> T, Error?: await yields (T, Error?)
			return &TupleType{Elements: []Type{payload, errType}}
		}
		return nil
	}
	c.error(a.AwaitPos, 0, "E064", fmt.Sprintf("cannot await non-task type %s", FormatType(innerType)), "await can only be used on Task values")
	return nil
}

// checkLiteral returns the type of a literal.
func (c *Checker) checkLiteral(l *ast.Literal) Type {
	switch l.Kind {
	case token.INT_LIT:
		return IntType
	case token.FLOAT_LIT:
		return FloatType
	case token.STRING_LIT:
		return StringType
	case token.CHAR_LIT:
		return CharType
	case token.TRUE, token.FALSE:
		return BoolType
	case token.NULL:
		return NullType_
	default:
		return nil
	}
}

// checkIdent looks up an identifier in scope.
func (c *Checker) checkIdent(id *ast.Ident) Type {
	if id.Name == "_" {
		return nil // discards are not typed
	}
	sym := c.scope.Lookup(id.Name)
	if sym == nil {
		c.error(id.NamePos, 0, "E021", fmt.Sprintf("undefined name '%s'", id.Name), "declare the name before using it")
		return nil
	}
	return sym.Type
}

// checkBinaryExpr checks a binary expression and returns its result type.
func (c *Checker) checkBinaryExpr(b *ast.BinaryExpr) Type {
	left := c.checkExpr(b.Left)
	right := c.checkExpr(b.Right)

	switch b.Op {
	case token.PLUS:
		// string concatenation or arithmetic — both sides must be the same kind
		if isString(left) && isString(right) {
			return StringType
		}
		if isNumeric(left) && isNumeric(right) {
			return numericResult(left, right)
		}
		c.error(b.OpPos, 0, "E025", fmt.Sprintf("operator + not defined on %s and %s", FormatType(left), FormatType(right)), "")
		return nil
	case token.MINUS, token.STAR, token.SLASH, token.PERCENT:
		if isNumeric(left) && isNumeric(right) {
			return numericResult(left, right)
		}
		c.error(b.OpPos, 0, "E025", fmt.Sprintf("operator %s not defined on %s and %s", b.Op, FormatType(left), FormatType(right)), "")
		return nil
	case token.EQ_EQ, token.BANG_EQ:
		return BoolType
	case token.LESS, token.LESS_EQ, token.GREATER, token.GREATER_EQ:
		if isNumeric(left) && isNumeric(right) {
			return BoolType
		}
		c.error(b.OpPos, 0, "E053", fmt.Sprintf("comparison operator %s not defined between %s and %s", b.Op, FormatType(left), FormatType(right)), "")
		return BoolType
	case token.AND, token.OR:
		if isBool(left) && isBool(right) {
			return BoolType
		}
		c.error(b.OpPos, 0, "E054", fmt.Sprintf("logical operator %s requires bool operands, got %s and %s", b.Op, FormatType(left), FormatType(right)), "")
		return BoolType
	default:
		return nil
	}
}

// checkUnaryExpr checks a unary expression.
func (c *Checker) checkUnaryExpr(u *ast.UnaryExpr) Type {
	operand := c.checkExpr(u.X)
	switch u.Op {
	case token.MINUS:
		if isNumeric(operand) {
			return operand
		}
		c.error(u.OpPos, 0, "E055", fmt.Sprintf("unary - not defined on %s", FormatType(operand)), "")
		return nil
	case token.NOT:
		if isBool(operand) {
			return BoolType
		}
		c.error(u.OpPos, 0, "E055", fmt.Sprintf("unary not requires bool, got %s", FormatType(operand)), "")
		return nil
	default:
		return nil
	}
}

// checkCallExpr checks a function call.
func (c *Checker) checkCallExpr(call *ast.CallExpr) Type {
	// Resolve the callee.
	var sig *FuncSignature
	var calleeType Type

	var resultType Type // for constructors, this is the record type

	switch fn := call.Func.(type) {
	case *ast.Ident:
		sym := c.scope.Lookup(fn.Name)
		if sym == nil {
			c.error(fn.NamePos, 0, "E021", fmt.Sprintf("undefined name '%s'", fn.Name), "")
			return nil
		}
		calleeType = sym.Type
		if s, ok := sym.Type.(*FuncSignature); ok {
			sig = s
		} else if r, ok := sym.Type.(*RecordType); ok && r.Ctor != nil {
			sig = r.Ctor
			resultType = r // record construction returns the record type
		}
	case *ast.FieldExpr:
		// Namespace call, record method call, or component property access
		c.checkAsyncSafety(fn)

		// Handle cross-window calls: @Window[id].ui_func(args)
		// This is allowed even outside ui func bodies.
		if compRef, ok := fn.X.(*ast.ComponentRefExpr); ok {
			if compRef.Component == "Window" {
				if methods, exists := c.windowMethods[compRef.ID]; exists && methods[fn.Field] {
					sym := c.global.Lookup(fn.Field)
					if sym != nil {
						if s, ok := sym.Type.(*FuncSignature); ok {
							sig = s
						}
					}
					if sig == nil {
						c.error(fn.Dot, 0, "E039", fmt.Sprintf("window '%s' has no ui func '%s'", compRef.ID, fn.Field), "")
						return nil
					}
					break
				}
			}
		}

		baseType := c.checkExpr(fn.X)
		switch t := baseType.(type) {
		case *NamespaceType:
			if member, exists := t.Members[fn.Field]; exists {
				sig = member
			} else {
				c.error(fn.Dot, 0, "E039", fmt.Sprintf("type %s has no field '%s'", t.Name, fn.Field), "")
				return nil
			}
		case *RecordType:
			if method, ok := t.Methods[fn.Field]; ok {
				sig = method
			} else {
				c.error(fn.Dot, 0, "E039", fmt.Sprintf("type %s has no method '%s'", t.Name, fn.Field), "")
				return nil
			}
		case *ComponentType:
			if prop, ok := t.Properties[fn.Field]; ok {
				calleeType = prop.Type
			} else if t.Primary != nil && t.Primary.Name == fn.Field {
				calleeType = t.Primary.Type
			} else {
				calleeType = baseType
			}
		default:
			calleeType = baseType
		}
	default:
		calleeType = c.checkExpr(call.Func)
		if s, ok := calleeType.(*FuncSignature); ok {
			sig = s
		}
	}

	if sig == nil {
		if calleeType != nil {
			c.error(call.LParen, 0, "E049", fmt.Sprintf("cannot call non-function type %s", FormatType(calleeType)), "")
		}
		return nil
	}

	// Validate arguments.
	c.checkCallArgs(sig, call)

	// Constructors return the record type, not void.
	if resultType != nil {
		return resultType
	}
	// Async functions wrap their return type in Task<T>.
	if sig.Async {
		payload := asyncPayloadType(sig.ReturnType)
		return &GenericType{Name: "Task", Params: []Type{payload}}
	}
	return sig.ReturnType
}

// asyncPayloadType extracts the task payload from an async function's return type.
// If the return type is a TupleType ending with Error?, it strips the Error? and
// wraps the remaining types. If the return type is only Error?, returns Void.
// Otherwise returns the return type directly.
func asyncPayloadType(ret Type) Type {
	if ret == nil {
		return &Void{}
	}
	if tuple, ok := ret.(*TupleType); ok && len(tuple.Elements) > 0 {
		last := tuple.Elements[len(tuple.Elements)-1]
		if nullable, ok := last.(*NullableType); ok {
			if rec, isRec := nullable.Inner.(*RecordType); isRec && rec.Name == "Error" {
				// Last element is Error? -- strip it.
				if len(tuple.Elements) == 1 {
					return &Void{} // only Error? return
				}
				if len(tuple.Elements) == 2 {
					return tuple.Elements[0] // single value + Error?
				}
				return &TupleType{Elements: tuple.Elements[:len(tuple.Elements)-1]}
			}
		}
	}
	if nullable, ok := ret.(*NullableType); ok {
		if rec, isRec := nullable.Inner.(*RecordType); isRec && rec.Name == "Error" {
			return &Void{} // only Error? return
		}
	}
	return ret
}

// checkCallArgs validates call arguments against a function signature.
func (c *Checker) checkCallArgs(sig *FuncSignature, call *ast.CallExpr) {
	// Separate positional and named args.
	var positional []ast.Expr
	named := make(map[string]ast.Expr)
	namedOrder := make([]string, 0)

	for _, arg := range call.Args {
		if arg.Name != "" {
			if _, exists := named[arg.Name]; exists {
				c.error(arg.Posn, 0, "E028", fmt.Sprintf("duplicate named argument '%s'", arg.Name), "")
			}
			named[arg.Name] = arg.Value
			namedOrder = append(namedOrder, arg.Name)
		} else {
			if len(named) > 0 {
				c.error(arg.Posn, 0, "E029", "positional argument not allowed after named argument", "put positional arguments before named arguments")
			}
			positional = append(positional, arg.Value)
		}
	}

	// Check positional args.
	for i, arg := range positional {
		if i >= len(sig.Params) {
			c.error(call.RParen, 0, "E026", fmt.Sprintf("expected %d arguments, got %d", len(sig.Params), len(call.Args)), "")
			return
		}
		argType := c.checkExpr(arg)
		paramType := sig.Params[i].Type
		if argType != nil && paramType != nil && !IsAssignable(argType, paramType) {
			c.error(arg.Pos(), arg.End().Col-arg.Pos().Col, "E022", fmt.Sprintf("cannot use %s as %s", FormatType(argType), FormatType(paramType)), "")
		}
		if sig.Params[i].Ref {
			if _, ok := arg.(*ast.Ident); !ok {
				c.error(arg.Pos(), 0, "E034", fmt.Sprintf("ref parameter '%s' requires a variable", sig.Params[i].Name), "")
			}
		}
	}

	// Check named args.
	for _, name := range namedOrder {
		argVal := named[name]
		found := false
		for _, param := range sig.Params {
			if param.Name == name {
				found = true
				argType := c.checkExpr(argVal)
				if argType != nil && param.Type != nil && !IsAssignable(argType, param.Type) {
					c.error(argVal.Pos(), argVal.End().Col-argVal.Pos().Col, "E022", fmt.Sprintf("cannot use %s as %s", FormatType(argType), FormatType(param.Type)), "")
				}
				if param.Ref {
					if _, ok := argVal.(*ast.Ident); !ok {
						c.error(argVal.Pos(), 0, "E034", fmt.Sprintf("ref parameter '%s' requires a variable", param.Name), "")
					}
				}
				break
			}
		}
		if !found {
			c.error(call.LParen, 0, "E027", fmt.Sprintf("unknown named argument '%s'", name), "")
		}
	}

	// Check that all required params are provided.
	provided := len(positional) + len(named)
	if provided < len(sig.Params) {
		// Find the first missing param.
		for i := len(positional); i < len(sig.Params); i++ {
			if _, ok := named[sig.Params[i].Name]; !ok {
				c.error(call.LParen, 0, "E026", fmt.Sprintf("expected %d arguments, got %d", len(sig.Params), provided), "")
				return
			}
		}
	}
}

// checkFieldExpr checks obj.field.
func (c *Checker) checkFieldExpr(f *ast.FieldExpr) Type {
	// Handle cross-window field access: @Window[id].field
	// This is allowed even outside ui func bodies.
	if compRef, ok := f.X.(*ast.ComponentRefExpr); ok {
		if compRef.Component == "Window" {
			if sym := c.global.Lookup(compRef.Component); sym != nil {
				if ct, ok := sym.Type.(*ComponentType); ok {
					if prop, ok := ct.Properties[f.Field]; ok {
						return prop.Type
					}
					if ct.Primary != nil && ct.Primary.Name == f.Field {
						return ct.Primary.Type
					}
					// Check if it's a ui func method.
					if methods, exists := c.windowMethods[compRef.ID]; exists && methods[f.Field] {
						methodSym := c.global.Lookup(f.Field)
						if methodSym != nil {
							return methodSym.Type
						}
					}
					c.error(f.Dot, 0, "E039", fmt.Sprintf("type %s has no field '%s'", compRef.Component, f.Field), "")
					return nil
				}
			}
		}
	}

	base := c.checkExpr(f.X)
	if base == nil {
		return nil
	}

	switch t := base.(type) {
	case *RecordType:
		for _, field := range t.Fields {
			if field.Name == f.Field {
				return field.Type
			}
		}
		// Check methods.
		if method, ok := t.Methods[f.Field]; ok {
			return method
		}
		c.error(f.Dot, 0, "E039", fmt.Sprintf("type %s has no field '%s'", t.Name, f.Field), "")
		return nil
	case *NamespaceType:
		if member, ok := t.Members[f.Field]; ok {
			return member
		}
		if ct, ok := t.Consts[f.Field]; ok {
			return ct
		}
		c.error(f.Dot, 0, "E039", fmt.Sprintf("type %s has no field '%s'", t.Name, f.Field), "")
		return nil
	case *ComponentType:
		if prop, ok := t.Properties[f.Field]; ok {
			return prop.Type
		}
		if t.Primary != nil && t.Primary.Name == f.Field {
			return t.Primary.Type
		}
		c.error(f.Dot, 0, "E039", fmt.Sprintf("type %s has no field '%s'", t.Name, f.Field), "")
		return nil
	default:
		return nil
	}
}

// checkIndexExpr checks obj[index].
func (c *Checker) checkIndexExpr(idx *ast.IndexExpr) Type {
	base := c.checkExpr(idx.X)
	indexType := c.checkExpr(idx.Index)

	switch t := base.(type) {
	case *GenericType:
		switch t.Name {
		case "List":
			if len(t.Params) == 1 {
				if indexType != nil && !IsAssignable(indexType, IntType) {
					c.error(idx.LBrack, 0, "E041", fmt.Sprintf("index must be int, got %s", FormatType(indexType)), "")
				}
				return t.Params[0]
			}
		case "Dict":
			if len(t.Params) == 2 {
				if indexType != nil && !IsAssignable(indexType, t.Params[0]) {
					c.error(idx.LBrack, 0, "E041", fmt.Sprintf("index must be %s, got %s", FormatType(t.Params[0]), FormatType(indexType)), "")
				}
				return t.Params[1]
			}
		}
	case *PrimitiveType:
		if t.Name == "string" {
			if indexType != nil && !IsAssignable(indexType, IntType) {
				c.error(idx.LBrack, 0, "E041", fmt.Sprintf("index must be int, got %s", FormatType(indexType)), "")
			}
			return CharType
		}
	}

	if base != nil {
		c.error(idx.LBrack, 0, "E040", fmt.Sprintf("type %s is not indexable", FormatType(base)), "")
	}
	return nil
}

// checkListLiteral infers the type of a list literal.
func (c *Checker) checkListLiteral(l *ast.ListLiteral) Type {
	if len(l.Elements) == 0 {
		return &GenericType{Name: "List", Params: []Type{AnyType}}
	}
	elemType := c.checkExpr(l.Elements[0])
	for _, el := range l.Elements[1:] {
		t := c.checkExpr(el)
		if elemType != nil && t != nil && !IsAssignable(t, elemType) {
			c.error(el.Pos(), el.End().Col-el.Pos().Col, "E022", fmt.Sprintf("list element type mismatch: expected %s, got %s", FormatType(elemType), FormatType(t)), "")
		}
	}
	if elemType == nil {
		elemType = AnyType
	}
	return &GenericType{Name: "List", Params: []Type{elemType}}
}

// checkDictLiteral infers the type of a dict literal.
func (c *Checker) checkDictLiteral(d *ast.DictLiteral) Type {
	if len(d.Pairs) == 0 {
		return &GenericType{Name: "Dict", Params: []Type{AnyType, AnyType}}
	}
	first := d.Pairs[0]
	keyType := c.checkExpr(first.Key)
	valType := c.checkExpr(first.Value)
	for _, pair := range d.Pairs[1:] {
		kt := c.checkExpr(pair.Key)
		vt := c.checkExpr(pair.Value)
		if keyType != nil && kt != nil && !IsAssignable(kt, keyType) {
			c.error(pair.Key.Pos(), pair.Key.End().Col-pair.Key.Pos().Col, "E022", fmt.Sprintf("dict key type mismatch: expected %s, got %s", FormatType(keyType), FormatType(kt)), "")
		}
		if valType != nil && vt != nil && !IsAssignable(vt, valType) {
			c.error(pair.Value.Pos(), pair.Value.End().Col-pair.Value.Pos().Col, "E022", fmt.Sprintf("dict value type mismatch: expected %s, got %s", FormatType(valType), FormatType(vt)), "")
		}
	}
	if keyType == nil {
		keyType = AnyType
	}
	if valType == nil {
		valType = AnyType
	}
	return &GenericType{Name: "Dict", Params: []Type{keyType, valType}}
}

// checkRangeExpr checks a range expression.
func (c *Checker) checkRangeExpr(r *ast.RangeExpr) Type {
	low := c.checkExpr(r.Low)
	high := c.checkExpr(r.High)
	if low != nil && !IsAssignable(low, IntType) {
		c.error(r.Low.Pos(), 0, "E048", fmt.Sprintf("range lower bound must be int, got %s", FormatType(low)), "")
	}
	if high != nil && !IsAssignable(high, IntType) {
		c.error(r.High.Pos(), 0, "E048", fmt.Sprintf("range upper bound must be int, got %s", FormatType(high)), "")
	}
	// Ranges are not directly typed in LealLang (used in loop context).
	return nil
}

// checkInterpString checks an interpolated string expression.
func (c *Checker) checkInterpString(i *ast.InterpStringExpr) Type {
	for _, seg := range i.Segments {
		if seg.IsExpr && seg.Expr != nil {
			c.checkExpr(seg.Expr)
		}
	}
	return StringType
}

// checkSwitchExpr checks a switch expression.
func (c *Checker) checkSwitchExpr(s *ast.SwitchExpr) Type {
	c.checkExpr(s.Subject)
	var resultType Type
	for _, arm := range s.Arms {
		for _, p := range arm.Patterns {
			c.checkExpr(p)
		}
		armType := c.checkExpr(arm.Value)
		if resultType == nil {
			resultType = armType
		} else if armType != nil && !IsAssignable(armType, resultType) {
			c.error(arm.Value.Pos(), 0, "E042",
				fmt.Sprintf("switch arms must all have the same type; got %s and %s", FormatType(resultType), FormatType(armType)), "")
		}
	}
	return resultType
}

// checkComponentRef checks a @Component[id] reference.
func (c *Checker) checkComponentRef(cr *ast.ComponentRefExpr) Type {
	// Component refs are only allowed inside ui func bodies.
	// Exception: @Window[id] is allowed as a value (e.g., window.open(@Window[id])).
	if !c.inUIFunc && cr.Component != "Window" {
		c.error(cr.AtPos, 0, "E075", "component reference outside ui func", "@Component[id] syntax is only allowed inside ui func bodies; use @Window[id].func() for cross-window calls")
		return nil
	}
	// Component refs are validated when we have a component registry.
	// For now, look up the component type by name.
	if sym := c.global.Lookup(cr.Component); sym != nil {
		if ct, ok := sym.Type.(*ComponentType); ok {
			return ct
		}
	}
	c.error(cr.AtPos, 0, "E043", fmt.Sprintf("unknown component type '%s'", cr.Component), "")
	return nil
}

// checkComponentDecl checks a UI component declaration.
func (c *Checker) checkComponentDecl(decl *ast.ComponentDecl) {
	if !c.inUIFunc {
		c.error(decl.CompPos, 0, "E076", "component declaration outside ui function", "component declarations are only allowed inside ui func bodies")
		return
	}

	// Look up the component type.
	sym := c.global.Lookup(decl.Component)
	if sym == nil {
		c.error(decl.CompPos, 0, "E043", fmt.Sprintf("unknown component type '%s'", decl.Component), "")
		return
	}
	ct, ok := sym.Type.(*ComponentType)
	if !ok {
		c.error(decl.CompPos, 0, "E043", fmt.Sprintf("'%s' is not a component type", decl.Component), "")
		return
	}

	// Check for duplicate component ID.
	if firstPos, exists := c.componentIDs[decl.ID]; exists {
		c.error(decl.CompPos, 0, "E070",
			fmt.Sprintf("duplicate component id '%s' in this scope", decl.ID),
			fmt.Sprintf("first defined at line %d", firstPos.Line))
		return
	}
	c.componentIDs[decl.ID] = decl.CompPos

	// Validate properties.
	for _, prop := range decl.Props {
		// Check if it's a regular property.
		propInfo, exists := ct.Properties[prop.Name]
		if !exists {
			// Check if it's the primary parameter.
			if ct.Primary != nil && ct.Primary.Name == prop.Name {
				propInfo = ct.Primary
			} else {
				c.error(prop.PropPos, 0, "E072",
					fmt.Sprintf("unknown property '%s' for component %s", prop.Name, decl.Component),
					fmt.Sprintf("valid properties: %s", propertyNames(ct.Properties)))
				continue
			}
		}
		valType := c.checkExpr(prop.Value)
		if valType != nil && !IsAssignable(valType, propInfo.Type) {
			c.error(prop.Value.Pos(), 0, "E022",
				fmt.Sprintf("cannot assign %s to property '%s' of type %s", FormatType(valType), prop.Name, FormatType(propInfo.Type)),
				"")
		}
	}

	// Validate event bindings.
	for _, ev := range decl.Events {
		eventType, exists := ct.Events[ev.Event]
		if !exists {
			c.error(ev.OnPos, 0, "E073",
				fmt.Sprintf("unknown event '%s' for component %s", ev.Event, decl.Component),
				fmt.Sprintf("valid events: %s", eventNames(ct.Events)))
			continue
		}

		// Resolve handler name.
		handlerIdent, ok := ev.Handler.(*ast.Ident)
		if !ok {
			c.error(ev.Handler.Pos(), 0, "E074", "event handler must be a function name", "")
			continue
		}
		handlerSym := c.scope.Lookup(handlerIdent.Name)
		if handlerSym == nil {
			c.error(ev.Handler.Pos(), 0, "E021", fmt.Sprintf("undefined name '%s'", handlerIdent.Name), "")
			continue
		}
		handlerSig, ok := handlerSym.Type.(*FuncSignature)
		if !ok {
			c.error(ev.Handler.Pos(), 0, "E074",
				fmt.Sprintf("'%s' is not a function", handlerIdent.Name), "")
			continue
		}

		// Validate handler signature: must have 0 params or 1 param matching event type.
		if len(handlerSig.Params) > 1 {
			c.error(ev.Handler.Pos(), 0, "E074",
				fmt.Sprintf("event handler '%s' has too many parameters (expected 0 or 1)", handlerIdent.Name), "")
			continue
		}
		if len(handlerSig.Params) == 1 {
			paramType := handlerSig.Params[0].Type
			if !IsAssignable(eventType, paramType) {
				c.error(ev.Handler.Pos(), 0, "E074",
					fmt.Sprintf("event handler '%s' parameter type %s does not match event type %s",
						handlerIdent.Name, FormatType(paramType), FormatType(eventType)), "")
			}
		}
	}

	// Recurse into children.
	for _, child := range decl.Children {
		c.checkComponentDecl(child)
	}
}

// propertyNames returns a comma-separated list of property names for diagnostics.
func propertyNames(props map[string]*ParamInfo) string {
	names := make([]string, 0, len(props))
	for name := range props {
		names = append(names, name)
	}
	return strings.Join(names, ", ")
}

// eventNames returns a comma-separated list of event names for diagnostics.
func eventNames(events map[string]Type) string {
	names := make([]string, 0, len(events))
	for name := range events {
		names = append(names, name)
	}
	return strings.Join(names, ", ")
}

// checkStmt checks a statement.
func (c *Checker) checkStmt(stmt ast.Stmt, returnType Type) {
	if stmt == nil {
		return
	}
	switch s := stmt.(type) {
	case *ast.ExprStmt:
		c.checkExpr(s.X)
	case *ast.AssignStmt:
		c.checkAssignStmt(s)
	case *ast.ReturnStmt:
		c.checkReturnStmt(s, returnType)
	case *ast.IfStmt:
		c.checkIfStmt(s, returnType)
	case *ast.SwitchStmt:
		c.checkSwitchStmt(s, returnType)
	case *ast.LoopStmt:
		c.checkLoopStmt(s, returnType)
	case *ast.VarDecl:
		c.checkVarDecl(s)
	case *ast.ConstDecl:
		c.checkConstDecl(s)
	case *ast.ComponentDecl:
		c.checkComponentDecl(s)
	case *ast.PassStmt, *ast.BreakStmt, *ast.ContinueStmt:
		// Valid, no checking needed.
	}
}

// checkAssignStmt checks an assignment.
func (c *Checker) checkAssignStmt(a *ast.AssignStmt) {
	// Handle implicit variable declaration: name = expr
	if ident, ok := a.Target.(*ast.Ident); ok {
		sym := c.scope.Lookup(ident.Name)
		if sym == nil {
			// Implicit variable declaration via assignment.
			valType := c.checkExpr(a.Value)
			newSym := &Symbol{
				Name: ident.Name,
				Type: valType,
				Kind: SymVar,
				Pos:  ident.NamePos,
			}
			if dup := c.scope.Define(newSym); dup != "" {
				c.error(ident.NamePos, 0, "E050", fmt.Sprintf("duplicate name '%s' in this scope", dup), "")
			}
			return
		}
		// Existing variable: check reassignment.
		if sym.Kind == SymConst {
			c.error(a.EqPos, 0, "E033", fmt.Sprintf("cannot reassign const '%s'", ident.Name), "")
		}
	}

	targetType := c.checkExpr(a.Target)
	valType := c.checkExpr(a.Value)
	c.checkAssignment(valType, targetType, a.Value.Pos(), a.Value.End().Col-a.Value.Pos().Col)
}

// checkReturnStmt checks a return statement.
func (c *Checker) checkReturnStmt(r *ast.ReturnStmt, returnType Type) {
	if r.Value != nil {
		valType := c.checkExpr(r.Value)
		if returnType == nil {
			c.error(r.ReturnPos, 0, "E037", "return with value in void function", "remove the return value or add a return type to the function")
		} else if valType != nil && returnType != nil {
			if !IsAssignable(valType, returnType) {
				c.error(r.ReturnPos, 0, "E036", fmt.Sprintf("return type mismatch: expected %s, got %s", FormatType(returnType), FormatType(valType)), "")
			}
		}
	}
	// Bare return in non-void function: check if all paths return a value
	// is a more advanced check (not implemented here yet).
}

// checkIfStmt checks an if statement.
func (c *Checker) checkIfStmt(i *ast.IfStmt, returnType Type) {
	condType := c.checkExpr(i.Condition)
	if condType != nil && !IsAssignable(condType, BoolType) {
		c.error(i.IfPos, 0, "E047", fmt.Sprintf("condition must be bool, got %s", FormatType(condType)), "")
	}
	for _, stmt := range i.Body {
		c.checkStmt(stmt, returnType)
	}
	for _, elif := range i.ElseIfs {
		elifCond := c.checkExpr(elif.Condition)
		if elifCond != nil && !IsAssignable(elifCond, BoolType) {
			c.error(elif.ElseIfPos, 0, "E047", fmt.Sprintf("condition must be bool, got %s", FormatType(elifCond)), "")
		}
		for _, stmt := range elif.Body {
			c.checkStmt(stmt, returnType)
		}
	}
	for _, stmt := range i.Else {
		c.checkStmt(stmt, returnType)
	}
}

// checkSwitchStmt checks a switch statement.
func (c *Checker) checkSwitchStmt(s *ast.SwitchStmt, returnType Type) {
	c.checkExpr(s.Subject)
	for _, cs := range s.Cases {
		for _, p := range cs.Patterns {
			c.checkExpr(p)
		}
		for _, stmt := range cs.Body {
			c.checkStmt(stmt, returnType)
		}
	}
}

// checkLoopStmt checks a loop statement.
func (c *Checker) checkLoopStmt(l *ast.LoopStmt, returnType Type) {
	// Create loop scope with loop variables BEFORE checking modifiers
	// (so loop variables are in scope for while/if conditions).
	loopScope := NewScope(c.scope)
	prevScope := c.scope
	c.scope = loopScope

	for _, iter := range l.Iterators {
		iterType := c.checkExpr(iter.Iterable)
		if iter.Variable != "_" {
			var varType Type
			if iterType == nil {
				// Range expression — loop variable is int.
				varType = IntType
			} else if gt, ok := iterType.(*GenericType); ok {
				switch gt.Name {
				case "List":
					if len(gt.Params) == 1 {
						varType = gt.Params[0]
					}
				case "Dict":
					varType = gt
				}
			}
			loopScope.Define(&Symbol{
				Name: iter.Variable,
				Type: varType,
				Kind: SymVar,
				Pos:  iter.IterPos,
			})
		}
	}

	// Check modifiers (loop variables are now in scope).
	if l.Step != nil {
		stepType := c.checkExpr(l.Step)
		if stepType != nil && !IsAssignable(stepType, IntType) {
			c.error(l.Step.Pos(), l.Step.End().Col-l.Step.Pos().Col, "E022", fmt.Sprintf("step must be int, got %s", FormatType(stepType)), "")
		}
	}
	if l.While != nil {
		whileType := c.checkExpr(l.While)
		if whileType != nil && !IsAssignable(whileType, BoolType) {
			c.error(l.While.Pos(), 0, "E047", fmt.Sprintf("while condition must be bool, got %s", FormatType(whileType)), "")
		}
	}
	if l.IfCond != nil {
		ifType := c.checkExpr(l.IfCond)
		if ifType != nil && !IsAssignable(ifType, BoolType) {
			c.error(l.IfCond.Pos(), 0, "E047", fmt.Sprintf("if condition must be bool, got %s", FormatType(ifType)), "")
		}
	}

	for _, stmt := range l.Body {
		c.checkStmt(stmt, returnType)
	}

	c.scope = prevScope
}

// Helper functions for type checking.

func isString(t Type) bool {
	if p, ok := t.(*PrimitiveType); ok {
		return p.Name == "string"
	}
	return false
}

func isNumeric(t Type) bool {
	if p, ok := t.(*PrimitiveType); ok {
		return p.Name == "int" || p.Name == "float"
	}
	return false
}

func isBool(t Type) bool {
	if p, ok := t.(*PrimitiveType); ok {
		return p.Name == "bool"
	}
	return false
}

func numericResult(a, b Type) Type {
	if p, ok := a.(*PrimitiveType); ok && p.Name == "float" {
		return FloatType
	}
	if p, ok := b.(*PrimitiveType); ok && p.Name == "float" {
		return FloatType
	}
	return IntType
}

// checkAssignment checks a value-to-target type assignment and reports the appropriate error.
func (c *Checker) checkAssignment(valType, targetType Type, pos token.Position, span int) {
	if valType == nil || targetType == nil {
		return
	}
	// Special case: null to non-nullable.
	if _, ok := valType.(*NullType); ok {
		if _, isNullable := targetType.(*NullableType); !isNullable {
			c.error(pos, span, "E023",
				fmt.Sprintf("cannot assign null to non-nullable type %s", FormatType(targetType)), "")
			return
		}
	}
	if !IsAssignable(valType, targetType) {
		c.error(pos, span, "E022",
			fmt.Sprintf("cannot use %s as %s", FormatType(valType), FormatType(targetType)), "")
	}
}

func joinPath(path []string) string {
	if len(path) == 0 {
		return ""
	}
	out := path[0]
	for _, part := range path[1:] {
		out += "." + part
	}
	return out
}

func (c *Checker) error(pos token.Position, span int, code, msg, hint string) {
	c.diag.ReportError(code, msg, pos, span, hint, "")
}
