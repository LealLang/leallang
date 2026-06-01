// Copyright (c) 2026 LealLang Contributors
// SPDX-License-Identifier: MIT

package interpreter

import (
	"strings"
	"testing"

	"github.com/LealLang/leallang/internal/ast"
)

func TestJsonRoundTrip(t *testing.T) {
	out := runSource(t, `package app.main

func main():
    result = json.stringify({"name": "test", "value": "42"})
    console.print_ln(result)
`)
	if !strings.Contains(out, `"name"`) || !strings.Contains(out, `"value"`) {
		t.Fatalf("unexpected output %q", out)
	}
}

// --- TaskVal unit tests ---

func TestValueTypeAndString(t *testing.T) {
	tests := []struct {
		name     string
		val      Value
		wantType string
		wantStr  string
	}{
		{"IntVal", IntVal(42), "int", "42"},
		{"IntVal neg", IntVal(-5), "int", "-5"},
		{"IntVal zero", IntVal(0), "int", "0"},
		{"FloatVal", FloatVal(3.14), "float", "3.14"},
		{"FloatVal zero", FloatVal(0), "float", "0"},
		{"StringVal", StringVal("hello"), "string", "hello"},
		{"StringVal empty", StringVal(""), "string", ""},
		{"BoolVal true", BoolVal(true), "bool", "true"},
		{"BoolVal false", BoolVal(false), "bool", "false"},
		{"CharVal", CharVal('x'), "char", "x"},
		{"CharVal zero", CharVal(0), "char", "\x00"},
		{"NullVal", Null, "null", "null"},
		{"ListVal", &ListVal{Elements: []Value{IntVal(1), IntVal(2)}}, "List", "[1, 2]"},
		{"ListVal empty", &ListVal{}, "List", "[]"},
		{"DictVal", &DictVal{Entries: map[string]Value{"a": IntVal(1)}}, "Dict", `{"a": 1}`},
		{"DictVal empty", &DictVal{}, "Dict", "{}"},
		{"RecordVal", &RecordVal{TypeName: "Point", Fields: map[string]Value{"X": IntVal(1)}}, "Point", "Point{X: 1}"},
		{"EnumVal", &EnumVal{TypeName: "Color", Member: "red"}, "Color", "red"},
		{"TupleVal", &TupleVal{Elements: []Value{IntVal(1), StringVal("two")}}, "tuple", "(1, two)"},
		{"TupleVal empty", &TupleVal{}, "tuple", "()"},
		{"RangeVal", &RangeVal{Low: 0, High: 10, Exclusive: true}, "range", "0..<10"},
		{"RangeVal inclusive", &RangeVal{Low: 0, High: 10, Exclusive: false}, "range", "0..10"},
		{"FuncVal", &FuncVal{Name: "test"}, "function", "<func test>"},
		{"FuncVal anon", &FuncVal{}, "function", "<func>"},
		{"BuiltinVal", &BuiltinVal{Name: "print_ln"}, "builtin", "<builtin print_ln>"},
		{"NamespaceVal", &NamespaceVal{Name: "console"}, "namespace", "<namespace console>"},
		{"ConstGroupVal", &ConstGroupVal{Name: "colors"}, "const group", "<const group colors>"},
		{"RecordTypeVal", &RecordTypeVal{Decl: &ast.TypeDecl{Name: "T"}}, "type", "<type T>"},
		{"BoundMethodVal", &BoundMethodVal{Decl: &ast.FuncDecl{Name: "Foo"}}, "method", "<method Foo>"},
		{"ComponentRefVal", &ComponentRefVal{Component: "Button", ID: "btn"}, "Button", "@Button[btn]"},
		{"TaskVal", newTaskVal(), "Task", "<task>"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.val.Type(); got != tt.wantType {
				t.Errorf("Type() = %q, want %q", got, tt.wantType)
			}
			if got := tt.val.String(); got != tt.wantStr {
				t.Errorf("String() = %q, want %q", got, tt.wantStr)
			}
		})
	}
}

func TestValueStringNil(t *testing.T) {
	if got := valueString(nil); got != "null" {
		t.Errorf("valueString(nil) = %q, want %q", got, "null")
	}
}

// --- valuesEqual coverage ---

