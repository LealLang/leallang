// Copyright (c) 2026 LealLang Contributors
// SPDX-License-Identifier: MIT

// Package parser parses LealLang token streams into ASTs.
package parser

import (
	"fmt"
	"strings"

	"github.com/LealLang/leallang/internal/ast"
	"github.com/LealLang/leallang/internal/diagnostics"
	"github.com/LealLang/leallang/internal/lexer"
	"github.com/LealLang/leallang/internal/token"
)

const (
	precLowest = iota
	precOr
	precAnd
	precNot
	precCompare
	precRange
	precAdd
	precMul
	precUnary
)

// Parser parses a token stream into an AST.
type Parser struct {
	tokens      []token.Token
	pos         int
	diagnostics *diagnostics.Diagnostics
	loopDepth   int
}

// New creates a parser for tokens.
func New(tokens []token.Token, diag *diagnostics.Diagnostics) *Parser {
	if diag == nil {
		diag = diagnostics.New()
	}
	if len(tokens) == 0 || tokens[len(tokens)-1].Kind != token.EOF {
		tokens = append(tokens, token.Token{Kind: token.EOF})
	}
	return &Parser{tokens: tokens, diagnostics: diag}
}

// Parse parses the token stream into a program.
func (p *Parser) Parse() *ast.Program {
	program := &ast.Program{}
	p.skipNewlines()

	if p.check(token.PACKAGE) {
		program.Package = p.parsePackageDecl()
	} else if !p.atEnd() {
		p.errorAt(p.peek(), "E019", "package declaration must be first", "start the file with a package declaration")
		p.synchronize()
		if p.check(token.PACKAGE) {
			program.Package = p.parsePackageDecl()
		}
	}

	for !p.atEnd() {
		p.skipNewlines()
		if p.atEnd() {
			break
		}
		if p.match(token.DEDENT) {
			continue
		}
		if p.check(token.PACKAGE) {
			p.errorAt(p.peek(), "E020", "multiple package declarations", "keep exactly one package declaration at the top of the file")
			p.advance()
			p.synchronize()
			continue
		}
		if p.check(token.IMPORT) {
			program.Imports = append(program.Imports, p.parseImportDecl())
			continue
		}
		decl := p.parseTopLevelDecl()
		if decl != nil {
			program.Decls = append(program.Decls, decl)
		}
	}

	return program
}

func (p *Parser) parsePackageDecl() *ast.PackageDecl {
	pkg := p.expect(token.PACKAGE, "package declaration")
	path := p.parseQualifiedPath()
	p.consumeTerminator()
	return &ast.PackageDecl{PkgPos: pkg.Pos, Path: path}
}

func (p *Parser) parseImportDecl() *ast.ImportDecl {
	imp := p.expect(token.IMPORT, "import declaration")
	path := p.parseQualifiedPath()
	alias := ""
	if p.match(token.AS) {
		aliasTok := p.expect(token.IDENT, "import alias")
		alias = aliasTok.Lexeme
	}
	p.consumeTerminator()
	return &ast.ImportDecl{ImportPos: imp.Pos, Path: path, Alias: alias}
}

func (p *Parser) parseQualifiedPath() []string {
	first := p.expect(token.IDENT, "identifier")
	path := []string{}
	if first.Kind == token.IDENT {
		path = append(path, first.Lexeme)
	}
	for p.match(token.DOT) {
		part := p.expect(token.IDENT, "identifier after '.'")
		if part.Kind == token.IDENT {
			path = append(path, part.Lexeme)
		}
	}
	return path
}

func (p *Parser) parseTopLevelDecl() ast.Decl {
	pub := p.match(token.PUB)
	switch {
	case p.check(token.FUNC):
		return p.parseFuncDecl(pub)
	case p.check(token.TYPE):
		return p.parseTypeDecl()
	case p.check(token.CONST):
		return p.parseConstDecl()
	case p.check(token.IDENT):
		if p.checkNext(token.COLON) || p.checkNext(token.EQ) {
			return p.parseVarDecl()
		}
	}

	if pub {
		p.errorAt(p.previous(), "E013", "expected declaration after pub", "use pub before func, type, or const")
	} else {
		p.errorAt(p.peek(), "E013", fmt.Sprintf("expected top-level declaration, got %s", p.peek().Kind), "use func, type, const, import, or a variable declaration")
	}
	if !p.atEnd() {
		p.advance()
	}
	p.synchronize()
	return nil
}

