// Copyright (c) 2026 LealLang Contributors
// SPDX-License-Identifier: MIT

// Package uiir defines a stable UI intermediate representation.
// The UI IR is independent of the AST, parser, checker, and any desktop framework.
package uiir

// OpKind distinguishes UI IR operations.
type OpKind int

const (
	// OpMount creates and attaches a new component node.
	OpMount OpKind = iota
	// OpPropSet sets a property value on a component.
	OpPropSet
	// OpEventBind binds an event handler to a component.
	OpEventBind
)

// Op is a single UI operation.
type Op struct {
	Kind      OpKind
	Component string // "Window", "Button", etc. (for OpMount)
	ID        string // "main", "save_btn" (for all ops)
	ParentID  string // "" for root (for OpMount)
	PropName  string // "title", "w", etc. (for OpPropSet)
	PropValue any    // Go-native value (string, int, bool, etc.) (for OpPropSet)
	EventName string // "click", "hover", etc. (for OpEventBind)
	HandlerID string // handler function name (for OpEventBind)
}

// Log is an ordered list of UI operations.
type Log struct {
	Ops []Op
}

// Record appends an operation to the log.
func (l *Log) Record(op Op) {
	l.Ops = append(l.Ops, op)
}
