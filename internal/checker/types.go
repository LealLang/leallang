// Copyright (c) 2026 LealLang Contributors
// SPDX-License-Identifier: MIT

package checker

// Type represents a resolved LealLang type.
type Type interface {
	typeKey() string
	String() string
}

// PrimitiveType represents string, int, float, bool, char, any.
type PrimitiveType struct {
	Name string
}

func (t *PrimitiveType) typeKey() string { return t.Name }
func (t *PrimitiveType) String() string  { return t.Name }

// NullType represents the null literal's type.
type NullType struct{}

func (t *NullType) typeKey() string { return "null" }
func (t *NullType) String() string  { return "null" }

// Void represents the absence of a type (void functions).
type Void struct{}

func (t *Void) typeKey() string { return "_void" }
func (t *Void) String() string  { return "void" }

// NullableType wraps an inner type with ?.
type NullableType struct {
	Inner Type
}

func (t *NullableType) typeKey() string { return t.Inner.typeKey() + "?" }
func (t *NullableType) String() string  { return t.Inner.String() + "?" }

// TupleType represents multiple return values.
type TupleType struct {
	Elements []Type
}

func (t *TupleType) typeKey() string {
	s := "tuple<"
	for i, el := range t.Elements {
		if i > 0 {
			s += ","
		}
		s += el.typeKey()
	}
	return s + ">"
}

func (t *TupleType) String() string {
	s := "("
	for i, el := range t.Elements {
		if i > 0 {
			s += ", "
		}
		s += el.String()
	}
	return s + ")"
}

// FieldInfo carries resolved field metadata.
type FieldInfo struct {
	Name    string
	Type    Type
	Default bool
	Pub     bool
}

// ParamInfo carries resolved parameter metadata.
type ParamInfo struct {
	Name string
	Type Type
	Ref  bool
}

// FuncSignature represents a function's type signature.
type FuncSignature struct {
	Name       string
	Params     []*ParamInfo
	ReturnType Type // nil for void, TupleType for multi-return
	Pub        bool
	Async      bool
	UI         bool
}

func (t *FuncSignature) typeKey() string {
	s := "func"
	if t.Async {
		s += ":async"
	}
	if t.UI {
		s += ":ui"
	}
	s += "(" + paramKeys(t.Params) + ")"
	if t.ReturnType != nil {
		s += "->" + t.ReturnType.typeKey()
	}
	return s
}

func (t *FuncSignature) String() string {
	prefix := "func"
	if t.Async {
		prefix = "async func"
	}
	if t.UI {
		prefix = "ui func"
	}
	s := prefix + " " + t.Name + "(" + paramString(t.Params) + ")"
	if t.ReturnType != nil {
		s += " -> " + t.ReturnType.String()
	}
	return s
}

func paramKeys(params []*ParamInfo) string {
	s := ""
	for i, p := range params {
		if i > 0 {
			s += ","
		}
		if p.Ref {
			s += "ref "
		}
		s += p.Type.typeKey()
	}
	return s
}

func paramString(params []*ParamInfo) string {
	s := ""
	for i, p := range params {
		if i > 0 {
			s += ", "
		}
		if p.Ref {
			s += "ref "
		}
		s += p.Name + ": " + p.Type.String()
	}
	return s
}

// RecordType represents a user-defined type (type Foo: ...).
type RecordType struct {
	Name    string
	Extends string
	Fields  []*FieldInfo
	Methods map[string]*FuncSignature
	Ctor    *FuncSignature
	Pub     bool
}

func (t *RecordType) typeKey() string { return t.Name }
func (t *RecordType) String() string  { return t.Name }

// GenericType represents List<T>, Dict<K,V>, etc.
type GenericType struct {
	Name   string
	Params []Type
}

func (t *GenericType) typeKey() string {
	s := t.Name + "<"
	for i, p := range t.Params {
		if i > 0 {
			s += ","
		}
		s += p.typeKey()
	}
	return s + ">"
}

func (t *GenericType) String() string {
	s := t.Name + "<"
	for i, p := range t.Params {
		if i > 0 {
			s += ", "
		}
		s += p.String()
	}
	return s + ">"
}

// NamespaceType represents a built-in namespace like console, file, etc.
type NamespaceType struct {
	Name    string
	Members map[string]*FuncSignature
	Consts  map[string]Type
}

func (t *NamespaceType) typeKey() string { return "_ns:" + t.Name }
func (t *NamespaceType) String() string  { return t.Name }

// ComponentType represents a UI component type.
type ComponentType struct {
	Name       string
	Primary    *ParamInfo
	Properties map[string]*ParamInfo
	Events     map[string]Type
}

func (t *ComponentType) typeKey() string { return t.Name }
func (t *ComponentType) String() string  { return t.Name }

// IsAssignable reports whether a value of type 'from' can be assigned to a target of type 'to'.
func IsAssignable(from, to Type) bool {
	if from == nil || to == nil {
		return false
	}
	// Exact match
	if from.typeKey() == to.typeKey() {
		return true
	}
	// null is only assignable to nullable types
	if _, ok := from.(*NullType); ok {
		_, toNull := to.(*NullableType)
		return toNull
	}
	// T is assignable to T?
	if _, toNull := to.(*NullableType); toNull {
		return IsAssignable(from, to.(*NullableType).Inner)
	}
	// Concrete types ARE assignable to any (storing a value in any).
	if _, toAny := to.(*PrimitiveType); toAny && to.(*PrimitiveType).Name == "any" {
		return true
	}
	// any is NOT assignable to concrete types without narrowing.
	if _, fromAny := from.(*PrimitiveType); fromAny && from.(*PrimitiveType).Name == "any" {
		return false
	}
	return false
}

// VoidType is the shared void sentinel.
var VoidType = &Void{}

// Predefined primitive types.
var (
	StringType = &PrimitiveType{Name: "string"}
	IntType    = &PrimitiveType{Name: "int"}
	FloatType  = &PrimitiveType{Name: "float"}
	BoolType   = &PrimitiveType{Name: "bool"}
	CharType   = &PrimitiveType{Name: "char"}
	AnyType    = &PrimitiveType{Name: "any"}
	NullType_  = &NullType{}
)

func primitiveByName(name string) Type {
	switch name {
	case "string":
		return StringType
	case "int":
		return IntType
	case "float":
		return FloatType
	case "bool":
		return BoolType
	case "char":
		return CharType
	case "any":
		return AnyType
	default:
		return nil
	}
}

// FormatType returns a human-readable string for a type, used in error messages.
func FormatType(t Type) string {
	if t == nil {
		return "<unknown>"
	}
	return t.String()
}