func (p *Parser) parseVarDecl() *ast.VarDecl {
	name := p.expect(token.IDENT, "variable name")
	decl := &ast.VarDecl{NamePos: name.Pos, Name: name.Lexeme}
	if p.match(token.COLON) {
		decl.Type = p.parseType(true)
	}
	if p.match(token.EQ) {
		decl.Value = p.parseExpression(precLowest)
	}
	p.consumeTerminator()
	return decl
}

func (p *Parser) parseConstDecl() *ast.ConstDecl {
	start := p.expect(token.CONST, "const declaration")
	name := p.expect(token.IDENT, "const name")
	decl := &ast.ConstDecl{ConstPos: start.Pos, Name: name.Lexeme}
	if p.match(token.COLON) {
		decl.Type = p.parseType(true)
	}
	if p.match(token.EQ) {
		decl.Value = p.parseExpression(precLowest)
	} else {
		p.errorAt(p.peek(), "E013", "expected '=' in const declaration", "const declarations require a value")
	}
	p.consumeTerminator()
	return decl
}

func (p *Parser) parseFuncDecl(pub bool) *ast.FuncDecl {
	start := p.expect(token.FUNC, "function declaration")
	name := p.expect(token.IDENT, "function name")
	fn := &ast.FuncDecl{FuncPos: start.Pos, Pub: pub, Name: name.Lexeme}
	p.expect(token.LPAREN, "'(' after function name")
	if !p.check(token.RPAREN) && !p.atEnd() {
		for {
			fn.Params = append(fn.Params, p.parseParam())
			if !p.match(token.COMMA) {
				break
			}
		}
	}
	p.expect(token.RPAREN, "')' after parameters")
	if p.match(token.ARROW) {
		fn.ReturnTypes = append(fn.ReturnTypes, p.parseType(true))
		for p.match(token.COMMA) {
			fn.ReturnTypes = append(fn.ReturnTypes, p.parseType(true))
		}
	}
	fn.Body = p.parseBlock()
	return fn
}

func (p *Parser) parseParam() *ast.Param {
	ref := p.match(token.REF)
	name := p.expect(token.IDENT, "parameter name")
	p.expect(token.COLON, "':' after parameter name")
	return &ast.Param{Ref: ref, Name: name.Lexeme, Type: p.parseType(true), Posn: name.Pos}
}

func (p *Parser) parseTypeDecl() *ast.TypeDecl {
	start := p.expect(token.TYPE, "type declaration")
	name := p.expect(token.IDENT, "type name")
	decl := &ast.TypeDecl{TypePos: start.Pos, Name: name.Lexeme}
	if p.check(token.IDENT) && p.peek().Lexeme == "ext" {
		p.advance()
		extends := p.expect(token.IDENT, "base type name")
		decl.Extends = extends.Lexeme
	}

	p.expect(token.COLON, "':' after type declaration")
	p.expect(token.NEWLINE, "newline after ':'")
	p.expect(token.INDENT, "indented type body")
	for !p.check(token.DEDENT) && !p.atEnd() {
		p.skipNewlines()
		if p.check(token.DEDENT) || p.atEnd() {
			break
		}
		pub := p.match(token.PUB)
		if p.check(token.FUNC) {
			decl.Methods = append(decl.Methods, p.parseFuncDecl(pub))
			continue
		}
		if p.check(token.IDENT) && p.peek().Lexeme == "constructor" {
			decl.Constructor = p.parseConstructorDecl()
			continue
		}
		if p.check(token.IDENT) {
			decl.Fields = append(decl.Fields, p.parseFieldDecl(pub))
			continue
		}
		p.errorAt(p.peek(), "E013", fmt.Sprintf("expected type member, got %s", p.peek().Kind), "use a field, constructor, or method")
		p.synchronize()
	}
	p.expect(token.DEDENT, "dedent after type body")
	return decl
}

