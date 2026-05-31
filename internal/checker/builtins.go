// Copyright (c) 2026 LealLang Contributors
// SPDX-License-Identifier: MIT

package checker

// RegisterBuiltins populates the global scope with built-in types, functions,
// namespaces, and component types.
func RegisterBuiltins(global *Scope) {
	registerPrimitives(global)
	registerErrorType(global)
	registerEventTypes(global)
	registerEnumTypes(global)
	registerNamespaces(global)
	registerConstGroups(global)
	registerComponentTypes(global)
}

func registerPrimitives(global *Scope) {
	prims := []struct {
		name string
		typ  Type
	}{
		{"string", StringType},
		{"int", IntType},
		{"float", FloatType},
		{"bool", BoolType},
		{"char", CharType},
		{"any", AnyType},
	}
	for _, p := range prims {
		global.Define(&Symbol{Name: p.name, Type: p.typ, Kind: SymType, Pub: true})
	}
	// Register Task as a generic type sentinel.
	global.Define(&Symbol{Name: "Task", Type: &GenericType{Name: "Task", Params: []Type{AnyType}}, Kind: SymType, Pub: true})
}

func registerErrorType(global *Scope) {
	errorType := &RecordType{
		Name: "Error",
		Fields: []*FieldInfo{
			{Name: "message", Type: StringType, Pub: true},
			{Name: "code", Type: StringType, Pub: true},
		},
		Methods: make(map[string]*FuncSignature),
		Pub:     true,
	}
	global.Define(&Symbol{Name: "Error", Type: errorType, Kind: SymType, Pub: true})
}

func registerEventTypes(global *Scope) {
	events := []struct {
		name   string
		fields []*FieldInfo
	}{
		{"ClickEvent", []*FieldInfo{
			{Name: "mouse_x", Type: IntType},
			{Name: "mouse_y", Type: IntType},
			{Name: "button", Type: IntType},
		}},
		{"HoverEvent", nil},
		{"LeaveEvent", nil},
		{"ScrollEvent", []*FieldInfo{
			{Name: "scroll_y", Type: FloatType},
			{Name: "delta_y", Type: FloatType},
		}},
		{"ResizeEvent", []*FieldInfo{
			{Name: "width", Type: IntType},
			{Name: "height", Type: IntType},
		}},
		{"FocusEvent", []*FieldInfo{
			{Name: "source", Type: StringType},
		}},
		{"ChangeEvent", nil},
	}
	for _, e := range events {
		et := &RecordType{
			Name:    e.name,
			Fields:  e.fields,
			Methods: make(map[string]*FuncSignature),
			Pub:     true,
		}
		global.Define(&Symbol{Name: e.name, Type: et, Kind: SymType, Pub: true})
	}
}

func registerEnumTypes(global *Scope) {
	enums := []struct {
		name    string
		members []string
	}{
		{"Color", []string{"blue", "red", "white", "black", "purple", "green", "yellow", "orange", "gray"}},
		{"Dock", []string{"left", "right", "center", "top", "bottom", "fill"}},
		{"Orientation", []string{"horizontal", "vertical"}},
		{"ToastType", []string{"error", "success"}},
		{"FontWeight", []string{"normal", "bold", "light"}},
		{"TextAlign", []string{"left", "right", "center"}},
		{"ScrollMode", []string{"vertical", "horizontal", "both"}},
		{"SortOrder", []string{"ascending", "descending"}},
	}
	for _, e := range enums {
		t := &EnumType{Name: e.name, Members: e.members}
		global.Define(&Symbol{Name: e.name, Type: t, Kind: SymType, Pub: true})
	}
}

