// Copyright (c) 2026 LealLang Contributors
// SPDX-License-Identifier: MIT

package interpreter

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/LealLang/leallang/internal/ast"
)

// Value is a runtime LealLang value.
type Value interface {
	Type() string
	String() string
}

type IntVal int64

func (v IntVal) Type() string   { return "int" }
func (v IntVal) String() string { return strconv.FormatInt(int64(v), 10) }

type FloatVal float64

func (v FloatVal) Type() string { return "float" }
func (v FloatVal) String() string {
	return strconv.FormatFloat(float64(v), 'f', -1, 64)
}

type StringVal string

func (v StringVal) Type() string   { return "string" }
func (v StringVal) String() string { return string(v) }

type BoolVal bool

func (v BoolVal) Type() string { return "bool" }
func (v BoolVal) String() string {
	if v {
		return "true"
	}
	return "false"
}

type CharVal rune

func (v CharVal) Type() string { return "char" }
func (v CharVal) String() string {
	if v == 0 {
		return "\x00"
	}
	return string(rune(v))
}

type NullVal struct{}

func (v *NullVal) Type() string   { return "null" }
func (v *NullVal) String() string { return "null" }

var Null = &NullVal{}

type ListVal struct {
	Elements []Value
}

func (v *ListVal) Type() string { return "List" }
func (v *ListVal) String() string {
	parts := make([]string, 0, len(v.Elements))
	for _, el := range v.Elements {
		parts = append(parts, valueString(el))
	}
	return "[" + strings.Join(parts, ", ") + "]"
}

type DictVal struct {
	Entries map[string]Value
}

func (v *DictVal) Type() string { return "Dict" }
func (v *DictVal) String() string {
	keys := make([]string, 0, len(v.Entries))
	for k := range v.Entries {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%q: %s", k, valueString(v.Entries[k])))
	}
	return "{" + strings.Join(parts, ", ") + "}"
}

type RecordVal struct {
	TypeName string
	Fields   map[string]Value
}

func (v *RecordVal) Type() string { return v.TypeName }
func (v *RecordVal) String() string {
	keys := make([]string, 0, len(v.Fields))
	for k := range v.Fields {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, k+": "+valueString(v.Fields[k]))
	}
	return v.TypeName + "{" + strings.Join(parts, ", ") + "}"
}

type EnumVal struct {
	TypeName string
	Member   string
}

func (v *EnumVal) Type() string   { return v.TypeName }
func (v *EnumVal) String() string { return v.Member }

type TupleVal struct {
	Elements []Value
}

func (v *TupleVal) Type() string { return "tuple" }
func (v *TupleVal) String() string {
	parts := make([]string, 0, len(v.Elements))
	for _, el := range v.Elements {
		parts = append(parts, valueString(el))
	}
	return "(" + strings.Join(parts, ", ") + ")"
}

type FuncVal struct {
	Name    string
	Params  []*ast.Param
	Body    []ast.Stmt
	Closure *Env
	Async   bool
	UI      bool
}

func (v *FuncVal) Type() string { return "function" }
func (v *FuncVal) String() string {
	if v.Name != "" {
		return "<func " + v.Name + ">"
	}
	return "<func>"
}

type BuiltinVal struct {
	Name string
	Fn   func(args []Value) (Value, error)
}

func (v *BuiltinVal) Type() string   { return "builtin" }
func (v *BuiltinVal) String() string { return "<builtin " + v.Name + ">" }

type NamespaceVal struct {
	Name    string
	Members map[string]*BuiltinVal
}

func (v *NamespaceVal) Type() string   { return "namespace" }
func (v *NamespaceVal) String() string { return "<namespace " + v.Name + ">" }

type ConstGroupVal struct {
	Name    string
	Members map[string]*EnumVal
}

func (v *ConstGroupVal) Type() string   { return "const group" }
func (v *ConstGroupVal) String() string { return "<const group " + v.Name + ">" }

type RangeVal struct {
	Low       int64
	High      int64
	Exclusive bool
}

func (v *RangeVal) Type() string { return "range" }
func (v *RangeVal) String() string {
	op := ".."
	if v.Exclusive {
		op = "..<"
	}
	return fmt.Sprintf("%d%s%d", v.Low, op, v.High)
}

type RecordTypeVal struct {
	Decl *ast.TypeDecl
}

func (v *RecordTypeVal) Type() string   { return "type" }
func (v *RecordTypeVal) String() string { return "<type " + v.Decl.Name + ">" }

type BoundMethodVal struct {
	Receiver *RecordVal
	Decl     *ast.FuncDecl
	Closure  *Env
}