func (p *Parser) parseFieldDecl(pub bool) *ast.FieldDecl {
	name := p.expect(token.IDENT, "field name")
	field := &ast.FieldDecl{FieldPos: name.Pos, Pub: pub, Name: name.Lexeme}
	p.expect(token.COLON, "':' after field name")
	field.Type = p.parseType(true)
	if p.match(token.EQ) {
		field.Default = p.parseExpression(precLowest)
	}
	p.consumeTerminator()
	return field
}

func (p *Parser) parseConstructorDecl() *ast.ConstructorDecl {
	start := p.expect(token.IDENT, "constructor")
	ctor := &ast.ConstructorDecl{ConstructorPos: start.Pos}
	p.expect(token.LPAREN, "'(' after constructor")
	if !p.check(token.RPAREN) && !p.atEnd() {
		for {
			ctor.Params = append(ctor.Params, p.parseParam())
			if !p.match(token.COMMA) {
				break
			}
		}
	}
	p.expect(token.RPAREN, "')' after constructor parameters")
	ctor.Body = p.parseBlock()
	return ctor
}

func (p *Parser) parseBlock() []ast.Stmt {
	p.expect(token.COLON, "':' before block")
	p.expect(token.NEWLINE, "newline after ':'")
	p.expect(token.INDENT, "indented block")

	var body []ast.Stmt
	for !p.check(token.DEDENT) && !p.atEnd() {
		p.skipNewlines()
		if p.check(token.DEDENT) || p.atEnd() {
			break
		}
		stmt := p.parseStmt()
		if stmt != nil {
			body = append(body, stmt)
		}
	}
	p.expect(token.DEDENT, "dedent after block")
	return body
}

func (p *Parser) parseStmt() ast.Stmt {
	p.skipNewlines()
	switch p.peek().Kind {
	case token.IF:
		return p.parseIfStmt()
	case token.ELSE:
		p.errorAt(p.peek(), "E015", "else without matching if", "place else immediately after an if block")
		p.advance()
		p.synchronize()
		return nil
	case token.SWITCH:
		if p.isSwitchExprAhead() {
			expr := p.parseExpression(precLowest)
			p.consumeTerminator()
			return &ast.ExprStmt{X: expr}
		}
		return p.parseSwitchStmt()
	case token.LOOP:
		return p.parseLoopStmt()
	case token.RETURN:
		return p.parseReturnStmt()
	case token.PASS:
		pass := p.advance()
		p.consumeTerminator()
		return &ast.PassStmt{PassPos: pass.Pos}
	case token.BREAK:
		br := p.advance()
		if p.loopDepth == 0 {
			p.errorAt(br, "E016", "break outside loop", "use break only inside loop blocks")
		}
		p.consumeTerminator()
		return &ast.BreakStmt{BreakPos: br.Pos}
	case token.CONTINUE:
		cont := p.advance()
		if p.loopDepth == 0 {
			p.errorAt(cont, "E016", "continue outside loop", "use continue only inside loop blocks")
		}
		p.consumeTerminator()
		return &ast.ContinueStmt{ContinuePos: cont.Pos}
	case token.CONST:
		return p.parseConstDecl()
	case token.IDENT:
		if p.checkNext(token.COLON) {
			return p.parseVarDecl()
		}
	}
	return p.parseSimpleStmt()
}

func (p *Parser) parseIfStmt() *ast.IfStmt {
	start := p.expect(token.IF, "if statement")
	stmt := &ast.IfStmt{IfPos: start.Pos, Condition: p.parseExpression(precLowest)}
	stmt.Body = p.parseBlock()
	p.skipNewlines()
	for p.match(token.ELSE) {
		elsePos := p.previous().Pos
		if p.match(token.IF) {
			clause := &ast.ElseIfClause{ElseIfPos: elsePos, Condition: p.parseExpression(precLowest)}
			clause.Body = p.parseBlock()
			stmt.ElseIfs = append(stmt.ElseIfs, clause)
			p.skipNewlines()
			continue
		}
		stmt.Else = p.parseBlock()
		break
	}
	return stmt
}