func registerNamespaces(global *Scope) {
	// console
	console := &NamespaceType{
		Name: "console",
		Members: map[string]*FuncSignature{
			"print":    {Name: "print", Params: []*ParamInfo{{Name: "value", Type: AnyType}}},
			"print_ln": {Name: "print_ln", Params: []*ParamInfo{{Name: "value", Type: AnyType}}},
		},
	}
	global.Define(&Symbol{Name: "console", Type: console, Kind: SymNamespace, Pub: true})

	// file
	fileNs := &NamespaceType{
		Name: "file",
		Members: map[string]*FuncSignature{
			"read_text": {
				Name: "read_text",
				Params: []*ParamInfo{{Name: "path", Type: StringType}},
				ReturnType: &TupleType{Elements: []Type{StringType, &NullableType{Inner: global.Lookup("Error").Type}}},
			},
			"write_text": {
				Name:   "write_text",
				Params: []*ParamInfo{{Name: "path", Type: StringType}, {Name: "value", Type: StringType}},
				ReturnType: &NullableType{Inner: global.Lookup("Error").Type},
			},
			"exists": {
				Name:       "exists",
				Params:     []*ParamInfo{{Name: "path", Type: StringType}},
				ReturnType: BoolType,
			},
		},
	}
	global.Define(&Symbol{Name: "file", Type: fileNs, Kind: SymNamespace, Pub: true})

	// window
	windowNs := &NamespaceType{
		Name: "window",
		Members: map[string]*FuncSignature{
			"open":  {Name: "open", Params: []*ParamInfo{{Name: "ref", Type: AnyType, Ref: true}}},
			"close": {Name: "close", Params: []*ParamInfo{{Name: "ref", Type: AnyType, Ref: true}}},
		},
	}
	global.Define(&Symbol{Name: "window", Type: windowNs, Kind: SymNamespace, Pub: true})

	// msg
	msgNs := &NamespaceType{
		Name: "msg",
		Members: map[string]*FuncSignature{
			"alert": {Name: "alert", Params: []*ParamInfo{{Name: "message", Type: StringType}}},
			"error": {Name: "error", Params: []*ParamInfo{{Name: "message", Type: StringType}}},
			"info":  {Name: "info", Params: []*ParamInfo{{Name: "message", Type: StringType}}},
		},
	}
	global.Define(&Symbol{Name: "msg", Type: msgNs, Kind: SymNamespace, Pub: true})

	// json
	jsonNs := &NamespaceType{
		Name: "json",
		Members: map[string]*FuncSignature{
			"parse": {
				Name:       "parse",
				Params:     []*ParamInfo{{Name: "value", Type: StringType}},
				ReturnType: &TupleType{Elements: []Type{AnyType, &NullableType{Inner: global.Lookup("Error").Type}}},
			},
			"stringify": {
				Name:       "stringify",
				Params:     []*ParamInfo{{Name: "value", Type: AnyType}},
				ReturnType: &TupleType{Elements: []Type{StringType, &NullableType{Inner: global.Lookup("Error").Type}}},
			},
		},
	}
	global.Define(&Symbol{Name: "json", Type: jsonNs, Kind: SymNamespace, Pub: true})

	// modal
	modalNs := &NamespaceType{
		Name: "modal",
		Members: map[string]*FuncSignature{
			"open":  {Name: "open", Params: []*ParamInfo{{Name: "ref", Type: AnyType, Ref: true}}},
			"close": {Name: "close", Params: []*ParamInfo{{Name: "ref", Type: AnyType, Ref: true}}},
		},
	}
	global.Define(&Symbol{Name: "modal", Type: modalNs, Kind: SymNamespace, Pub: true})

	// toast
	toastNs := &NamespaceType{
		Name: "toast",
		Members: map[string]*FuncSignature{
			"show": {
				Name:   "show",
				Params: []*ParamInfo{{Name: "message", Type: StringType}, {Name: "duration", Type: IntType}, {Name: "type", Type: AnyType}},
			},
		},
	}
	global.Define(&Symbol{Name: "toast", Type: toastNs, Kind: SymNamespace, Pub: true})

	// system
	systemNs := &NamespaceType{
		Name: "system",
		Members: map[string]*FuncSignature{
			"exit": {
				Name:   "exit",
				Params: []*ParamInfo{{Name: "code", Type: IntType}},
			},
			"args": {
				Name:       "args",
				ReturnType: &GenericType{Name: "List", Params: []Type{StringType}},
			},
			"env": {
				Name:       "env",
				Params:     []*ParamInfo{{Name: "key", Type: StringType}},
				ReturnType: &NullableType{Inner: StringType},
			},
		},
	}
	global.Define(&Symbol{Name: "system", Type: systemNs, Kind: SymNamespace, Pub: true})

	// clipboard
	clipNs := &NamespaceType{
		Name: "clipboard",
		Members: map[string]*FuncSignature{
			"copy": {
				Name:   "copy",
				Params: []*ParamInfo{{Name: "text", Type: StringType}},
			},
			"paste": {
				Name:       "paste",
				ReturnType: StringType,
			},
			"has_text": {
				Name:       "has_text",
				ReturnType: BoolType,
			},
		},
	}
	global.Define(&Symbol{Name: "clipboard", Type: clipNs, Kind: SymNamespace, Pub: true})

	// screen
	screenNs := &NamespaceType{
		Name: "screen",
		Members: map[string]*FuncSignature{
			"get_width": {
				Name:       "get_width",
				ReturnType: IntType,
			},
			"get_height": {
				Name:       "get_height",
				ReturnType: IntType,
			},
		},
	}
	global.Define(&Symbol{Name: "screen", Type: screenNs, Kind: SymNamespace, Pub: true})
}

