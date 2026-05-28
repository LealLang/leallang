// Copyright (c) 2026 LealLang Contributors
// SPDX-License-Identifier: MIT

package interpreter

// Env is a runtime lexical environment.
type Env struct {
	parent *Env
	vars   map[string]*cell
}

type cell struct {
	value     Value
	constBind bool
}

func NewEnv(parent *Env) *Env {
	return &Env{parent: parent, vars: make(map[string]*cell)}
}

func (e *Env) Set(name string, val Value) {
	e.setCell(name, &cell{value: val})
}

func (e *Env) setConst(name string, val Value) {
	e.setCell(name, &cell{value: val, constBind: true})
}

func (e *Env) setCell(name string, c *cell) {
	e.vars[name] = c
}

func (e *Env) Assign(name string, val Value) bool {
	c, ok := e.lookupCell(name)
	if !ok || c.constBind {
		return false
	}
	c.value = val
	return true
}

func (e *Env) Get(name string) (Value, bool) {
	c, ok := e.lookupCell(name)
	if !ok {
		return nil, false
	}
	return c.value, true
}

func (e *Env) lookupCell(name string) (*cell, bool) {
	for cur := e; cur != nil; cur = cur.parent {
		if c, ok := cur.vars[name]; ok {
			return c, true
		}
	}
	return nil, false
}

func (e *Env) assignCell(name string, source *cell) bool {
	c, ok := e.lookupCell(name)
	if !ok || c.constBind {
		return false
	}
	c.value = source.value
	return true
}

func (e *Env) bindAlias(name string, source *cell) {
	e.vars[name] = source
}