func (p *Parser) parseSwitchStmt() *ast.SwitchStmt {
	start := p.expect(token.SWITCH, "switch statement")
	stmt := &ast.SwitchStmt{SwitchPos: start.Pos, Subject: p.parseExpression(precLowest)}
	p.expect(token.COLON, "':' after switch subject")
	p.expect(token.NEWLINE, "newline after switch ':'")
	p.expect(token.INDENT, "indented switch body")
	for !p.check(token.DEDENT) && !p.atEnd() {
		p.skipNewlines()
		if p.check(token.DEDENT) || p.atEnd() {
			break
		}
		casePos := p.peek().Pos
		patterns := p.parsePatterns()
		body := p.parseBlock()
		stmt.Cases = append(stmt.Cases, &ast.SwitchCase{CasePos: casePos, Patterns: patterns, Body: body})
	}
	if len(stmt.Cases) == 0 {
		p.errorAt(start, "E018", "empty switch", "add at least one switch arm")
	}
	p.expect(token.DEDENT, "dedent after switch")
	return stmt
}

func (p *Parser) parseLoopStmt() *ast.LoopStmt {
	start := p.expect(token.LOOP, "loop statement")
	stmt := &ast.LoopStmt{LoopPos: start.Pos}

	if p.isIteratorStart() {
		stmt.Iterators = append(stmt.Iterators, p.parseLoopIterator())
		for p.match(token.COMMA) {
			if p.isLoopModifier() {
				break
			}
			if !p.isIteratorStart() {
				p.errorAt(p.peek(), "E013", "expected loop iterator or modifier", "use name in iterable, step, while, or if")
				break
			}
			stmt.Iterators = append(stmt.Iterators, p.parseLoopIterator())
		}
	} else {
		iterable := p.parseExpression(precLowest)
		stmt.Iterators = append(stmt.Iterators, &ast.LoopIterator{IterPos: iterable.Pos(), Variable: "_", Iterable: iterable})
	}

	for !p.check(token.COLON) && !p.atEnd() {
		p.match(token.COMMA)
		switch {
		case p.match(token.STEP):
			if stmt.Step != nil {
				p.errorAt(p.previous(), "E017", "duplicate step modifier in loop", "keep only one step modifier")
			}
			stmt.Step = p.parseExpression(precLowest)
		case p.match(token.WHILE):
			if stmt.While != nil {
				p.errorAt(p.previous(), "E017", "duplicate while modifier in loop", "keep only one while modifier")
			}
			stmt.While = p.parseExpression(precLowest)
		case p.match(token.IF):
			if stmt.IfCond != nil {
				p.errorAt(p.previous(), "E017", "duplicate if modifier in loop", "keep only one if modifier")
			}
			stmt.IfCond = p.parseExpression(precLowest)
		default:
			break
		}
		if !p.isLoopModifier() && !p.check(token.COMMA) {
			break
		}
	}

	p.loopDepth++
	stmt.Body = p.parseBlock()
	p.loopDepth--
	return stmt
}

func (p *Parser) parseLoopIterator() *ast.LoopIterator {
	name := p.advance()
	variable := name.Lexeme
	if name.Kind == token.UNDERSCORE {
		variable = "_"
	}
	p.expect(token.IN, "'in' in loop iterator")
	return &ast.LoopIterator{IterPos: name.Pos, Variable: variable, Iterable: p.parseExpression(precLowest)}
}

func (p *Parser) parseReturnStmt() *ast.ReturnStmt {
	start := p.expect(token.RETURN, "return statement")
	stmt := &ast.ReturnStmt{ReturnPos: start.Pos}
	if !p.isTerminator() {
		stmt.Value = p.parseExpression(precLowest)
	}
	p.consumeTerminator()
	return stmt
}

func (p *Parser) parseSimpleStmt() ast.Stmt {
	expr := p.parseExpression(precLowest)
	if p.match(token.EQ) {
		eq := p.previous()
		if !isAssignable(expr) {
			p.errorAt(eq, "E014", "invalid assignment target", "assign to an identifier, field, index, or component reference")
		}
		value := p.parseExpression(precLowest)
		p.consumeTerminator()
		return &ast.AssignStmt{Target: expr, EqPos: eq.Pos, Value: value}
	}
	p.consumeTerminator()
	return &ast.ExprStmt{X: expr}
}

func isAssignable(expr ast.Expr) bool {
	switch expr.(type) {
	case *ast.Ident, *ast.FieldExpr, *ast.IndexExpr, *ast.ComponentRefExpr:
		return true
	default:
		return false
	}
}

