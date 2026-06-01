// Copyright (c) 2026 LealLang Contributors
// SPDX-License-Identifier: MIT

// Package ast defines the LealLang abstract syntax tree.
package ast

import "github.com/LealLang/leallang/internal/token"

// Node is implemented by every AST node.
type Node interface {
	Pos() token.Position
	End() token.Position
}

// Expr is implemented by every expression node.
type Expr interface {
	Node
	exprNode()
}

// Stmt is implemented by every statement node.
type Stmt interface {
	Node
	stmtNode()
}

// Decl is implemented by every declaration node.
type Decl interface {
	Node
	declNode()
}

// TypeExpr represents all type annotations.
type TypeExpr interface {
	Node
	typeNode()
}

// Program is the root node of every .ll file.
type Program struct {
	Package *PackageDecl
	Imports []*ImportDecl
	Decls   []Decl
}

// PackageDecl represents `package app.main`.
type PackageDecl struct {
	PkgPos token.Position
	Path   []string
}

// ImportDecl represents `import app.settings` or `import app.settings as s`.
type ImportDecl struct {
	ImportPos token.Position
	Path      []string
	Alias     string
}

// VarDecl represents a variable declaration or inferred declaration syntax.
type VarDecl struct {
	NamePos token.Position
	Name    string
	Type    TypeExpr
	Value   Expr
}

// ConstDecl represents a const declaration.
type ConstDecl struct {
	ConstPos token.Position
	Name     string
	Type     TypeExpr
	Value    Expr
}

// FuncDecl represents a function or method declaration.
type FuncDecl struct {
	FuncPos     token.Position
	Pub         bool
	Async       bool
	UI          bool
	Name        string
	Params      []*Param
	ReturnTypes []TypeExpr
	Body        []Stmt
}

// Param represents a function or constructor parameter.
type Param struct {
	Ref  bool
	Name string
	Type TypeExpr
	Posn token.Position
}

// Pos returns the parameter start position.
func (p *Param) Pos() token.Position { return p.Posn }

// End returns the parameter end position.
func (p *Param) End() token.Position {
	if p.Type != nil {
		return p.Type.End()
	}
	return endOfName(p.Posn, p.Name)
}

// TypeDecl represents a record/type declaration.
type TypeDecl struct {
	TypePos     token.Position
	Name        string
	Extends     string
	Fields      []*FieldDecl
	Constructor *ConstructorDecl
	Methods     []*FuncDecl
}

// FieldDecl represents a field inside a type declaration.
type FieldDecl struct {
	FieldPos token.Position
	Pub      bool
	Name     string
	Type     TypeExpr
	Default  Expr
}

// ConstructorDecl represents a constructor declaration.
type ConstructorDecl struct {
	ConstructorPos token.Position
	Params         []*Param
	Body           []Stmt
}

// SimpleType represents string, int, bool, float, char, any, or a named type.
type SimpleType struct {
	TypePos token.Position
	Name    string
}

// NullableType represents a nullable type annotation.
type NullableType struct {
	Inner TypeExpr
	QPos  token.Position
}

// GenericType represents a generic type annotation.
type GenericType struct {
	TypePos token.Position
	Name    string
	Params  []TypeExpr
}

// ExprStmt wraps an expression used as a statement.
type ExprStmt struct {
	X Expr
}

// AssignStmt represents target = value.
type AssignStmt struct {
	Target Expr
	EqPos  token.Position
	Value  Expr
}

// ReturnStmt represents return with an optional value.
type ReturnStmt struct {
	ReturnPos token.Position
	Value     Expr
}

// PassStmt represents pass.
type PassStmt struct {
	PassPos token.Position
}

// BreakStmt represents break.
type BreakStmt struct {
	BreakPos token.Position
}

// ContinueStmt represents continue.
type ContinueStmt struct {
	ContinuePos token.Position
}

// IfStmt represents if / else if / else.
type IfStmt struct {
	IfPos     token.Position
	Condition Expr
	Body      []Stmt
	ElseIfs   []*ElseIfClause
	Else      []Stmt
}

// ElseIfClause represents an else-if branch.
type ElseIfClause struct {
	ElseIfPos token.Position
	Condition Expr
	Body      []Stmt
}

// SwitchStmt represents a statement-form switch.
type SwitchStmt struct {
	SwitchPos token.Position
	Subject   Expr
	Cases     []*SwitchCase
}

// SwitchCase represents a statement-switch case.
type SwitchCase struct {
	CasePos  token.Position
	Patterns []Expr
	Body     []Stmt
}