func (v *BoundMethodVal) Type() string   { return "method" }
func (v *BoundMethodVal) String() string { return "<method " + v.Decl.Name + ">" }

type ComponentRefVal struct {
	Component string
	ID        string
}

func (v *ComponentRefVal) Type() string   { return v.Component }
func (v *ComponentRefVal) String() string { return "@" + v.Component + "[" + v.ID + "]" }

// TaskVal represents an async task that may still be running.
type TaskVal struct {
	result Value
	err    Value // Null on success, Error record on failure
	done   chan struct{}
}

func newTaskVal() *TaskVal {
	return &TaskVal{done: make(chan struct{})}
}

func (v *TaskVal) Type() string { return "Task" }
func (v *TaskVal) String() string {
	return "<task>"
}

// resolve sets the result and error, then signals completion.
func (v *TaskVal) resolve(result, err Value) {
	v.result = result
	v.err = err
	close(v.done)
}

// await blocks until the task completes and returns (result, err).
func (v *TaskVal) await() (Value, Value) {
	<-v.done
	return v.result, v.err
}

func valueString(v Value) string {
	if v == nil {
		return "null"
	}
	return v.String()
}

func valuesEqual(a, b Value) bool {
	switch av := a.(type) {
	case IntVal:
		bv, ok := b.(IntVal)
		return ok && av == bv
	case FloatVal:
		bv, ok := b.(FloatVal)
		return ok && av == bv
	case StringVal:
		bv, ok := b.(StringVal)
		return ok && av == bv
	case BoolVal:
		bv, ok := b.(BoolVal)
		return ok && av == bv
	case CharVal:
		bv, ok := b.(CharVal)
		return ok && av == bv
	case *NullVal:
		_, ok := b.(*NullVal)
		return ok
	case *EnumVal:
		bv, ok := b.(*EnumVal)
		return ok && av.TypeName == bv.TypeName && av.Member == bv.Member
	case *ListVal:
		bv, ok := b.(*ListVal)
		if !ok || len(av.Elements) != len(bv.Elements) {
			return false
		}
		for i := range av.Elements {
			if !valuesEqual(av.Elements[i], bv.Elements[i]) {
				return false
			}
		}
		return true
	case *DictVal:
		bv, ok := b.(*DictVal)
		if !ok || len(av.Entries) != len(bv.Entries) {
			return false
		}
		for k, v := range av.Entries {
			if !valuesEqual(v, bv.Entries[k]) {
				return false
			}
		}
		return true
	case *RecordVal:
		bv, ok := b.(*RecordVal)
		if !ok || av.TypeName != bv.TypeName || len(av.Fields) != len(bv.Fields) {
			return false
		}
		for k, v := range av.Fields {
			if !valuesEqual(v, bv.Fields[k]) {
				return false
			}
		}
		return true
	case *TupleVal:
		bv, ok := b.(*TupleVal)
		if !ok || len(av.Elements) != len(bv.Elements) {
			return false
		}
		for i := range av.Elements {
			if !valuesEqual(av.Elements[i], bv.Elements[i]) {
				return false
			}
		}
		return true
	case *ComponentRefVal:
		bv, ok := b.(*ComponentRefVal)
		return ok && av.Component == bv.Component && av.ID == bv.ID
	default:
		return a == b
	}
}

func dictKey(v Value) (string, bool) {
	switch key := v.(type) {
	case StringVal:
		return string(key), true
	case IntVal:
		return key.String(), true
	case BoolVal:
		return key.String(), true
	case CharVal:
		return key.String(), true
	case *EnumVal:
		return key.Member, true
	default:
		return "", false
	}
}

func valueToGo(v Value) any {
	switch val := v.(type) {
	case IntVal:
		return int64(val)
	case FloatVal:
		return float64(val)
	case StringVal:
		return string(val)
	case BoolVal:
		return bool(val)
	case CharVal:
		return string(rune(val))
	case *NullVal:
		return nil
	case *ListVal:
		out := make([]any, 0, len(val.Elements))
		for _, el := range val.Elements {
			out = append(out, valueToGo(el))
		}
		return out
	case *DictVal:
		out := make(map[string]any, len(val.Entries))
		for k, el := range val.Entries {
			out[k] = valueToGo(el)
		}
		return out
	case *RecordVal:
		out := make(map[string]any, len(val.Fields))
		for k, el := range val.Fields {
			out[k] = valueToGo(el)
		}
		return out
	case *EnumVal:
		return val.Member
	case *TupleVal:
		out := make([]any, 0, len(val.Elements))
		for _, el := range val.Elements {
			out = append(out, valueToGo(el))
		}
		return out
	default:
		return val.String()
	}
}