func (p *Parser) parsePatterns() []ast.Expr {
	if p.match(token.UNDERSCORE) {
		return nil
	}
	patterns := []ast.Expr{p.parseExpression(precLowest)}
	for p.match(token.PIPE) {
		if p.match(token.UNDERSCORE) {
			patterns = append(patterns, &ast.Ident{NamePos: p.previous().Pos, Name: "_"})
		} else {
			patterns = append(patterns, p.parseExpression(precLowest))
		}
	}
	return patterns
}

func (p *Parser) parseExpression(minPrec int) ast.Expr {
	expr := p.parsePrefix()
	for {
		op := p.peek()
		prec := infixPrecedence(op.Kind)
		if prec <= minPrec {
			break
		}
		p.advance()
		right := p.parseExpression(prec)
		if op.Kind == token.DOT_DOT || op.Kind == token.DOT_DOT_LESS {
			expr = &ast.RangeExpr{Low: expr, High: right, Exclusive: op.Kind == token.DOT_DOT_LESS, OpPos: op.Pos}
		} else {
			expr = &ast.BinaryExpr{Left: expr, Op: op.Kind, OpPos: op.Pos, Right: right}
		}
	}
	return expr
}

func (p *Parser) parsePrefix() ast.Expr {
	switch p.peek().Kind {
	case token.NOT:
		op := p.advance()
		return &ast.UnaryExpr{OpPos: op.Pos, Op: op.Kind, X: p.parseExpression(precNot)}
	case token.MINUS:
		op := p.advance()
		return &ast.UnaryExpr{OpPos: op.Pos, Op: op.Kind, X: p.parseExpression(precUnary)}
	default:
		return p.parsePostfix(p.parsePrimary())
	}
}

func (p *Parser) parsePostfix(expr ast.Expr) ast.Expr {
	for {
		switch {
		case p.match(token.DOT):
			dot := p.previous()
			field := p.expect(token.IDENT, "field name after '.'")
			expr = &ast.FieldExpr{X: expr, Dot: dot.Pos, Field: field.Lexeme}
		case p.match(token.LBRACKET):
			lbrack := p.previous()
			index := p.parseExpression(precLowest)
			rbrack := p.expect(token.RBRACKET, "']' after index")
			expr = &ast.IndexExpr{X: expr, LBrack: lbrack.Pos, Index: index, RBrack: rbrack.Pos}
		case p.match(token.LPAREN):
			expr = p.finishCall(expr, p.previous())
		default:
			return expr
		}
	}
}

func (p *Parser) parsePrimary() ast.Expr {
	tok := p.advance()
	switch tok.Kind {
	case token.INT_LIT, token.FLOAT_LIT, token.STRING_LIT, token.CHAR_LIT, token.TRUE, token.FALSE, token.NULL:
		return &ast.Literal{ValuePos: tok.Pos, Kind: tok.Kind, Value: tok.Lexeme}
	case token.INTERP_STRING_LIT:
		return p.parseInterpolatedString(tok)
	case token.IDENT:
		return &ast.Ident{NamePos: tok.Pos, Name: tok.Lexeme}
	case token.UNDERSCORE:
		return &ast.Ident{NamePos: tok.Pos, Name: "_"}
	case token.LPAREN:
		expr := p.parseExpression(precLowest)
		p.expect(token.RPAREN, "')' after expression")
		return expr
	case token.LBRACKET:
		return p.parseListLiteral(tok)
	case token.LBRACE:
		return p.parseDictLiteral(tok)
	case token.AT:
		return p.parseComponentRef(tok)
	case token.SWITCH:
		return p.parseSwitchExpr(tok)
	default:
		p.errorAt(tok, "E013", fmt.Sprintf("expected expression, got %s", tok.Kind), "use a literal, identifier, or expression")
		return &ast.BadExpr{Start: tok.Pos, Stop: tok.Pos}
	}
}

func (p *Parser) finishCall(fn ast.Expr, lparen token.Token) ast.Expr {
	call := &ast.CallExpr{Func: fn, LParen: lparen.Pos}
	seenNamed := false
	if !p.check(token.RPAREN) && !p.atEnd() {
		for {
			argPos := p.peek().Pos
			name := ""
			if p.check(token.IDENT) && p.checkNext(token.COLON) {
				name = p.advance().Lexeme
				p.advance()
				seenNamed = true
			} else if seenNamed {
				p.errorAt(p.peek(), "E011", "named argument before positional", "put positional arguments before named arguments")
			}
			call.Args = append(call.Args, &ast.Argument{Name: name, Value: p.parseExpression(precLowest), Posn: argPos})
			if !p.match(token.COMMA) {
				break
			}
		}
	}
	rparen := p.expect(token.RPAREN, "')' after arguments")
	call.RParen = rparen.Pos
	return call
}