// LoopStmt represents all loop forms.
type LoopStmt struct {
	LoopPos   token.Position
	Iterators []*LoopIterator
	Step      Expr
	While     Expr
	IfCond    Expr
	Body      []Stmt
}

// LoopIterator represents one loop iterator clause.
type LoopIterator struct {
	IterPos  token.Position
	Variable string
	Iterable Expr
}

// Literal represents a literal value.
type Literal struct {
	ValuePos token.Position
	Kind     token.TokenKind
	Value    string
}

// InterpStringExpr represents an interpolated string.
type InterpStringExpr struct {
	StringPos token.Position
	Segments  []InterpSegment
}

// InterpSegment is one text or expression segment in an interpolated string.
type InterpSegment struct {
	IsExpr bool
	Text   string
	Expr   Expr
}

// Ident represents any plain identifier.
type Ident struct {
	NamePos token.Position
	Name    string
}

// BinaryExpr represents a binary operation.
type BinaryExpr struct {
	Left  Expr
	Op    token.TokenKind
	OpPos token.Position
	Right Expr
}

// UnaryExpr represents a unary operation.
type UnaryExpr struct {
	OpPos token.Position
	Op    token.TokenKind
	X     Expr
}

// AwaitExpr represents `await <expr>`.
type AwaitExpr struct {
	AwaitPos token.Position
	X        Expr
}

// CallExpr represents a function call.
type CallExpr struct {
	Func   Expr
	LParen token.Position
	Args   []*Argument
	RParen token.Position
}

// Argument represents a call argument.
type Argument struct {
	Name  string
	Value Expr
	Posn  token.Position
}

// Pos returns the argument start position.
func (a *Argument) Pos() token.Position { return a.Posn }

// End returns the argument end position.
func (a *Argument) End() token.Position {
	if a.Value != nil {
		return a.Value.End()
	}
	return endOfName(a.Posn, a.Name)
}

// FieldExpr represents obj.field.
type FieldExpr struct {
	X     Expr
	Dot   token.Position
	Field string
}

// IndexExpr represents obj[index].
type IndexExpr struct {
	X      Expr
	LBrack token.Position
	Index  Expr
	RBrack token.Position
}

// ComponentRefExpr represents @Button[save_btn].
type ComponentRefExpr struct {
	AtPos     token.Position
	Component string
	ID        string
}

// ComponentDecl represents a UI component declaration: ComponentType[id]:
type ComponentDecl struct {
	CompPos   token.Position
	Component string // "Window", "Button", "Col"
	ID        string // "main", "save_btn"
	Props     []*ComponentProp
	Events    []*EventBinding
	Children  []*ComponentDecl
}

// ComponentProp represents a property assignment inside a component block.
type ComponentProp struct {
	PropPos token.Position
	Name    string
	Value   Expr
}

// EventBinding represents an event handler binding inside a component block.
type EventBinding struct {
	OnPos   token.Position
	Event   string
	Handler Expr // Ident referencing a function name
}

// RangeExpr represents inclusive and exclusive ranges.
type RangeExpr struct {
	Low       Expr
	High      Expr
	Exclusive bool
	OpPos     token.Position
}

// ListLiteral represents [a, b, c].
type ListLiteral struct {
	LBrack   token.Position
	Elements []Expr
	RBrack   token.Position
}

// DictLiteral represents {key: value}.
type DictLiteral struct {
	LBrace token.Position
	Pairs  []*DictPair
	RBrace token.Position
}

// DictPair represents one dictionary key/value pair.
type DictPair struct {
	Key   Expr
	Value Expr
}

// SwitchExpr represents an expression-form switch.
type SwitchExpr struct {
	SwitchPos token.Position
	Subject   Expr
	Arms      []*SwitchArm
}

// SwitchArm represents one expression-switch arm.
type SwitchArm struct {
	ArmPos   token.Position
	Patterns []Expr
	Value    Expr
}

// BadExpr preserves parser recovery points in expression positions.
type BadExpr struct {
	Start token.Position
	Stop  token.Position
}