func goToValue(v any) Value {
	switch val := v.(type) {
	case nil:
		return Null
	case bool:
		return BoolVal(val)
	case string:
		return StringVal(val)
	case float64:
		if val == float64(int64(val)) {
			return IntVal(int64(val))
		}
		return FloatVal(val)
	case []any:
		items := make([]Value, 0, len(val))
		for _, item := range val {
			items = append(items, goToValue(item))
		}
		return &ListVal{Elements: items}
	case map[string]any:
		items := make(map[string]Value, len(val))
		for k, item := range val {
			items[k] = goToValue(item)
		}
		return &DictVal{Entries: items}
	default:
		return StringVal(fmt.Sprint(val))
	}
}

func decodeJSONValue(data []byte) (Value, error) {
	decoder := json.NewDecoder(strings.NewReader(string(data)))
	decoder.UseNumber()
	var raw any
	if err := decoder.Decode(&raw); err != nil {
		return Null, err
	}
	return jsonNumberToValue(raw), nil
}

func jsonNumberToValue(v any) Value {
	switch val := v.(type) {
	case nil:
		return Null
	case bool:
		return BoolVal(val)
	case string:
		return StringVal(val)
	case json.Number:
		if i, err := val.Int64(); err == nil {
			return IntVal(i)
		}
		f, err := val.Float64()
		if err != nil {
			return StringVal(val.String())
		}
		return FloatVal(f)
	case []any:
		items := make([]Value, 0, len(val))
		for _, item := range val {
			items = append(items, jsonNumberToValue(item))
		}
		return &ListVal{Elements: items}
	case map[string]any:
		items := make(map[string]Value, len(val))
		for k, item := range val {
			items[k] = jsonNumberToValue(item)
		}
		return &DictVal{Entries: items}
	default:
		return goToValue(val)
	}
}

// cloneForTask deep-copies sendable values for crossing async task boundaries.
// Non-sendable values return an error.
func cloneForTask(v Value) (Value, error) {
	switch val := v.(type) {
	// Primitives and immutable values — safe to share.
	case IntVal, FloatVal, StringVal, BoolVal, CharVal, *NullVal, *EnumVal, *RangeVal:
		return v, nil

	// Pointer-backed containers — deep-copy recursively.
	case *TupleVal:
		elems := make([]Value, len(val.Elements))
		for i, el := range val.Elements {
			cloned, err := cloneForTask(el)
			if err != nil {
				return nil, err
			}
			elems[i] = cloned
		}
		return &TupleVal{Elements: elems}, nil

	case *ListVal:
		elems := make([]Value, len(val.Elements))
		for i, el := range val.Elements {
			cloned, err := cloneForTask(el)
			if err != nil {
				return nil, err
			}
			elems[i] = cloned
		}
		return &ListVal{Elements: elems}, nil

	case *DictVal:
		entries := make(map[string]Value, len(val.Entries))
		for k, el := range val.Entries {
			cloned, err := cloneForTask(el)
			if err != nil {
				return nil, err
			}
			entries[k] = cloned
		}
		return &DictVal{Entries: entries}, nil

	case *RecordVal:
		fields := make(map[string]Value, len(val.Fields))
		for k, el := range val.Fields {
			cloned, err := cloneForTask(el)
			if err != nil {
				return nil, err
			}
			fields[k] = cloned
		}
		return &RecordVal{TypeName: val.TypeName, Fields: fields}, nil

	// Non-sendable values — reject.
	case *FuncVal, *BuiltinVal, *NamespaceVal, *ConstGroupVal, *RecordTypeVal, *BoundMethodVal, *ComponentRefVal, *TaskVal:
		return nil, fmt.Errorf("value of type %s cannot cross async task boundary", v.Type())

	default:
		return nil, fmt.Errorf("value of type %s cannot cross async task boundary", v.Type())
	}
}

func errorRecord(message, code string) *RecordVal {
	return &RecordVal{
		TypeName: "Error",
		Fields: map[string]Value{
			"message": StringVal(message),
			"code":    StringVal(code),
		},
	}
}

// valueToUI converts a LealLang runtime value to a Go-native value
// suitable for the UI IR.
func valueToUI(v Value) any {
	switch val := v.(type) {
	case IntVal:
		return int64(val)
	case FloatVal:
		return float64(val)
	case StringVal:
		return string(val)
	case BoolVal:
		return bool(val)
	case CharVal:
		return rune(val)
	case *EnumVal:
		return val.Member
	default:
		return nil
	}
}