func (p *Parser) parseListLiteral(start token.Token) ast.Expr {
	lit := &ast.ListLiteral{LBrack: start.Pos}
	p.skipLiteralSeparators()
	if !p.check(token.RBRACKET) {
		for {
			lit.Elements = append(lit.Elements, p.parseExpression(precLowest))
			p.skipLiteralSeparators()
			if !p.match(token.COMMA) {
				break
			}
			p.skipLiteralSeparators()
			if p.check(token.RBRACKET) {
				break
			}
		}
	}
	rbrack := p.expect(token.RBRACKET, "']' after list literal")
	lit.RBrack = rbrack.Pos
	return lit
}

func (p *Parser) parseDictLiteral(start token.Token) ast.Expr {
	lit := &ast.DictLiteral{LBrace: start.Pos}
	p.skipLiteralSeparators()
	if !p.check(token.RBRACE) {
		for {
			key := p.parseExpression(precLowest)
			p.expect(token.COLON, "':' between dictionary key and value")
			value := p.parseExpression(precLowest)
			lit.Pairs = append(lit.Pairs, &ast.DictPair{Key: key, Value: value})
			p.skipLiteralSeparators()
			if !p.match(token.COMMA) {
				break
			}
			p.skipLiteralSeparators()
			if p.check(token.RBRACE) {
				break
			}
		}
	}
	rbrace := p.expect(token.RBRACE, "'}' after dictionary literal")
	lit.RBrace = rbrace.Pos
	return lit
}

func (p *Parser) skipLiteralSeparators() {
	for p.match(token.NEWLINE, token.INDENT, token.DEDENT) {
	}
}

func (p *Parser) parseComponentRef(at token.Token) ast.Expr {
	component := p.expect(token.IDENT, "component type after '@'")
	p.expect(token.LBRACKET, "'[' after component type")
	id := p.expect(token.IDENT, "component id")
	p.expect(token.RBRACKET, "']' after component id")
	return &ast.ComponentRefExpr{AtPos: at.Pos, Component: component.Lexeme, ID: id.Lexeme}
}

func (p *Parser) parseSwitchExpr(start token.Token) ast.Expr {
	expr := &ast.SwitchExpr{SwitchPos: start.Pos, Subject: p.parseExpression(precLowest)}
	p.expect(token.COLON, "':' after switch subject")
	p.expect(token.NEWLINE, "newline after switch ':'")
	p.expect(token.INDENT, "indented switch expression body")
	for !p.check(token.DEDENT) && !p.atEnd() {
		p.skipNewlines()
		if p.check(token.DEDENT) || p.atEnd() {
			break
		}
		armPos := p.peek().Pos
		patterns := p.parsePatterns()
		p.expect(token.ARROW, "'->' in switch expression arm")
		value := p.parseExpression(precLowest)
		expr.Arms = append(expr.Arms, &ast.SwitchArm{ArmPos: armPos, Patterns: patterns, Value: value})
		p.consumeTerminator()
	}
	if len(expr.Arms) == 0 {
		p.errorAt(start, "E018", "empty switch", "add at least one switch arm")
	}
	p.expect(token.DEDENT, "dedent after switch expression")
	return expr
}