func (*VarDecl) declNode()          {}
func (*VarDecl) stmtNode()          {}
func (*ConstDecl) declNode()        {}
func (*ConstDecl) stmtNode()        {}
func (*FuncDecl) declNode()         {}
func (*TypeDecl) declNode()         {}
func (*ExprStmt) stmtNode()         {}
func (*AssignStmt) stmtNode()       {}
func (*ReturnStmt) stmtNode()       {}
func (*PassStmt) stmtNode()         {}
func (*BreakStmt) stmtNode()        {}
func (*ContinueStmt) stmtNode()     {}
func (*IfStmt) stmtNode()           {}
func (*SwitchStmt) stmtNode()       {}
func (*LoopStmt) stmtNode()         {}
func (*ComponentDecl) stmtNode()    {}
func (*ComponentDecl) declNode()    {}
func (*Literal) exprNode()          {}
func (*InterpStringExpr) exprNode() {}
func (*Ident) exprNode()            {}
func (*BinaryExpr) exprNode()       {}
func (*UnaryExpr) exprNode()        {}
func (*AwaitExpr) exprNode()        {}
func (*CallExpr) exprNode()         {}
func (*FieldExpr) exprNode()        {}
func (*IndexExpr) exprNode()        {}
func (*ComponentRefExpr) exprNode() {}
func (*RangeExpr) exprNode()        {}
func (*ListLiteral) exprNode()      {}
func (*DictLiteral) exprNode()      {}
func (*SwitchExpr) exprNode()       {}
func (*BadExpr) exprNode()          {}
func (*SimpleType) typeNode()       {}
func (*NullableType) typeNode()     {}
func (*GenericType) typeNode()      {}

func (p *Program) Pos() token.Position {
	if p.Package != nil {
		return p.Package.Pos()
	}
	if len(p.Imports) > 0 {
		return p.Imports[0].Pos()
	}
	if len(p.Decls) > 0 {
		return p.Decls[0].Pos()
	}
	return token.Position{}
}

func (p *Program) End() token.Position {
	if len(p.Decls) > 0 {
		return p.Decls[len(p.Decls)-1].End()
	}
	if len(p.Imports) > 0 {
		return p.Imports[len(p.Imports)-1].End()
	}
	if p.Package != nil {
		return p.Package.End()
	}
	return token.Position{}
}

