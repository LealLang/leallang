// Copyright (c) 2026 LealLang Contributors
// SPDX-License-Identifier: MIT

package checker

func collectionMethodSignature(receiver *GenericType, method string, global *Scope) *FuncSignature {
	switch receiver.Name {
	case "List":
		return listMethodSignature(receiver, method)
	case "Dict":
		return dictMethodSignature(receiver, method, global)
	default:
		return nil
	}
}

func listMethodSignature(receiver *GenericType, method string) *FuncSignature {
	if len(receiver.Params) != 1 {
		return nil
	}
	elem := receiver.Params[0]
	switch method {
	case "push":
		return methodSig(method, nil, param("value", elem))
	case "pop":
		return methodSig(method, elem)
	case "insert":
		return methodSig(method, nil, param("index", IntType), param("value", elem))
	case "remove":
		return methodSig(method, BoolType, param("value", elem))
	case "index_of":
		return methodSig(method, IntType, param("value", elem))
	case "has_index":
		return methodSig(method, BoolType, param("index", IntType))
	case "count":
		return methodSig(method, IntType)
	case "clear":
		return methodSig(method, nil)
	default:
		return nil
	}
}

func dictMethodSignature(receiver *GenericType, method string, global *Scope) *FuncSignature {
	if len(receiver.Params) != 2 {
		return nil
	}
	key := receiver.Params[0]
	value := receiver.Params[1]
	switch method {
	case "has_key":
		return methodSig(method, BoolType, param("key", key))
	case "try_add":
		return methodSig(method, BoolType, param("key", key), param("value", value))
	case "try_set":
		return methodSig(method, BoolType, param("key", key), param("value", value))
	case "try_remove":
		return methodSig(method, BoolType, param("key", key))
	case "count":
		return methodSig(method, IntType)
	case "clear":
		return methodSig(method, nil)
	case "get":
		return methodSig(method, &TupleType{Elements: []Type{value, nullableError(global)}}, param("key", key))
	default:
		return nil
	}
}

func methodSig(name string, ret Type, params ...*ParamInfo) *FuncSignature {
	return &FuncSignature{Name: name, Params: params, ReturnType: ret}
}

func param(name string, typ Type) *ParamInfo {
	return &ParamInfo{Name: name, Type: typ}
}

func nullableError(global *Scope) Type {
	if sym := global.Lookup("Error"); sym != nil {
		return &NullableType{Inner: sym.Type}
	}
	return &NullableType{Inner: AnyType}
}