func (p *Parser) parseInterpolatedString(tok token.Token) ast.Expr {
	raw := tok.Lexeme
	body := raw
	if strings.HasPrefix(body, "$\"") {
		body = body[2:]
	}
	if strings.HasSuffix(body, "\"") {
		body = body[:len(body)-1]
	}

	expr := &ast.InterpStringExpr{StringPos: tok.Pos}
	var text strings.Builder
	for i := 0; i < len(body); i++ {
		ch := body[i]
		if ch == '\\' {
			if i+1 < len(body) {
				text.WriteByte(ch)
				i++
				text.WriteByte(body[i])
			} else {
				text.WriteByte(ch)
			}
			continue
		}
		if ch != '{' {
			text.WriteByte(ch)
			continue
		}
		if text.Len() > 0 {
			expr.Segments = append(expr.Segments, ast.InterpSegment{Text: text.String()})
			text.Reset()
		}
		start := i + 1
		depth := 1
		j := start
		for ; j < len(body); j++ {
			if body[j] == '\\' {
				j++
				continue
			}
			switch body[j] {
			case '{':
				depth++
			case '}':
				depth--
				if depth == 0 {
					frag := strings.TrimSpace(body[start:j])
					if frag == "" {
						p.errorAt(tok, "E013", "expected expression inside interpolation", "put an expression between braces")
						expr.Segments = append(expr.Segments, ast.InterpSegment{IsExpr: true, Expr: &ast.BadExpr{Start: tok.Pos, Stop: tok.Pos}})
					} else {
						expr.Segments = append(expr.Segments, ast.InterpSegment{IsExpr: true, Expr: p.parseInterpolationExpr(tok, frag)})
					}
					break
				}
			}
			if depth == 0 {
				break
			}
		}
		if depth != 0 {
			p.errorAt(tok, "E012", "unterminated interpolation brace", "add a matching '}'")
			text.WriteString(body[i:])
			break
		}
		i = j
	}
	if text.Len() > 0 {
		expr.Segments = append(expr.Segments, ast.InterpSegment{Text: text.String()})
	}
	return expr
}

func (p *Parser) parseInterpolationExpr(tok token.Token, src string) ast.Expr {
	diag := p.diagnostics
	tokens := lexer.New(tok.Pos.File, src, diag).Tokenize()
	fragment := New(tokens, diag)
	expr := fragment.parseExpression(precLowest)
	if !fragment.check(token.EOF) {
		p.errorAt(tok, "E013", "expected end of interpolation expression", "keep one expression inside interpolation braces")
	}
	return expr
}

func (p *Parser) parseType(inTypeContext bool) ast.TypeExpr {
	start := p.peek()
	name := ""
	switch start.Kind {
	case token.STRING_KW, token.INT_KW, token.FLOAT_KW, token.BOOL_KW, token.CHAR_KW, token.ANY, token.IDENT:
		name = p.advance().Lexeme
	default:
		p.errorAt(start, "E013", fmt.Sprintf("expected type, got %s", start.Kind), "use a type name")
		p.advance()
		return &ast.SimpleType{TypePos: start.Pos, Name: "<error>"}
	}
	for p.match(token.DOT) {
		part := p.expect(token.IDENT, "identifier after '.' in type")
		name += "." + part.Lexeme
	}

	var typ ast.TypeExpr = &ast.SimpleType{TypePos: start.Pos, Name: name}
	if inTypeContext && isGenericName(name) && p.match(token.LESS) {
		generic := &ast.GenericType{TypePos: start.Pos, Name: name}
		if !p.check(token.GREATER) && !p.atEnd() {
			for {
				generic.Params = append(generic.Params, p.parseType(true))
				if !p.match(token.COMMA) {
					break
				}
			}
		}
		p.expect(token.GREATER, "'>' after generic type parameters")
		typ = generic
	}
	if p.match(token.QUESTION) {
		typ = &ast.NullableType{Inner: typ, QPos: p.previous().Pos}
	}
	return typ
}

func isGenericName(name string) bool {
	switch name {
	case "List", "Dict", "Stack", "Queue":
		return true
	default:
		return false
	}
}

func infixPrecedence(kind token.TokenKind) int {
	switch kind {
	case token.OR:
		return precOr
	case token.AND:
		return precAnd
	case token.EQ_EQ, token.BANG_EQ, token.LESS, token.LESS_EQ, token.GREATER, token.GREATER_EQ:
		return precCompare
	case token.DOT_DOT, token.DOT_DOT_LESS:
		return precRange
	case token.PLUS, token.MINUS:
		return precAdd
	case token.STAR, token.SLASH, token.PERCENT:
		return precMul
	default:
		return precLowest
	}
}

func (p *Parser) isIteratorStart() bool {
	return p.isLoopVariableToken(p.peek().Kind) && p.checkNext(token.IN)
}