func registerConstGroups(global *Scope) {
	colorType := global.Lookup("Color")
	dockType := global.Lookup("Dock")
	orientType := global.Lookup("Orientation")
	toastType := global.Lookup("ToastType")
	fontType := global.Lookup("FontWeight")
	alignType := global.Lookup("TextAlign")
	scrollType := global.Lookup("ScrollMode")
	sortType := global.Lookup("SortOrder")

	groups := []struct {
		name   string
		typ    Type
		consts map[string]Type
	}{
		{"colors", colorType.Type, map[string]Type{
			"blue": colorType.Type, "red": colorType.Type, "white": colorType.Type,
			"black": colorType.Type, "purple": colorType.Type, "green": colorType.Type,
			"yellow": colorType.Type, "orange": colorType.Type, "gray": colorType.Type,
		}},
		{"dock", dockType.Type, map[string]Type{
			"left": dockType.Type, "right": dockType.Type, "center": dockType.Type,
			"top": dockType.Type, "bottom": dockType.Type, "fill": dockType.Type,
		}},
		{"orientation", orientType.Type, map[string]Type{
			"horizontal": orientType.Type, "vertical": orientType.Type,
		}},
		{"toast_type", toastType.Type, map[string]Type{
			"error": toastType.Type, "success": toastType.Type,
		}},
		{"font_weight", fontType.Type, map[string]Type{
			"normal": fontType.Type, "bold": fontType.Type, "light": fontType.Type,
		}},
		{"text_align", alignType.Type, map[string]Type{
			"left": alignType.Type, "right": alignType.Type, "center": alignType.Type,
		}},
		{"scroll_mode", scrollType.Type, map[string]Type{
			"vertical": scrollType.Type, "horizontal": scrollType.Type, "both": scrollType.Type,
		}},
		{"sort_order", sortType.Type, map[string]Type{
			"ascending": sortType.Type, "descending": sortType.Type,
		}},
	}
	for _, g := range groups {
		ns := &NamespaceType{Name: g.name, Consts: g.consts}
		global.Define(&Symbol{Name: g.name, Type: ns, Kind: SymConstGroup, Pub: true})
	}
}

