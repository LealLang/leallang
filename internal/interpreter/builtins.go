// Copyright (c) 2026 LealLang Contributors
// SPDX-License-Identifier: MIT

package interpreter

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

func (interp *Interpreter) registerBuiltins() {
	interp.globals.Set("console", &NamespaceVal{Name: "console", Members: map[string]*BuiltinVal{
		"print": {
			Name: "console.print",
			Fn: func(args []Value) (Value, error) {
				if len(args) > 0 {
					fmt.Fprint(interp.stdout, args[0].String())
				}
				return Null, nil
			},
		},
		"print_ln": {
			Name: "console.print_ln",
			Fn: func(args []Value) (Value, error) {
				if len(args) > 0 {
					fmt.Fprintln(interp.stdout, args[0].String())
				} else {
					fmt.Fprintln(interp.stdout)
				}
				return Null, nil
			},
		},
	}})

	interp.globals.Set("file", &NamespaceVal{Name: "file", Members: map[string]*BuiltinVal{
		"read_text": {
			Name: "file.read_text",
			Fn: func(args []Value) (Value, error) {
				path, ok := firstString(args)
				if !ok {
					return &TupleVal{Elements: []Value{StringVal(""), errorRecord("file.read_text expects a string path", "TypeError")}}, nil
				}
				data, err := os.ReadFile(path)
				if err != nil {
					return &TupleVal{Elements: []Value{StringVal(""), errorRecord(err.Error(), "IOError")}}, nil
				}
				return &TupleVal{Elements: []Value{StringVal(string(data)), Null}}, nil
			},
		},
		"write_text": {
			Name: "file.write_text",
			Fn: func(args []Value) (Value, error) {
				if len(args) < 2 {
					return errorRecord("file.write_text expects path and value", "TypeError"), nil
				}
				path, ok1 := args[0].(StringVal)
				value, ok2 := args[1].(StringVal)
				if !ok1 || !ok2 {
					return errorRecord("file.write_text expects string arguments", "TypeError"), nil
				}
				if err := os.WriteFile(string(path), []byte(value), 0o644); err != nil {
					return errorRecord(err.Error(), "IOError"), nil
				}
				return Null, nil
			},
		},
		"exists": {
			Name: "file.exists",
			Fn: func(args []Value) (Value, error) {
				path, ok := firstString(args)
				if !ok {
					return BoolVal(false), nil
				}
				_, err := os.Stat(path)
				return BoolVal(err == nil), nil
			},
		},
	}})

	interp.globals.Set("json", &NamespaceVal{Name: "json", Members: map[string]*BuiltinVal{
		"parse": {
			Name: "json.parse",
			Fn: func(args []Value) (Value, error) {
				text, ok := firstString(args)
				if !ok {
					return &TupleVal{Elements: []Value{Null, errorRecord("json.parse expects a string", "TypeError")}}, nil
				}
				val, err := decodeJSONValue([]byte(text))
				if err != nil {
					return &TupleVal{Elements: []Value{Null, errorRecord(err.Error(), "JSONError")}}, nil
				}
				return &TupleVal{Elements: []Value{val, Null}}, nil
			},
		},
		"stringify": {
			Name: "json.stringify",
			Fn: func(args []Value) (Value, error) {
				if len(args) == 0 {
					return &TupleVal{Elements: []Value{StringVal(""), errorRecord("json.stringify expects a value", "TypeError")}}, nil
				}
				data, err := json.Marshal(valueToGo(args[0]))
				if err != nil {
					return &TupleVal{Elements: []Value{StringVal(""), errorRecord(err.Error(), "JSONError")}}, nil
				}
				return &TupleVal{Elements: []Value{StringVal(string(data)), Null}}, nil
			},
		},
	}})

	interp.globals.Set("system", &NamespaceVal{Name: "system", Members: map[string]*BuiltinVal{
		"exit": {
			Name: "system.exit",
			Fn: func(args []Value) (Value, error) {
				code := int64(0)
				if len(args) > 0 {
					if v, ok := args[0].(IntVal); ok {
						code = int64(v)
					}
				}
				return Null, exitError{code: int(code)}
			},
		},
		"args": {
			Name: "system.args",
			Fn: func(args []Value) (Value, error) {
				items := make([]Value, 0, len(interp.args))
				for _, arg := range interp.args {
					items = append(items, StringVal(arg))
				}
				return &ListVal{Elements: items}, nil
			},
		},
		"env": {
			Name: "system.env",
			Fn: func(args []Value) (Value, error) {
				key, ok := firstString(args)
				if !ok {
					return Null, nil
				}
				value, found := os.LookupEnv(key)
				if !found {
					return Null, nil
				}
				return StringVal(value), nil
			},
		},
	}})

	interp.registerStubNamespace("window", "open", "close")
	interp.registerStubNamespace("msg", "alert", "error", "info")
	interp.registerStubNamespace("modal", "open", "close")
	interp.registerStubNamespace("toast", "show")
	interp.registerStubNamespace("clipboard", "copy", "paste", "has_text")
	interp.registerStubNamespace("screen", "get_width", "get_height")
	interp.registerConstGroups()
}

func (interp *Interpreter) registerStubNamespace(name string, members ...string) {
	ns := &NamespaceVal{Name: name, Members: make(map[string]*BuiltinVal, len(members))}
	for _, member := range members {
		fullName := name + "." + member
		ns.Members[member] = &BuiltinVal{Name: fullName, Fn: func(args []Value) (Value, error) {
			fmt.Fprintf(interp.stderr, "[stub] %s\n", fullName)
			return Null, nil
		}}
	}
	interp.globals.Set(name, ns)
}

func (interp *Interpreter) registerConstGroups() {
	groups := map[string]struct {
		typeName string
		members  []string
	}{
		"colors":      {"Color", []string{"blue", "red", "white", "black", "purple", "green", "yellow", "orange", "gray"}},
		"dock":        {"Dock", []string{"left", "right", "center", "top", "bottom", "fill"}},
		"orientation": {"Orientation", []string{"horizontal", "vertical"}},
		"toast_type":  {"ToastType", []string{"error", "success"}},
		"font_weight": {"FontWeight", []string{"normal", "bold", "light"}},
		"text_align":  {"TextAlign", []string{"left", "right", "center"}},
		"scroll_mode": {"ScrollMode", []string{"vertical", "horizontal", "both"}},
		"sort_order":  {"SortOrder", []string{"ascending", "descending"}},
	}
	for name, group := range groups {
		members := make(map[string]*EnumVal, len(group.members))
		for _, member := range group.members {
			members[member] = &EnumVal{TypeName: group.typeName, Member: member}
		}
		interp.globals.Set(name, &ConstGroupVal{Name: name, Members: members})
	}
}

func firstString(args []Value) (string, bool) {
	if len(args) == 0 {
		return "", false
	}
	v, ok := args[0].(StringVal)
	return string(v), ok
}

type exitError struct {
	code int
}

func (e exitError) Error() string {
	return fmt.Sprintf("exit %d", e.code)
}

func discardWriter(w io.Writer) io.Writer {
	if w != nil {
		return w
	}
	return io.Discard
}