func (p *Parser) isLoopVariableToken(kind token.TokenKind) bool {
	switch kind {
	case token.IDENT, token.UNDERSCORE, token.STRING_KW, token.INT_KW, token.FLOAT_KW, token.BOOL_KW, token.CHAR_KW, token.ANY:
		return true
	default:
		return false
	}
}

func (p *Parser) isLoopModifier() bool {
	return p.check(token.STEP) || p.check(token.WHILE) || p.check(token.IF)
}

func (p *Parser) isSwitchExprAhead() bool {
	i := p.pos + 1
	depth := 0
	for i < len(p.tokens) {
		kind := p.tokens[i].Kind
		switch kind {
		case token.LPAREN, token.LBRACKET, token.LBRACE:
			depth++
		case token.RPAREN, token.RBRACKET, token.RBRACE:
			if depth > 0 {
				depth--
			}
		case token.COLON:
			if depth == 0 {
				goto headerDone
			}
		case token.EOF:
			return false
		}
		i++
	}
	return false

headerDone:
	i++
	if i >= len(p.tokens) || p.tokens[i].Kind != token.NEWLINE {
		return false
	}
	i++
	if i >= len(p.tokens) || p.tokens[i].Kind != token.INDENT {
		return false
	}
	i++
	depth = 0
	for i < len(p.tokens) {
		kind := p.tokens[i].Kind
		switch kind {
		case token.LPAREN, token.LBRACKET, token.LBRACE:
			depth++
		case token.RPAREN, token.RBRACKET, token.RBRACE:
			if depth > 0 {
				depth--
			}
		case token.ARROW:
			if depth == 0 {
				return true
			}
		case token.COLON:
			if depth == 0 {
				return false
			}
		case token.NEWLINE, token.DEDENT, token.EOF:
			return false
		}
		i++
	}
	return false
}

func (p *Parser) consumeTerminator() {
	if p.match(token.NEWLINE) {
		return
	}
	if p.check(token.DEDENT) || p.check(token.EOF) {
		return
	}
}

func (p *Parser) isTerminator() bool {
	return p.check(token.NEWLINE) || p.check(token.DEDENT) || p.check(token.EOF)
}

func (p *Parser) skipNewlines() {
	for p.match(token.NEWLINE) {
	}
}

func (p *Parser) synchronize() {
	for !p.atEnd() {
		if p.previous().Kind == token.NEWLINE {
			return
		}
		switch p.peek().Kind {
		case token.NEWLINE:
			p.advance()
			return
		case token.DEDENT, token.FUNC, token.TYPE, token.PACKAGE, token.IMPORT, token.CONST, token.PUB:
			return
		}
		p.advance()
	}
}

func (p *Parser) expect(kind token.TokenKind, what string) token.Token {
	if p.check(kind) {
		return p.advance()
	}
	tok := p.peek()
	p.errorAt(tok, "E013", fmt.Sprintf("expected %s, got %s", what, tok.Kind), "check the syntax near this token")
	return tok
}

func (p *Parser) errorAt(tok token.Token, code, msg, hint string) {
	span := len(tok.Lexeme)
	if span < 1 {
		span = 1
	}
	p.diagnostics.ReportError(code, msg, tok.Pos, span, hint, "")
}

func (p *Parser) match(kinds ...token.TokenKind) bool {
	for _, kind := range kinds {
		if p.check(kind) {
			p.advance()
			return true
		}
	}
	return false
}

func (p *Parser) check(kind token.TokenKind) bool {
	if p.atEnd() && kind != token.EOF {
		return false
	}
	return p.peek().Kind == kind
}

func (p *Parser) checkNext(kind token.TokenKind) bool {
	if p.pos+1 >= len(p.tokens) {
		return false
	}
	return p.tokens[p.pos+1].Kind == kind
}

func (p *Parser) advance() token.Token {
	if !p.atEnd() {
		p.pos++
	}
	return p.previous()
}

func (p *Parser) atEnd() bool {
	return p.peek().Kind == token.EOF
}

func (p *Parser) peek() token.Token {
	if p.pos >= len(p.tokens) {
		return p.tokens[len(p.tokens)-1]
	}
	return p.tokens[p.pos]
}

func (p *Parser) previous() token.Token {
	if p.pos == 0 {
		return p.tokens[0]
	}
	return p.tokens[p.pos-1]
}