func registerComponentTypes(global *Scope) {
	// Common property types
	dockType := global.Lookup("Dock").Type
	colorType := global.Lookup("Color").Type
	orientType := global.Lookup("Orientation").Type
	scrollType := global.Lookup("ScrollMode").Type
	fontType := global.Lookup("FontWeight").Type
	alignType := global.Lookup("TextAlign").Type
	sortType := global.Lookup("SortOrder").Type

	clickEvent := global.Lookup("ClickEvent").Type
	hoverEvent := global.Lookup("HoverEvent").Type
	leaveEvent := global.Lookup("LeaveEvent").Type
	scrollEvent := global.Lookup("ScrollEvent").Type
	resizeEvent := global.Lookup("ResizeEvent").Type
	focusEvent := global.Lookup("FocusEvent").Type
	changeEvent := global.Lookup("ChangeEvent").Type

	colorOrString := colorType // In practice, both Color constants and hex strings are accepted

	components := []struct {
		name       string
		primary    *ParamInfo
		properties map[string]*ParamInfo
		events     map[string]Type
	}{
		// Window
		{"Window",
			&ParamInfo{Name: "title", Type: StringType},
			map[string]*ParamInfo{
				"w": {Name: "w", Type: &NullableType{Inner: IntType}},
				"h": {Name: "h", Type: &NullableType{Inner: IntType}},
				"bg": {Name: "bg", Type: colorOrString},
			},
			nil,
		},
		// Button
		{"Button",
			&ParamInfo{Name: "text", Type: StringType},
			map[string]*ParamInfo{
				"enabled": {Name: "enabled", Type: BoolType},
				"tooltip": {Name: "tooltip", Type: &NullableType{Inner: StringType}},
				"dock":    {Name: "dock", Type: &NullableType{Inner: dockType}},
				"w":       {Name: "w", Type: &NullableType{Inner: IntType}},
				"h":       {Name: "h", Type: &NullableType{Inner: IntType}},
				"bg":      {Name: "bg", Type: colorOrString},
			},
			map[string]Type{"click": clickEvent, "hover": hoverEvent, "leave": leaveEvent},
		},
		// Label
		{"Label",
			&ParamInfo{Name: "text", Type: StringType},
			map[string]*ParamInfo{
				"dock":    {Name: "dock", Type: &NullableType{Inner: dockType}},
				"w":       {Name: "w", Type: &NullableType{Inner: IntType}},
				"h":       {Name: "h", Type: &NullableType{Inner: IntType}},
				"color":   {Name: "color", Type: colorOrString},
				"tooltip": {Name: "tooltip", Type: &NullableType{Inner: StringType}},
				"font_weight": {Name: "font_weight", Type: &NullableType{Inner: fontType}},
				"text_align":  {Name: "text_align", Type: &NullableType{Inner: alignType}},
			},
			nil,
		},
		// Panel
		{"Panel",
			nil,
			map[string]*ParamInfo{
				"bg":    {Name: "bg", Type: colorOrString},
				"dock":  {Name: "dock", Type: &NullableType{Inner: dockType}},
				"w":     {Name: "w", Type: &NullableType{Inner: IntType}},
				"h":     {Name: "h", Type: &NullableType{Inner: IntType}},
				"label": {Name: "label", Type: &NullableType{Inner: StringType}},
			},
			nil,
		},
		// TextInput
		{"TextInput",
			nil,
			map[string]*ParamInfo{
				"value":       {Name: "value", Type: StringType},
				"placeholder": {Name: "placeholder", Type: &NullableType{Inner: StringType}},
				"enabled":     {Name: "enabled", Type: BoolType},
				"w":           {Name: "w", Type: &NullableType{Inner: IntType}},
				"h":           {Name: "h", Type: &NullableType{Inner: IntType}},
			},
			map[string]Type{"change": changeEvent, "focus": focusEvent},
		},
		// TextArea
		{"TextArea",
			nil,
			map[string]*ParamInfo{
				"value":       {Name: "value", Type: StringType},
				"placeholder": {Name: "placeholder", Type: &NullableType{Inner: StringType}},
				"enabled":     {Name: "enabled", Type: BoolType},
				"w":           {Name: "w", Type: &NullableType{Inner: IntType}},
				"h":           {Name: "h", Type: &NullableType{Inner: IntType}},
			},
			map[string]Type{"change": changeEvent, "focus": focusEvent},
		},
		// Checkbox
		{"Checkbox",
			&ParamInfo{Name: "text", Type: StringType},
			map[string]*ParamInfo{
				"checked": {Name: "checked", Type: BoolType},
				"enabled": {Name: "enabled", Type: BoolType},
			},
			map[string]Type{"change": changeEvent},
		},
		// RadioButton
		{"RadioButton",
			&ParamInfo{Name: "text", Type: StringType},
			map[string]*ParamInfo{
				"selected": {Name: "selected", Type: BoolType},
				"enabled":  {Name: "enabled", Type: BoolType},
			},
			map[string]Type{"change": changeEvent},
		},
		// Toggle
		{"Toggle",
			nil,
			map[string]*ParamInfo{
				"checked": {Name: "checked", Type: BoolType},
				"enabled": {Name: "enabled", Type: BoolType},
			},
			map[string]Type{"change": changeEvent},
		},
		// Slider
		{"Slider",
			nil,
			map[string]*ParamInfo{
				"value": {Name: "value", Type: FloatType},
				"min":   {Name: "min", Type: FloatType},
				"max":   {Name: "max", Type: FloatType},
				"step":  {Name: "step", Type: FloatType},
			},
			map[string]Type{"change": changeEvent},
		},
		// Dropdown
		{"Dropdown",
			nil,
			map[string]*ParamInfo{
				"options": {Name: "options", Type: &GenericType{Name: "List", Params: []Type{StringType}}},
				"selected": {Name: "selected", Type: IntType},
			},
			map[string]Type{"change": changeEvent},
		},
		// NumberInput
		{"NumberInput",
			nil,
			map[string]*ParamInfo{
				"value": {Name: "value", Type: FloatType},
				"min":   {Name: "min", Type: &NullableType{Inner: FloatType}},
				"max":   {Name: "max", Type: &NullableType{Inner: FloatType}},
				"step":  {Name: "step", Type: FloatType},
			},
			map[string]Type{"change": changeEvent},
		},
		// Image
		{"Image",
			&ParamInfo{Name: "src", Type: StringType},
			map[string]*ParamInfo{
				"w":     {Name: "w", Type: &NullableType{Inner: IntType}},
				"h":     {Name: "h", Type: &NullableType{Inner: IntType}},
				"dock":  {Name: "dock", Type: &NullableType{Inner: dockType}},
			},
			nil,
		},
		// ProgressBar
		{"ProgressBar",
			nil,
			map[string]*ParamInfo{
				"value": {Name: "value", Type: FloatType},
				"min":   {Name: "min", Type: FloatType},
				"max":   {Name: "max", Type: FloatType},
				"w":     {Name: "w", Type: &NullableType{Inner: IntType}},
				"h":     {Name: "h", Type: &NullableType{Inner: IntType}},
			},
			nil,
		},
		// Row
		{"Row",
			nil,
			map[string]*ParamInfo{
				"gap":  {Name: "gap", Type: IntType},
				"dock": {Name: "dock", Type: &NullableType{Inner: dockType}},
				"w":    {Name: "w", Type: &NullableType{Inner: IntType}},
				"h":    {Name: "h", Type: &NullableType{Inner: IntType}},
			},
			nil,
		},
		// Col
		{"Col",
			nil,
			map[string]*ParamInfo{
				"gap":  {Name: "gap", Type: IntType},
				"dock": {Name: "dock", Type: &NullableType{Inner: dockType}},
				"w":    {Name: "w", Type: &NullableType{Inner: IntType}},
				"h":    {Name: "h", Type: &NullableType{Inner: IntType}},
			},
			nil,
		},
		// Grid
		{"Grid",
			nil,
			map[string]*ParamInfo{
				"columns": {Name: "columns", Type: IntType},
				"gap":     {Name: "gap", Type: IntType},
				"dock":    {Name: "dock", Type: &NullableType{Inner: dockType}},
			},
			nil,
		},
		// ScrollPanel
		{"ScrollPanel",
			nil,
			map[string]*ParamInfo{
				"scroll_mode": {Name: "scroll_mode", Type: &NullableType{Inner: scrollType}},
				"dock":        {Name: "dock", Type: &NullableType{Inner: dockType}},
				"w":           {Name: "w", Type: &NullableType{Inner: IntType}},
				"h":           {Name: "h", Type: &NullableType{Inner: IntType}},
			},
			map[string]Type{"scroll": scrollEvent},
		},
		// ResizablePanel
		{"ResizablePanel",
			nil,
			map[string]*ParamInfo{
				"dock": {Name: "dock", Type: &NullableType{Inner: dockType}},
				"w":    {Name: "w", Type: &NullableType{Inner: IntType}},
				"h":    {Name: "h", Type: &NullableType{Inner: IntType}},
				"min_w": {Name: "min_w", Type: IntType},
				"min_h": {Name: "min_h", Type: IntType},
			},
			map[string]Type{"resize": resizeEvent},
		},
		// StackLayout
		{"StackLayout",
			nil,
			map[string]*ParamInfo{
				"orientation": {Name: "orientation", Type: &NullableType{Inner: orientType}},
				"dock":        {Name: "dock", Type: &NullableType{Inner: dockType}},
			},
			nil,
		},
		// Line
		{"Line",
			nil,
			map[string]*ParamInfo{
				"dock":   {Name: "dock", Type: &NullableType{Inner: dockType}},
				"color":  {Name: "color", Type: colorOrString},
				"weight": {Name: "weight", Type: IntType},
			},
			nil,
		},
		// Tabs
		{"Tabs",
			nil,
			map[string]*ParamInfo{
				"dock": {Name: "dock", Type: &NullableType{Inner: dockType}},
				"w":    {Name: "w", Type: &NullableType{Inner: IntType}},
				"h":    {Name: "h", Type: &NullableType{Inner: IntType}},
			},
			map[string]Type{"change": changeEvent},
		},
		// Tab
		{"Tab",
			&ParamInfo{Name: "title", Type: StringType},
			nil,
			nil,
		},
		// MenuBar
		{"MenuBar",
			nil,
			map[string]*ParamInfo{
				"dock": {Name: "dock", Type: &NullableType{Inner: dockType}},
			},
			nil,
		},
		// Menu
		{"Menu",
			&ParamInfo{Name: "title", Type: StringType},
			nil,
			nil,
		},
		// MenuItem
		{"MenuItem",
			&ParamInfo{Name: "text", Type: StringType},
			map[string]*ParamInfo{
				"shortcut": {Name: "shortcut", Type: &NullableType{Inner: StringType}},
				"enabled":  {Name: "enabled", Type: BoolType},
			},
			map[string]Type{"click": clickEvent},
		},
		// MenuSeparator
		{"MenuSeparator",
			nil,
			nil,
			nil,
		},
		// ContextMenu
		{"ContextMenu",
			nil,
			nil,
			nil,
		},
		// Toolbar
		{"Toolbar",
			nil,
			map[string]*ParamInfo{
				"dock": {Name: "dock", Type: &NullableType{Inner: dockType}},
			},
			nil,
		},
		// Modal
		{"Modal",
			nil,
			map[string]*ParamInfo{
				"title": {Name: "title", Type: StringType},
				"w":     {Name: "w", Type: &NullableType{Inner: IntType}},
				"h":     {Name: "h", Type: &NullableType{Inner: IntType}},
			},
			nil,
		},
		// Table
		{"Table",
			nil,
			map[string]*ParamInfo{
				"columns": {Name: "columns", Type: &GenericType{Name: "List", Params: []Type{StringType}}},
				"sortable": {Name: "sortable", Type: BoolType},
				"sort_order": {Name: "sort_order", Type: &NullableType{Inner: sortType}},
				"dock":   {Name: "dock", Type: &NullableType{Inner: dockType}},
				"w":      {Name: "w", Type: &NullableType{Inner: IntType}},
				"h":      {Name: "h", Type: &NullableType{Inner: IntType}},
			},
			nil,
		},
		// TreeView
		{"TreeView",
			nil,
			map[string]*ParamInfo{
				"dock": {Name: "dock", Type: &NullableType{Inner: dockType}},
				"w":    {Name: "w", Type: &NullableType{Inner: IntType}},
				"h":    {Name: "h", Type: &NullableType{Inner: IntType}},
			},
			nil,
		},
		// Splitter
		{"Splitter",
			nil,
			map[string]*ParamInfo{
				"orientation": {Name: "orientation", Type: &NullableType{Inner: orientType}},
				"dock":        {Name: "dock", Type: &NullableType{Inner: dockType}},
			},
			nil,
		},
		// ListView
		{"ListView",
			nil,
			map[string]*ParamInfo{
				"dock":       {Name: "dock", Type: &NullableType{Inner: dockType}},
				"w":          {Name: "w", Type: &NullableType{Inner: IntType}},
				"h":          {Name: "h", Type: &NullableType{Inner: IntType}},
				"item_height": {Name: "item_height", Type: IntType},
			},
			nil,
		},
		// DatePicker
		{"DatePicker",
			nil,
			map[string]*ParamInfo{
				"value": {Name: "value", Type: &NullableType{Inner: StringType}},
			},
			map[string]Type{"change": changeEvent},
		},
		// ColorPicker
		{"ColorPicker",
			nil,
			map[string]*ParamInfo{
				"value": {Name: "value", Type: StringType},
			},
			map[string]Type{"change": changeEvent},
		},
		// Notify
		{"Notify",
			nil,
			map[string]*ParamInfo{
				"position": {Name: "position", Type: &NullableType{Inner: dockType}},
			},
			nil,
		},
	}

	for _, comp := range components {
		ct := &ComponentType{
			Name:       comp.name,
			Primary:    comp.primary,
			Properties: comp.properties,
			Events:     comp.events,
		}
		if ct.Properties == nil {
			ct.Properties = make(map[string]*ParamInfo)
		}
		if ct.Events == nil {
			ct.Events = make(map[string]Type)
		}
		global.Define(&Symbol{Name: comp.name, Type: ct, Kind: SymType, Pub: true})
	}
}

// EnumType represents a built-in enum-like type (Color, Dock, etc.).
type EnumType struct {
	Name    string
	Members []string
}

func (t *EnumType) typeKey() string { return t.Name }
func (t *EnumType) String() string  { return t.Name }
