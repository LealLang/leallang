// Copyright (c) 2026 LealLang Contributors
// SPDX-License-Identifier: MIT

package ast

import (
	"fmt"
	"io"
	"strings"
)

// Print writes an indented, human-readable representation of node.
func Print(w io.Writer, node Node) {
	p := printer{w: w}
	p.node(node, 0)
}

type printer struct {
	w io.Writer
}

func (p *printer) line(indent int, format string, args ...any) {
	_, _ = fmt.Fprintf(p.w, "%s%s\n", strings.Repeat("  ", indent), fmt.Sprintf(format, args...))
}

func (p *printer) node(node Node, indent int) {
	switch n := node.(type) {
	case nil:
		p.line(indent, "<nil>")
	case *Program:
		p.line(indent, "Program")
		if n.Package != nil {
			p.node(n.Package, indent+1)
		}
		for _, imp := range n.Imports {
			p.node(imp, indent+1)
		}
		for _, decl := range n.Decls {
			p.node(decl, indent+1)
		}
	case *PackageDecl:
		p.line(indent, "PackageDecl")
		p.line(indent+1, "Path: %q", n.Path)
	case *ImportDecl:
		if n.Alias != "" {
			p.line(indent, "ImportDecl alias=%s", n.Alias)
		} else {
			p.line(indent, "ImportDecl")
		}
		p.line(indent+1, "Path: %q", n.Path)
	case *VarDecl:
		p.line(indent, "VarDecl name=%s", n.Name)
		p.typeLine(indent+1, "Type", n.Type)
		if n.Value != nil {
			p.line(indent+1, "Value:")
			p.node(n.Value, indent+2)
		}
	case *ConstDecl:
		p.line(indent, "ConstDecl name=%s", n.Name)
		p.typeLine(indent+1, "Type", n.Type)
		if n.Value != nil {
			p.line(indent+1, "Value:")
			p.node(n.Value, indent+2)
		}
	case *FuncDecl:
		p.line(indent, "FuncDecl name=%s pub=%t", n.Name, n.Pub)
		for _, param := range n.Params {
			p.line(indent+1, "Param name=%s type=%s ref=%t", param.Name, typeString(param.Type), param.Ref)
		}
		if n.ReturnType == nil {
			p.line(indent+1, "ReturnType: <none>")
		} else {
			p.line(indent+1, "ReturnType: %s", typeString(n.ReturnType))
		}
		p.line(indent+1, "Body:")
		for _, stmt := range n.Body {
			p.node(stmt, indent+2)
		}
	case *TypeDecl:
		if n.Extends != "" {
			p.line(indent, "TypeDecl name=%s extends=%s", n.Name, n.Extends)
		} else {
			p.line(indent, "TypeDecl name=%s", n.Name)
		}
		for _, field := range n.Fields {
			p.node(field, indent+1)
		}
		if n.Constructor != nil {
			p.node(n.Constructor, indent+1)
		}
		for _, method := range n.Methods {
			p.node(method, indent+1)
		}
	case *FieldDecl:
		p.line(indent, "FieldDecl name=%s type=%s", n.Name, typeString(n.Type))
		if n.Default != nil {
			p.line(indent+1, "Default:")
			p.node(n.Default, indent+2)
		}
	case *ConstructorDecl:
		p.line(indent, "ConstructorDecl")
		for _, param := range n.Params {
			p.line(indent+1, "Param name=%s type=%s ref=%t", param.Name, typeString(param.Type), param.Ref)
		}
		p.line(indent+1, "Body:")
		for _, stmt := range n.Body {
			p.node(stmt, indent+2)
		}
	case *ExprStmt:
		p.line(indent, "ExprStmt")
		p.node(n.X, indent+1)
	case *AssignStmt:
		p.line(indent, "AssignStmt")
		p.line(indent+1, "Target:")
		p.node(n.Target, indent+2)
		p.line(indent+1, "Value:")
		p.node(n.Value, indent+2)
	case *ReturnStmt:
		p.line(indent, "ReturnStmt")
		if n.Value != nil {
			p.node(n.Value, indent+1)
		}
	case *PassStmt:
		p.line(indent, "PassStmt")
	case *BreakStmt:
		p.line(indent, "BreakStmt")
	case *ContinueStmt:
		p.line(indent, "ContinueStmt")
	case *IfStmt:
		p.line(indent, "IfStmt")
		p.line(indent+1, "Condition:")
		p.node(n.Condition, indent+2)
		p.line(indent+1, "Body:")
		for _, stmt := range n.Body {
			p.node(stmt, indent+2)
		}
		for _, elseif := range n.ElseIfs {
			p.node(elseif, indent+1)
		}
		if n.Else != nil {
			p.line(indent+1, "Else:")
			for _, stmt := range n.Else {
				p.node(stmt, indent+2)
			}
		}
	case *ElseIfClause:
		p.line(indent, "ElseIfClause")
		p.line(indent+1, "Condition:")
		p.node(n.Condition, indent+2)
		p.line(indent+1, "Body:")
		for _, stmt := range n.Body {
			p.node(stmt, indent+2)
		}
	case *SwitchStmt:
		p.line(indent, "SwitchStmt")
		p.line(indent+1, "Subject:")
		p.node(n.Subject, indent+2)
		for _, c := range n.Cases {
			p.node(c, indent+1)
		}
	case *SwitchCase:
		p.line(indent, "SwitchCase")
		p.patterns(n.Patterns, indent+1)
		p.line(indent+1, "Body:")
		for _, stmt := range n.Body {
			p.node(stmt, indent+2)
		}
	case *LoopStmt:
		p.line(indent, "LoopStmt")
		for _, iter := range n.Iterators {
			p.node(iter, indent+1)
		}
		if n.Step != nil {
			p.line(indent+1, "Step:")
			p.node(n.Step, indent+2)
		}
		if n.While != nil {
			p.line(indent+1, "While:")
			p.node(n.While, indent+2)
		}
		if n.IfCond != nil {
			p.line(indent+1, "If:")
			p.node(n.IfCond, indent+2)
		}
		p.line(indent+1, "Body:")
		for _, stmt := range n.Body {
			p.node(stmt, indent+2)
		}
	case *LoopIterator:
		p.line(indent, "LoopIterator variable=%s", n.Variable)
		p.node(n.Iterable, indent+1)
	case *Literal:
		p.line(indent, "Literal kind=%s value=%q", n.Kind, n.Value)
	case *InterpStringExpr:
		p.line(indent, "InterpStringExpr")
		for _, segment := range n.Segments {
			if segment.IsExpr {
				p.line(indent+1, "Segment(expr):")
				p.node(segment.Expr, indent+2)
			} else {
				p.line(indent+1, "Segment(text): %q", segment.Text)
			}
		}
	case *Ident:
		p.line(indent, "Ident %q", n.Name)
	case *BinaryExpr:
		p.line(indent, "BinaryExpr op=%s", n.Op)
		p.node(n.Left, indent+1)
		p.node(n.Right, indent+1)
	case *UnaryExpr:
		p.line(indent, "UnaryExpr op=%s", n.Op)
		p.node(n.X, indent+1)
	case *CallExpr:
		p.line(indent, "CallExpr")
		p.node(n.Func, indent+1)
		p.line(indent+1, "Args:")
		for _, arg := range n.Args {
			if arg.Name == "" {
				p.line(indent+2, "Arg (positional)")
			} else {
				p.line(indent+2, "Arg name=%s", arg.Name)
			}
			p.node(arg.Value, indent+3)
		}
	case *FieldExpr:
		p.line(indent, "FieldExpr")
		p.node(n.X, indent+1)
		p.line(indent+1, "Field %q", n.Field)
	case *IndexExpr:
		p.line(indent, "IndexExpr")
		p.node(n.X, indent+1)
		p.node(n.Index, indent+1)
	case *ComponentRefExpr:
		p.line(indent, "ComponentRefExpr component=%s id=%s", n.Component, n.ID)
	case *RangeExpr:
		p.line(indent, "RangeExpr exclusive=%t", n.Exclusive)
		p.node(n.Low, indent+1)
		p.node(n.High, indent+1)
	case *ListLiteral:
		p.line(indent, "ListLiteral")
		for _, elem := range n.Elements {
			p.node(elem, indent+1)
		}
	case *DictLiteral:
		p.line(indent, "DictLiteral")
		for _, pair := range n.Pairs {
			p.line(indent+1, "Pair")
			p.node(pair.Key, indent+2)
			p.node(pair.Value, indent+2)
		}
	case *SwitchExpr:
		p.line(indent, "SwitchExpr")
		p.line(indent+1, "Subject:")
		p.node(n.Subject, indent+2)
		for _, arm := range n.Arms {
			p.node(arm, indent+1)
		}
	case *SwitchArm:
		p.line(indent, "SwitchArm")
		p.patterns(n.Patterns, indent+1)
		p.line(indent+1, "Value:")
		p.node(n.Value, indent+2)
	case *BadExpr:
		p.line(indent, "BadExpr")
	case TypeExpr:
		p.line(indent, "%s", typeString(n))
	default:
		p.line(indent, "%T", node)
	}
}

func (p *printer) patterns(patterns []Expr, indent int) {
	if patterns == nil {
		p.line(indent, "Patterns: _")
		return
	}
	p.line(indent, "Patterns:")
	for _, pattern := range patterns {
		p.node(pattern, indent+1)
	}
}

func (p *printer) typeLine(indent int, label string, t TypeExpr) {
	if t == nil {
		p.line(indent, "%s: <none>", label)
		return
	}
	p.line(indent, "%s: %s", label, typeString(t))
}

func typeString(t TypeExpr) string {
	switch n := t.(type) {
	case nil:
		return "<none>"
	case *SimpleType:
		return fmt.Sprintf("SimpleType(%s)", n.Name)
	case *NullableType:
		return fmt.Sprintf("NullableType(%s)", typeString(n.Inner))
	case *GenericType:
		parts := make([]string, 0, len(n.Params))
		for _, param := range n.Params {
			parts = append(parts, typeString(param))
		}
		return fmt.Sprintf("GenericType(%s<%s>)", n.Name, strings.Join(parts, ", "))
	default:
		return fmt.Sprintf("%T", t)
	}
}
