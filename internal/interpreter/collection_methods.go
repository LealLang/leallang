// Copyright (c) 2026 LealLang Contributors
// SPDX-License-Identifier: MIT

package interpreter

import (
	"fmt"

	"github.com/LealLang/leallang/internal/token"
)

func (interp *Interpreter) listMethod(receiver *ListVal, name string, pos token.Position) (Value, error) {
	switch name {
	case "push":
		return &BuiltinVal{Name: "List.push", Fn: func(args []Value) (Value, error) {
			if len(args) != 1 {
				return nil, interp.runtimeError("E105", pos, 0, "List.push expects 1 argument", "")
			}
			receiver.Elements = append(receiver.Elements, args[0])
			return Null, nil
		}}, nil
	case "pop":
		return &BuiltinVal{Name: "List.pop", Fn: func(args []Value) (Value, error) {
			if len(args) != 0 {
				return nil, interp.runtimeError("E105", pos, 0, "List.pop expects 0 arguments", "")
			}
			if len(receiver.Elements) == 0 {
				return nil, interp.runtimeError("E102", pos, 0, "cannot pop from an empty list", "")
			}
			last := receiver.Elements[len(receiver.Elements)-1]
			receiver.Elements = receiver.Elements[:len(receiver.Elements)-1]
			return last, nil
		}}, nil
	case "insert":
		return &BuiltinVal{Name: "List.insert", Fn: func(args []Value) (Value, error) {
			if len(args) != 2 {
				return nil, interp.runtimeError("E105", pos, 0, "List.insert expects 2 arguments", "")
			}
			index, ok := intArg(args[0])
			if !ok {
				return nil, interp.runtimeError("E101", pos, 0, "List.insert index must be int", "")
			}
			if index < 0 || index > len(receiver.Elements) {
				return nil, interp.runtimeError("E102", pos, 0, "list index out of bounds", "")
			}
			receiver.Elements = append(receiver.Elements, Null)
			copy(receiver.Elements[index+1:], receiver.Elements[index:])
			receiver.Elements[index] = args[1]
			return Null, nil
		}}, nil
	case "remove":
		return &BuiltinVal{Name: "List.remove", Fn: func(args []Value) (Value, error) {
			if len(args) != 1 {
				return nil, interp.runtimeError("E105", pos, 0, "List.remove expects 1 argument", "")
			}
			for i, item := range receiver.Elements {
				if valuesEqual(item, args[0]) {
					receiver.Elements = append(receiver.Elements[:i], receiver.Elements[i+1:]...)
					return BoolVal(true), nil
				}
			}
			return BoolVal(false), nil
		}}, nil
	case "index_of":
		return &BuiltinVal{Name: "List.index_of", Fn: func(args []Value) (Value, error) {
			if len(args) != 1 {
				return nil, interp.runtimeError("E105", pos, 0, "List.index_of expects 1 argument", "")
			}
			for i, item := range receiver.Elements {
				if valuesEqual(item, args[0]) {
					return IntVal(i), nil
				}
			}
			return IntVal(-1), nil
		}}, nil
	case "has_index":
		return &BuiltinVal{Name: "List.has_index", Fn: func(args []Value) (Value, error) {
			if len(args) != 1 {
				return nil, interp.runtimeError("E105", pos, 0, "List.has_index expects 1 argument", "")
			}
			index, ok := intArg(args[0])
			if !ok {
				return nil, interp.runtimeError("E101", pos, 0, "List.has_index index must be int", "")
			}
			return BoolVal(index >= 0 && index < len(receiver.Elements)), nil
		}}, nil
	case "count":
		return &BuiltinVal{Name: "List.count", Fn: func(args []Value) (Value, error) {
			if len(args) != 0 {
				return nil, interp.runtimeError("E105", pos, 0, "List.count expects 0 arguments", "")
			}
			return IntVal(len(receiver.Elements)), nil
		}}, nil
	case "clear":
		return &BuiltinVal{Name: "List.clear", Fn: func(args []Value) (Value, error) {
			if len(args) != 0 {
				return nil, interp.runtimeError("E105", pos, 0, "List.clear expects 0 arguments", "")
			}
			receiver.Elements = receiver.Elements[:0]
			return Null, nil
		}}, nil
	default:
		return nil, interp.runtimeError("E101", pos, 0, fmt.Sprintf("List has no method '%s'", name), "")
	}
}