func TestValuesEqualAllTypes(t *testing.T) {
	tests := []struct {
		name string
		a, b Value
		want bool
	}{
		{"int equal", IntVal(1), IntVal(1), true},
		{"int not equal", IntVal(1), IntVal(2), false},
		{"float equal", FloatVal(1.5), FloatVal(1.5), true},
		{"float not equal", FloatVal(1.5), FloatVal(2.5), false},
		{"string equal", StringVal("a"), StringVal("a"), true},
		{"string not equal", StringVal("a"), StringVal("b"), false},
		{"bool equal", BoolVal(true), BoolVal(true), true},
		{"bool not equal", BoolVal(true), BoolVal(false), false},
		{"char equal", CharVal('a'), CharVal('a'), true},
		{"char not equal", CharVal('a'), CharVal('b'), false},
		{"null equal", Null, Null, true},
		{"enum equal", &EnumVal{TypeName: "C", Member: "a"}, &EnumVal{TypeName: "C", Member: "a"}, true},
		{"enum different member", &EnumVal{TypeName: "C", Member: "a"}, &EnumVal{TypeName: "C", Member: "b"}, false},
		{"enum different type", &EnumVal{TypeName: "C", Member: "a"}, &EnumVal{TypeName: "D", Member: "a"}, false},
		{"list equal", &ListVal{[]Value{IntVal(1)}}, &ListVal{[]Value{IntVal(1)}}, true},
		{"list different len", &ListVal{[]Value{IntVal(1)}}, &ListVal{[]Value{IntVal(1), IntVal(2)}}, false},
		{"list different values", &ListVal{[]Value{IntVal(1)}}, &ListVal{[]Value{IntVal(2)}}, false},
		{"dict equal", &DictVal{map[string]Value{"a": IntVal(1)}}, &DictVal{map[string]Value{"a": IntVal(1)}}, true},
		{"dict different len", &DictVal{map[string]Value{"a": IntVal(1)}}, &DictVal{map[string]Value{"a": IntVal(1), "b": IntVal(2)}}, false},
		{"dict different values", &DictVal{map[string]Value{"a": IntVal(1)}}, &DictVal{map[string]Value{"a": IntVal(2)}}, false},
		{"record equal", &RecordVal{"T", map[string]Value{"x": IntVal(1)}}, &RecordVal{"T", map[string]Value{"x": IntVal(1)}}, true},
		{"record different type", &RecordVal{"T", map[string]Value{"x": IntVal(1)}}, &RecordVal{"U", map[string]Value{"x": IntVal(1)}}, false},
		{"record different fields", &RecordVal{"T", map[string]Value{"x": IntVal(1)}}, &RecordVal{"T", map[string]Value{"x": IntVal(2)}}, false},
		{"tuple equal", &TupleVal{[]Value{IntVal(1), StringVal("a")}}, &TupleVal{[]Value{IntVal(1), StringVal("a")}}, true},
		{"tuple different len", &TupleVal{[]Value{IntVal(1)}}, &TupleVal{[]Value{IntVal(1), IntVal(2)}}, false},
		{"tuple different values", &TupleVal{[]Value{IntVal(1)}}, &TupleVal{[]Value{IntVal(2)}}, false},
		{"component ref equal", &ComponentRefVal{"Button", "btn"}, &ComponentRefVal{"Button", "btn"}, true},
		{"component ref different", &ComponentRefVal{"Button", "btn"}, &ComponentRefVal{"Label", "btn"}, false},
		{"cross type int float", IntVal(1), FloatVal(1.0), false},
		{"cross type int string", IntVal(1), StringVal("1"), false},
		{"cross type null int", Null, IntVal(0), false},
		{"default fallback diff ptr", &RangeVal{0, 10, false}, &RangeVal{0, 10, false}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := valuesEqual(tt.a, tt.b); got != tt.want {
				t.Errorf("valuesEqual(%v, %v) = %v, want %v", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

// --- dictKey coverage ---

func TestDictKeyAllTypes(t *testing.T) {
	tests := []struct {
		name    string
		val     Value
		wantKey string
		wantOk  bool
	}{
		{"string", StringVal("hello"), "hello", true},
		{"int", IntVal(42), "42", true},
		{"bool true", BoolVal(true), "true", true},
		{"bool false", BoolVal(false), "false", true},
		{"char", CharVal('x'), "x", true},
		{"enum", &EnumVal{TypeName: "C", Member: "red"}, "red", true},
		{"list invalid", &ListVal{}, "", false},
		{"dict invalid", &DictVal{}, "", false},
		{"null invalid", Null, "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, ok := dictKey(tt.val)
			if ok != tt.wantOk {
				t.Errorf("dictKey(%v) ok = %v, want %v", tt.val, ok, tt.wantOk)
			}
			if ok && key != tt.wantKey {
				t.Errorf("dictKey(%v) key = %q, want %q", tt.val, key, tt.wantKey)
			}
		})
	}
}

// --- valueToGo coverage ---

func TestValueToGoAllTypes(t *testing.T) {
	tests := []struct {
		name string
		val  Value
	}{
		{"int", IntVal(42)},
		{"float", FloatVal(3.14)},
		{"string", StringVal("hello")},
		{"bool", BoolVal(true)},
		{"char", CharVal('x')},
		{"null", Null},
		{"list", &ListVal{[]Value{IntVal(1), IntVal(2)}}},
		{"dict", &DictVal{map[string]Value{"a": IntVal(1)}}},
		{"record", &RecordVal{"T", map[string]Value{"x": IntVal(1)}}},
		{"enum", &EnumVal{"Color", "red"}},
		{"tuple", &TupleVal{[]Value{IntVal(1), StringVal("a")}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := valueToGo(tt.val)
			if result == nil && tt.val != Null {
				t.Errorf("valueToGo(%v) returned nil", tt.val)
			}
		})
	}
}

// --- goToValue coverage ---

func TestGoToValueAllTypes(t *testing.T) {
	tests := []struct {
		name string
		val  any
		want string // expected Type()
	}{
		{"nil", nil, "null"},
		{"bool", true, "bool"},
		{"string", "hello", "string"},
		{"float64 int", float64(42), "int"},
		{"float64 float", float64(3.14), "float"},
		{"[]any", []any{float64(1), float64(2)}, "List"},
		{"map", map[string]any{"a": float64(1)}, "Dict"},
		{"default", struct{}{}, "string"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := goToValue(tt.val)
			if got := result.Type(); got != tt.want {
				t.Errorf("goToValue(%v).Type() = %q, want %q", tt.val, got, tt.want)
			}
		})
	}
}

// --- decodeJSONValue / jsonNumberToValue coverage ---

func TestDecodeJSONValue(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantType string
	}{
		{"null", "null", "null"},
		{"bool", "true", "bool"},
		{"string", `"hello"`, "string"},
		{"int", "42", "int"},
		{"float", "3.14", "float"},
		{"array", "[1, 2, 3]", "List"},
		{"object", `{"a": 1}`, "Dict"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := decodeJSONValue([]byte(tt.input))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got := result.Type(); got != tt.wantType {
				t.Errorf("decodeJSONValue(%q).Type() = %q, want %q", tt.input, got, tt.wantType)
			}
		})
	}
}

func TestJsonNumberToValueNested(t *testing.T) {
	// Nested array
	result := jsonNumberToValue([]any{float64(1), float64(2)})
	lst, ok := result.(*ListVal)
	if !ok || len(lst.Elements) != 2 {
		t.Fatalf("expected ListVal with 2 elements, got %T", result)
	}
	// Nested object
	result = jsonNumberToValue(map[string]any{"x": float64(1)})
	dict, ok := result.(*DictVal)
	if !ok || len(dict.Entries) != 1 {
		t.Fatalf("expected DictVal with 1 entry, got %T", result)
	}
	// Default fallback
	result = jsonNumberToValue(struct{}{})
	if result.Type() != "string" {
		t.Errorf("expected string type for unknown, got %q", result.Type())
	}
}

// --- Signal.Error coverage ---

func TestSignalError(t *testing.T) {
	tests := []struct {
		name   string
		signal *Signal
		want   string
	}{
		{"return", &Signal{Kind: signalReturn, Value: IntVal(42)}, "return 42"},
		{"return nil", &Signal{Kind: signalReturn, Value: nil}, "return null"},
		{"break", &Signal{Kind: signalBreak}, "break"},
		{"continue", &Signal{Kind: signalContinue}, "continue"},
		{"unknown", &Signal{Kind: 99}, "signal"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.signal.Error(); got != tt.want {
				t.Errorf("Signal.Error() = %q, want %q", got, tt.want)
			}
		})
	}
}

// --- Source-level tests for uncovered interpreter paths ---

func TestDiscardWriterNil(t *testing.T) {
	// Test that discardWriter returns io.Discard when given nil.
	w := discardWriter(nil)
	if w == nil {
		t.Fatal("expected io.Discard, got nil")
	}
	_, err := w.Write([]byte("test"))
	if err != nil {
		t.Fatalf("unexpected error writing to discard: %v", err)
	}
}