func (p *PackageDecl) Pos() token.Position { return p.PkgPos }
func (p *PackageDecl) End() token.Position { return endOfPath(p.PkgPos, "package", p.Path) }
func (i *ImportDecl) Pos() token.Position  { return i.ImportPos }
func (i *ImportDecl) End() token.Position {
	if i.Alias != "" {
		return endOfName(i.ImportPos, "import "+joinPath(i.Path)+" as "+i.Alias)
	}
	return endOfPath(i.ImportPos, "import", i.Path)
}
func (v *VarDecl) Pos() token.Position { return v.NamePos }
func (v *VarDecl) End() token.Position {
	if v.Value != nil {
		return v.Value.End()
	}
	if v.Type != nil {
		return v.Type.End()
	}
	return endOfName(v.NamePos, v.Name)
}
func (c *ConstDecl) Pos() token.Position { return c.ConstPos }
func (c *ConstDecl) End() token.Position {
	if c.Value != nil {
		return c.Value.End()
	}
	if c.Type != nil {
		return c.Type.End()
	}
	return endOfName(c.ConstPos, "const "+c.Name)
}
func (f *FuncDecl) Pos() token.Position { return f.FuncPos }
func (f *FuncDecl) End() token.Position {
	if len(f.Body) > 0 {
		return f.Body[len(f.Body)-1].End()
	}
	if len(f.ReturnTypes) > 0 {
		return f.ReturnTypes[len(f.ReturnTypes)-1].End()
	}
	return endOfName(f.FuncPos, "func "+f.Name)
}
func (t *TypeDecl) Pos() token.Position { return t.TypePos }
func (t *TypeDecl) End() token.Position {
	if len(t.Methods) > 0 {
		return t.Methods[len(t.Methods)-1].End()
	}
	if t.Constructor != nil {
		return t.Constructor.End()
	}
	if len(t.Fields) > 0 {
		return t.Fields[len(t.Fields)-1].End()
	}
	return endOfName(t.TypePos, "type "+t.Name)
}
func (f *FieldDecl) Pos() token.Position { return f.FieldPos }
func (f *FieldDecl) End() token.Position {
	if f.Default != nil {
		return f.Default.End()
	}
	if f.Type != nil {
		return f.Type.End()
	}
	return endOfName(f.FieldPos, f.Name)
}
func (c *ConstructorDecl) Pos() token.Position { return c.ConstructorPos }
func (c *ConstructorDecl) End() token.Position {
	if len(c.Body) > 0 {
		return c.Body[len(c.Body)-1].End()
	}
	return endOfName(c.ConstructorPos, "constructor")
}
func (s *SimpleType) Pos() token.Position { return s.TypePos }
func (s *SimpleType) End() token.Position { return endOfName(s.TypePos, s.Name) }
func (n *NullableType) Pos() token.Position {
	if n.Inner != nil {
		return n.Inner.Pos()
	}
	return n.QPos
}
func (n *NullableType) End() token.Position { return advancePosition(n.QPos, "?") }
func (g *GenericType) Pos() token.Position  { return g.TypePos }
func (g *GenericType) End() token.Position {
	if len(g.Params) > 0 {
		return advancePosition(g.Params[len(g.Params)-1].End(), ">")
	}
	return endOfName(g.TypePos, g.Name)
}
func (e *ExprStmt) Pos() token.Position { return nodePos(e.X) }
func (e *ExprStmt) End() token.Position { return nodeEnd(e.X) }
func (a *AssignStmt) Pos() token.Position {
	return nodePos(a.Target)
}
func (a *AssignStmt) End() token.Position {
	if a.Value != nil {
		return a.Value.End()
	}
	return advancePosition(a.EqPos, "=")
}
func (r *ReturnStmt) Pos() token.Position { return r.ReturnPos }
func (r *ReturnStmt) End() token.Position {
	if r.Value != nil {
		return r.Value.End()
	}
	return endOfName(r.ReturnPos, "return")
}
func (p *PassStmt) Pos() token.Position     { return p.PassPos }
func (p *PassStmt) End() token.Position     { return endOfName(p.PassPos, "pass") }
func (b *BreakStmt) Pos() token.Position    { return b.BreakPos }
func (b *BreakStmt) End() token.Position    { return endOfName(b.BreakPos, "break") }
func (c *ContinueStmt) Pos() token.Position { return c.ContinuePos }
func (c *ContinueStmt) End() token.Position { return endOfName(c.ContinuePos, "continue") }
func (i *IfStmt) Pos() token.Position       { return i.IfPos }
func (i *IfStmt) End() token.Position {
	if len(i.Else) > 0 {
		return i.Else[len(i.Else)-1].End()
	}
	if len(i.ElseIfs) > 0 {
		return i.ElseIfs[len(i.ElseIfs)-1].End()
	}
	if len(i.Body) > 0 {
		return i.Body[len(i.Body)-1].End()
	}
	return nodeEnd(i.Condition)
}
func (e *ElseIfClause) Pos() token.Position { return e.ElseIfPos }
func (e *ElseIfClause) End() token.Position {
	if len(e.Body) > 0 {
		return e.Body[len(e.Body)-1].End()
	}
	return nodeEnd(e.Condition)
}
func (s *SwitchStmt) Pos() token.Position { return s.SwitchPos }
func (s *SwitchStmt) End() token.Position {
	if len(s.Cases) > 0 {
		return s.Cases[len(s.Cases)-1].End()
	}
	return nodeEnd(s.Subject)
}
func (s *SwitchCase) Pos() token.Position { return s.CasePos }
func (s *SwitchCase) End() token.Position {
	if len(s.Body) > 0 {
		return s.Body[len(s.Body)-1].End()
	}
	if len(s.Patterns) > 0 {
		return s.Patterns[len(s.Patterns)-1].End()
	}
	return s.CasePos
}
func (l *LoopStmt) Pos() token.Position { return l.LoopPos }
func (l *LoopStmt) End() token.Position {
	if len(l.Body) > 0 {
		return l.Body[len(l.Body)-1].End()
	}
	if l.IfCond != nil {
		return l.IfCond.End()
	}
	if l.While != nil {
		return l.While.End()
	}
	if l.Step != nil {
		return l.Step.End()
	}
	if len(l.Iterators) > 0 {
		return l.Iterators[len(l.Iterators)-1].End()
	}
	return endOfName(l.LoopPos, "loop")
}
func (l *LoopIterator) Pos() token.Position { return l.IterPos }
func (l *LoopIterator) End() token.Position {
	if l.Iterable != nil {
		return l.Iterable.End()
	}
	return endOfName(l.IterPos, l.Variable)
}
func (l *Literal) Pos() token.Position          { return l.ValuePos }
func (l *Literal) End() token.Position          { return advancePosition(l.ValuePos, l.Value) }
func (i *InterpStringExpr) Pos() token.Position { return i.StringPos }
func (i *InterpStringExpr) End() token.Position {
	return i.StringPos
}
func (i *Ident) Pos() token.Position            { return i.NamePos }
func (i *Ident) End() token.Position            { return endOfName(i.NamePos, i.Name) }
func (b *BinaryExpr) Pos() token.Position       { return nodePos(b.Left) }
func (b *BinaryExpr) End() token.Position       { return nodeEnd(b.Right) }
func (u *UnaryExpr) Pos() token.Position        { return u.OpPos }
func (u *UnaryExpr) End() token.Position        { return nodeEnd(u.X) }
func (a *AwaitExpr) Pos() token.Position        { return a.AwaitPos }
func (a *AwaitExpr) End() token.Position        { return nodeEnd(a.X) }
func (c *CallExpr) Pos() token.Position         { return nodePos(c.Func) }
func (c *CallExpr) End() token.Position         { return advancePosition(c.RParen, ")") }
func (f *FieldExpr) Pos() token.Position        { return nodePos(f.X) }
func (f *FieldExpr) End() token.Position        { return endOfName(f.Dot, "."+f.Field) }
func (i *IndexExpr) Pos() token.Position        { return nodePos(i.X) }
func (i *IndexExpr) End() token.Position        { return advancePosition(i.RBrack, "]") }
func (c *ComponentRefExpr) Pos() token.Position { return c.AtPos }
func (c *ComponentRefExpr) End() token.Position {
	return endOfName(c.AtPos, "@"+c.Component+"["+c.ID+"]")
}
func (c *ComponentDecl) Pos() token.Position { return c.CompPos }
func (c *ComponentDecl) End() token.Position {
	if len(c.Children) > 0 {
		return c.Children[len(c.Children)-1].End()
	}
	if len(c.Events) > 0 {
		return c.Events[len(c.Events)-1].End()
	}
	if len(c.Props) > 0 {
		return c.Props[len(c.Props)-1].End()
	}
	return endOfName(c.CompPos, c.Component+"["+c.ID+"]")
}
func (p *ComponentProp) Pos() token.Position { return p.PropPos }
func (p *ComponentProp) End() token.Position {
	if p.Value != nil {
		return p.Value.End()
	}
	return endOfName(p.PropPos, p.Name)
}
func (e *EventBinding) Pos() token.Position { return e.OnPos }
func (e *EventBinding) End() token.Position {
	if e.Handler != nil {
		return e.Handler.End()
	}
	return endOfName(e.OnPos, "on "+e.Event)
}
func (r *RangeExpr) Pos() token.Position   { return nodePos(r.Low) }
func (r *RangeExpr) End() token.Position   { return nodeEnd(r.High) }
func (l *ListLiteral) Pos() token.Position { return l.LBrack }
func (l *ListLiteral) End() token.Position { return advancePosition(l.RBrack, "]") }
func (d *DictLiteral) Pos() token.Position { return d.LBrace }
func (d *DictLiteral) End() token.Position { return advancePosition(d.RBrace, "}") }
func (s *SwitchExpr) Pos() token.Position  { return s.SwitchPos }
func (s *SwitchExpr) End() token.Position {
	if len(s.Arms) > 0 {
		return s.Arms[len(s.Arms)-1].End()
	}
	return nodeEnd(s.Subject)
}
func (s *SwitchArm) Pos() token.Position { return s.ArmPos }
func (s *SwitchArm) End() token.Position {
	if s.Value != nil {
		return s.Value.End()
	}
	if len(s.Patterns) > 0 {
		return s.Patterns[len(s.Patterns)-1].End()
	}
	return s.ArmPos
}
func (b *BadExpr) Pos() token.Position { return b.Start }
func (b *BadExpr) End() token.Position { return b.Stop }

func nodePos(n Node) token.Position {
	if n == nil {
		return token.Position{}
	}
	return n.Pos()
}

func nodeEnd(n Node) token.Position {
	if n == nil {
		return token.Position{}
	}
	return n.End()
}

func endOfPath(pos token.Position, prefix string, path []string) token.Position {
	text := prefix
	if len(path) > 0 {
		text += " " + joinPath(path)
	}
	return advancePosition(pos, text)
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

func endOfName(pos token.Position, name string) token.Position {
	return advancePosition(pos, name)
}

func advancePosition(pos token.Position, text string) token.Position {
	end := pos
	if end.Line == 0 {
		return end
	}
	if text == "" {
		return end
	}
	for _, r := range text {
		end.Offset++
		if r == '\n' {
			end.Line++
			end.Col = 1
		} else {
			end.Col++
		}
	}
	return end
}
