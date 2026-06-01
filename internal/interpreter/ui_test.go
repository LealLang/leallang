// Copyright (c) 2026 LealLang Contributors
// SPDX-License-Identifier: MIT

package interpreter

import (
	"testing"

	"github.com/LealLang/leallang/internal/diagnostics"
	"github.com/LealLang/leallang/internal/lexer"
	"github.com/LealLang/leallang/internal/parser"
	"github.com/LealLang/leallang/internal/uiir"
)

func runWithUIBackend(t *testing.T, src string) (*uiir.Log, *diagnostics.Diagnostics) {
	t.Helper()
	diag := diagnostics.New()
	tokens := lexer.New("test.ll", src, diag).Tokenize()
	program := parser.New(tokens, diag).Parse()
	if diag.HasErrors() {
		t.Fatalf("parse errors:\n%s", diag.Format())
	}

	interp := New(diag)
	log := &uiir.Log{}
	interp.SetUIBackend(log)
	interp.SetOutput(discardWriter(nil), discardWriter(nil))
	interp.registerBuiltins()

	if err := interp.Run(program); err != nil {
		t.Fatalf("run error: %v", err)
	}
	return log, diag
}

func TestUIMountOrder(t *testing.T) {
	log, diag := runWithUIBackend(t, `package app.main

ui func view():
    Window[main]:
        title = "App"

        Col[root]:
            Label[title]:
                text = "Hello"

            Button[save_btn]:
                text = "Save"

func main():
    view()
`)
	if diag.HasErrors() {
		t.Fatalf("unexpected errors:\n%s", diag.Format())
	}

	// Expected mount order: Window, Col, Label, Button (depth-first)
	expected := []struct {
		component string
		id        string
		parentID  string
	}{
		{"Window", "main", ""},
		{"Col", "root", "main"},
		{"Label", "title", "root"},
		{"Button", "save_btn", "root"},
	}

	mounts := filterOps(log, uiir.OpMount)
	if len(mounts) != len(expected) {
		t.Fatalf("mount count = %d, want %d", len(mounts), len(expected))
	}
	for i, want := range expected {
		if mounts[i].Component != want.component {
			t.Fatalf("mount[%d].Component = %q, want %q", i, mounts[i].Component, want.component)
		}
		if mounts[i].ID != want.id {
			t.Fatalf("mount[%d].ID = %q, want %q", i, mounts[i].ID, want.id)
		}
		if mounts[i].ParentID != want.parentID {
			t.Fatalf("mount[%d].ParentID = %q, want %q", i, mounts[i].ParentID, want.parentID)
		}
	}
}

func TestUIPropValues(t *testing.T) {
	log, diag := runWithUIBackend(t, `package app.main

ui func view():
    Window[main]:
        title = "LealLang App"
        w = 800
        h = 600

func main():
    view()
`)
	if diag.HasErrors() {
		t.Fatalf("unexpected errors:\n%s", diag.Format())
	}

	props := filterOps(log, uiir.OpPropSet)
	if len(props) != 3 {
		t.Fatalf("prop count = %d, want 3", len(props))
	}

	// Check title prop
	titleProp := findProp(props, "main", "title")
	if titleProp == nil {
		t.Fatal("title prop not found")
	}
	if titleProp.PropValue != "LealLang App" {
		t.Fatalf("title = %v, want 'LealLang App'", titleProp.PropValue)
	}

	// Check w prop
	wProp := findProp(props, "main", "w")
	if wProp == nil {
		t.Fatal("w prop not found")
	}
	if wProp.PropValue != int64(800) {
		t.Fatalf("w = %v, want 800", wProp.PropValue)
	}

	// Check h prop
	hProp := findProp(props, "main", "h")
	if hProp == nil {
		t.Fatal("h prop not found")
	}
	if hProp.PropValue != int64(600) {
		t.Fatalf("h = %v, want 600", hProp.PropValue)
	}
}

func TestUIEventBindings(t *testing.T) {
	log, diag := runWithUIBackend(t, `package app.main

func handle_save():
    pass

ui func view():
    Button[save_btn]:
        text = "Save"
        on click = handle_save

func main():
    view()
`)
	if diag.HasErrors() {
		t.Fatalf("unexpected errors:\n%s", diag.Format())
	}

	events := filterOps(log, uiir.OpEventBind)
	if len(events) != 1 {
		t.Fatalf("event count = %d, want 1", len(events))
	}
	if events[0].ID != "save_btn" {
		t.Fatalf("event ID = %q, want save_btn", events[0].ID)
	}
	if events[0].EventName != "click" {
		t.Fatalf("event name = %q, want click", events[0].EventName)
	}
	if events[0].HandlerID != "handle_save" {
		t.Fatalf("handler = %q, want handle_save", events[0].HandlerID)
	}
}

func TestUIComponentRefAssignment(t *testing.T) {
	log, diag := runWithUIBackend(t, `package app.main

func handle_save():
    @Label[title].text = "Saved"

ui func view():
    Window[main]:
        title = "App"

        Label[title]:
            text = "Hello"

        Button[save_btn]:
            text = "Save"
            on click = handle_save

func main():
    view()
    handle_save()
`)
	if diag.HasErrors() {
		t.Fatalf("unexpected errors:\n%s", diag.Format())
	}

	// Find the prop set for the component ref assignment
	props := filterOps(log, uiir.OpPropSet)
	found := false
	for _, op := range props {
		if op.ID == "title" && op.PropName == "text" && op.PropValue == "Saved" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("component ref assignment not recorded")
	}
}

func TestUIBackwardCompatibility(t *testing.T) {
	// Existing tests should still pass without a UI backend set.
	diag := diagnostics.New()
	tokens := lexer.New("test.ll", `package app.main

func main():
    x = 42
`, diag).Tokenize()
	program := parser.New(tokens, diag).Parse()
	if diag.HasErrors() {
		t.Fatalf("parse errors:\n%s", diag.Format())
	}

	interp := New(diag)
	interp.SetOutput(discardWriter(nil), discardWriter(nil))
	interp.registerBuiltins()

	if err := interp.Run(program); err != nil {
		t.Fatalf("run error: %v", err)
	}
}

// Helper functions

func filterOps(log *uiir.Log, kind uiir.OpKind) []uiir.Op {
	var result []uiir.Op
	for _, op := range log.Ops {
		if op.Kind == kind {
			result = append(result, op)
		}
	}
	return result
}

func findProp(props []uiir.Op, id, name string) *uiir.Op {
	for i, op := range props {
		if op.ID == id && op.PropName == name {
			return &props[i]
		}
	}
	return nil
}
