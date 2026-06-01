// Copyright (c) 2026 LealLang Contributors
// SPDX-License-Identifier: MIT

package uiir

// FakeBackend records UI operations for testing.
// It provides a headless implementation of UI operations that can be
// inspected in tests without requiring a real desktop framework.
type FakeBackend struct {
	log *Log
}

// New creates a new FakeBackend.
func New() *FakeBackend {
	return &FakeBackend{log: &Log{}}
}

// Log returns the operation log for inspection.
func (fb *FakeBackend) Log() *Log {
	return fb.log
}

// Mount records a component mount operation.
func (fb *FakeBackend) Mount(component, id, parentID string) {
	fb.log.Record(Op{Kind: OpMount, Component: component, ID: id, ParentID: parentID})
}

// PropSet records a property set operation.
func (fb *FakeBackend) PropSet(id, propName string, propValue any) {
	fb.log.Record(Op{Kind: OpPropSet, ID: id, PropName: propName, PropValue: propValue})
}

// EventBind records an event bind operation.
func (fb *FakeBackend) EventBind(id, eventName, handlerID string) {
	fb.log.Record(Op{Kind: OpEventBind, ID: id, EventName: eventName, HandlerID: handlerID})
}
