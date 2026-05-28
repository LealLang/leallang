// Copyright (c) 2026 LealLang Contributors
// SPDX-License-Identifier: MIT

package checker

import "github.com/LealLang/leallang/internal/token"

// SymbolKind distinguishes symbol categories.
type SymbolKind int

const (
	SymVar       SymbolKind = iota
	SymConst
	SymFunc
	SymType
	SymParam
	SymNamespace
	SymConstGroup
)

// Symbol represents a named entity in a scope.
type Symbol struct {
	Name string
	Type Type
	Kind SymbolKind
	Pos  token.Position
	Pub  bool
	Ref  bool // for ref parameters
}

// Scope represents a lexical scope with a parent chain.
type Scope struct {
	Parent  *Scope
	symbols map[string]*Symbol
}

// NewScope creates a child scope.
func NewScope(parent *Scope) *Scope {
	return &Scope{Parent: parent, symbols: make(map[string]*Symbol)}
}

// Define adds a symbol to this scope. Returns an error message if already defined.
func (s *Scope) Define(sym *Symbol) string {
	if _, exists := s.symbols[sym.Name]; exists {
		return sym.Name
	}
	s.symbols[sym.Name] = sym
	return ""
}

// Lookup searches the scope chain for a symbol by name.
func (s *Scope) Lookup(name string) *Symbol {
	for scope := s; scope != nil; scope = scope.Parent {
		if sym, ok := scope.symbols[name]; ok {
			return sym
		}
	}
	return nil
}

// LookupLocal searches only this scope for a symbol by name.
func (s *Scope) LookupLocal(name string) *Symbol {
	return s.symbols[name]
}