func (interp *Interpreter) dictMethod(receiver *DictVal, name string, pos token.Position) (Value, error) {
	switch name {
	case "has_key":
		return &BuiltinVal{Name: "Dict.has_key", Fn: func(args []Value) (Value, error) {
			key, ok, err := stringKeyArg(interp, args, pos, "Dict.has_key")
			if err != nil || !ok {
				return nil, err
			}
			_, exists := receiver.Entries[key]
			return BoolVal(exists), nil
		}}, nil
	case "try_add":
		return &BuiltinVal{Name: "Dict.try_add", Fn: func(args []Value) (Value, error) {
			key, ok, err := stringKeyValueArgs(interp, args, pos, "Dict.try_add")
			if err != nil || !ok {
				return nil, err
			}
			if _, exists := receiver.Entries[key]; exists {
				return BoolVal(false), nil
			}
			receiver.Entries[key] = args[1]
			return BoolVal(true), nil
		}}, nil
	case "try_set":
		return &BuiltinVal{Name: "Dict.try_set", Fn: func(args []Value) (Value, error) {
			key, ok, err := stringKeyValueArgs(interp, args, pos, "Dict.try_set")
			if err != nil || !ok {
				return nil, err
			}
			if _, exists := receiver.Entries[key]; !exists {
				return BoolVal(false), nil
			}
			receiver.Entries[key] = args[1]
			return BoolVal(true), nil
		}}, nil
	case "try_remove":
		return &BuiltinVal{Name: "Dict.try_remove", Fn: func(args []Value) (Value, error) {
			key, ok, err := stringKeyArg(interp, args, pos, "Dict.try_remove")
			if err != nil || !ok {
				return nil, err
			}
			if _, exists := receiver.Entries[key]; !exists {
				return BoolVal(false), nil
			}
			delete(receiver.Entries, key)
			return BoolVal(true), nil
		}}, nil
	case "count":
		return &BuiltinVal{Name: "Dict.count", Fn: func(args []Value) (Value, error) {
			if len(args) != 0 {
				return nil, interp.runtimeError("E105", pos, 0, "Dict.count expects 0 arguments", "")
			}
			return IntVal(len(receiver.Entries)), nil
		}}, nil
	case "clear":
		return &BuiltinVal{Name: "Dict.clear", Fn: func(args []Value) (Value, error) {
			if len(args) != 0 {
				return nil, interp.runtimeError("E105", pos, 0, "Dict.clear expects 0 arguments", "")
			}
			clear(receiver.Entries)
			return Null, nil
		}}, nil
	case "get":
		return &BuiltinVal{Name: "Dict.get", Fn: func(args []Value) (Value, error) {
			key, ok, err := stringKeyArg(interp, args, pos, "Dict.get")
			if err != nil || !ok {
				return nil, err
			}
			if value, exists := receiver.Entries[key]; exists {
				return &TupleVal{Elements: []Value{value, Null}}, nil
			}
			return &TupleVal{Elements: []Value{Null, errorRecord("key not found: "+key, "KeyError")}}, nil
		}}, nil
	default:
		return nil, interp.runtimeError("E101", pos, 0, fmt.Sprintf("Dict has no method '%s'", name), "")
	}
}

func intArg(value Value) (int, bool) {
	i, ok := value.(IntVal)
	return int(i), ok
}

func stringKeyArg(interp *Interpreter, args []Value, pos token.Position, name string) (string, bool, error) {
	if len(args) != 1 {
		return "", false, interp.runtimeError("E105", pos, 0, name+" expects 1 argument", "")
	}
	key, ok := args[0].(StringVal)
	if !ok {
		return "", false, interp.runtimeError("E101", pos, 0, name+" key must be string", "")
	}
	return string(key), true, nil
}

func stringKeyValueArgs(interp *Interpreter, args []Value, pos token.Position, name string) (string, bool, error) {
	if len(args) != 2 {
		return "", false, interp.runtimeError("E105", pos, 0, name+" expects 2 arguments", "")
	}
	key, ok := args[0].(StringVal)
	if !ok {
		return "", false, interp.runtimeError("E101", pos, 0, name+" key must be string", "")
	}
	return string(key), true, nil
}
