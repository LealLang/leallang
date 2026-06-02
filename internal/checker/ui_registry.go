// Copyright (c) 2026 LealLang Contributors
// SPDX-License-Identifier: MIT

package checker

type uiComponentSpec struct {
	name       string
	primary    *ParamInfo
	properties map[string]*ParamInfo
	events     map[string]Type
}

type uiRegistryTypes struct {
	dock        Type
	color       Type
	orientation Type
	scrollMode  Type
	fontWeight  Type
	textAlign   Type
	sortOrder   Type
	toastType   Type
	click       Type
	hover       Type
	leave       Type
	scroll      Type
	resize      Type
	focus       Type
	change      Type
}

func registerUIComponentTypes(global *Scope) {
	types := uiRegistryTypes{
		dock:        mustLookupType(global, "Dock"),
		color:       mustLookupType(global, "Color"),
		orientation: mustLookupType(global, "Orientation"),
		scrollMode:  mustLookupType(global, "ScrollMode"),
		fontWeight:  mustLookupType(global, "FontWeight"),
		textAlign:   mustLookupType(global, "TextAlign"),
		sortOrder:   mustLookupType(global, "SortOrder"),
		toastType:   mustLookupType(global, "ToastType"),
		click:       mustLookupType(global, "ClickEvent"),
		hover:       mustLookupType(global, "HoverEvent"),
		leave:       mustLookupType(global, "LeaveEvent"),
		scroll:      mustLookupType(global, "ScrollEvent"),
		resize:      mustLookupType(global, "ResizeEvent"),
		focus:       mustLookupType(global, "FocusEvent"),
		change:      mustLookupType(global, "ChangeEvent"),
	}

	for _, spec := range uiComponentSpecs(types) {
		ct := &ComponentType{
			Name:       spec.name,
			Primary:    spec.primary,
			Properties: spec.properties,
			Events:     spec.events,
		}
		if ct.Properties == nil {
			ct.Properties = make(map[string]*ParamInfo)
		}
		if ct.Events == nil {
			ct.Events = make(map[string]Type)
		}
		global.Define(&Symbol{Name: spec.name, Type: ct, Kind: SymType, Pub: true})
	}
}

func uiComponentSpecs(t uiRegistryTypes) []uiComponentSpec {
	return []uiComponentSpec{
		{
			name:    "Window",
			primary: prop("title", StringType),
			properties: props(
				prop("w", nullable(IntType)),
				prop("h", nullable(IntType)),
				prop("bg", t.color),
			),
			events: events(event("resize", t.resize)),
		},
		{
			name:    "Button",
			primary: prop("text", StringType),
			properties: props(
				prop("enabled", BoolType),
				prop("tooltip", nullable(StringType)),
				prop("dock", nullable(t.dock)),
				prop("w", nullable(IntType)),
				prop("h", nullable(IntType)),
				prop("bg", t.color),
			),
			events: events(event("click", t.click), event("hover", t.hover), event("leave", t.leave)),
		},
		{
			name:    "Label",
			primary: prop("text", StringType),
			properties: props(
				prop("dock", nullable(t.dock)),
				prop("w", nullable(IntType)),
				prop("h", nullable(IntType)),
				prop("color", t.color),
				prop("tooltip", nullable(StringType)),
				prop("font_weight", nullable(t.fontWeight)),
				prop("text_align", nullable(t.textAlign)),
			),
		},
		{
			name: "Panel",
			properties: props(
				prop("bg", t.color),
				prop("dock", nullable(t.dock)),
				prop("w", nullable(IntType)),
				prop("h", nullable(IntType)),
				prop("label", nullable(StringType)),
			),
		},
		{
			name: "TextInput",
			properties: props(
				prop("value", StringType),
				prop("placeholder", nullable(StringType)),
				prop("enabled", BoolType),
				prop("w", nullable(IntType)),
				prop("h", nullable(IntType)),
			),
			events: events(event("change", t.change), event("focus", t.focus)),
		},
		{
			name: "TextArea",
			properties: props(
				prop("value", StringType),
				prop("placeholder", nullable(StringType)),
				prop("enabled", BoolType),
				prop("w", nullable(IntType)),
				prop("h", nullable(IntType)),
			),
			events: events(event("change", t.change), event("focus", t.focus)),
		},
		{
			name:    "Checkbox",
			primary: prop("label", StringType),
			properties: props(
				prop("checked", BoolType),
				prop("enabled", BoolType),
			),
			events: events(event("change", t.change)),
		},
		{
			name:    "RadioButton",
			primary: prop("label", StringType),
			properties: props(
				prop("selected", BoolType),
				prop("enabled", BoolType),
				prop("options", dictOf(StringType, StringType)),
				prop("selected_text", nullable(StringType)),
				prop("selected_value", nullable(StringType)),
			),
			events: events(event("change", t.change)),
		},
		{
			name:    "Toggle",
			primary: prop("label", StringType),
			properties: props(
				prop("on", BoolType),
				prop("checked", BoolType),
				prop("enabled", BoolType),
			),
			events: events(event("change", t.change)),
		},
		{
			name: "Slider",
			properties: props(
				prop("value", FloatType),
				prop("min", FloatType),
				prop("max", FloatType),
				prop("step", FloatType),
			),
			events: events(event("change", t.change)),
		},
		{
			name: "Dropdown",
			properties: props(
				prop("options", listOf(StringType)),
				prop("selected", IntType),
			),
			events: events(event("change", t.change)),
		},
		{
			name: "NumberInput",
			properties: props(
				prop("value", FloatType),
				prop("min", nullable(FloatType)),
				prop("max", nullable(FloatType)),
				prop("step", FloatType),
			),
			events: events(event("change", t.change)),
		},
		{
			name:    "Image",
			primary: prop("src", StringType),
			properties: props(
				prop("w", nullable(IntType)),
				prop("h", nullable(IntType)),
				prop("dock", nullable(t.dock)),
				prop("tooltip", nullable(StringType)),
			),
		},
		{
			name: "ProgressBar",
			properties: props(
				prop("value", FloatType),
				prop("min", FloatType),
				prop("max", FloatType),
				prop("w", nullable(IntType)),
				prop("h", nullable(IntType)),
			),
		},
		containerSpec("Row", t),
		containerSpec("Col", t),
		{
			name: "Grid",
			properties: props(
				prop("cols", IntType),
				prop("rows", nullable(IntType)),
				prop("gap", IntType),
				prop("dock", nullable(t.dock)),
				prop("w", nullable(IntType)),
				prop("h", nullable(IntType)),
			),
		},
		{
			name: "ScrollPanel",
			properties: props(
				prop("scroll_mode", nullable(t.scrollMode)),
				prop("dock", nullable(t.dock)),
				prop("w", nullable(IntType)),
				prop("h", nullable(IntType)),
			),
			events: events(event("scroll", t.scroll)),
		},
		{
			name: "ResizablePanel",
			properties: props(
				prop("dock", nullable(t.dock)),
				prop("w", nullable(IntType)),
				prop("h", nullable(IntType)),
				prop("min_w", nullable(IntType)),
				prop("max_w", nullable(IntType)),
				prop("min_h", nullable(IntType)),
				prop("max_h", nullable(IntType)),
			),
			events: events(event("resize", t.resize)),
		},
		{
			name: "StackLayout",
			properties: props(
				prop("orientation", nullable(t.orientation)),
				prop("dock", nullable(t.dock)),
				prop("w", nullable(IntType)),
				prop("h", nullable(IntType)),
			),
		},
		{
			name: "Line",
			properties: props(
				prop("orientation", t.orientation),
				prop("thickness", IntType),
				prop("size", nullable(IntType)),
				prop("dock", nullable(t.dock)),
				prop("color", t.color),
			),
		},
		{
			name: "Tabs",
			properties: props(
				prop("dock", nullable(t.dock)),
				prop("w", nullable(IntType)),
				prop("h", nullable(IntType)),
			),
			events: events(event("change", t.change)),
		},
		{name: "Tab", primary: prop("title", StringType)},
		{name: "MenuBar", properties: props(prop("dock", nullable(t.dock)))},
		{name: "Menu", primary: prop("label", StringType)},
		{
			name:    "MenuItem",
			primary: prop("text", StringType),
			properties: props(
				prop("shortcut", nullable(StringType)),
				prop("enabled", BoolType),
			),
			events: events(event("click", t.click)),
		},
		{name: "MenuSeparator"},
		{name: "ContextMenu"},
		{name: "Toolbar", properties: props(prop("dock", nullable(t.dock)))},
		{
			name: "Modal",
			properties: props(
				prop("title", StringType),
				prop("w", nullable(IntType)),
				prop("h", nullable(IntType)),
			),
		},
		{
			name: "Table",
			properties: props(
				prop("data", listOf(dictOf(StringType, StringType))),
				prop("columns", nullable(listOf(StringType))),
				prop("sortable", BoolType),
				prop("sort_order", nullable(t.sortOrder)),
				prop("dock", nullable(t.dock)),
				prop("w", nullable(IntType)),
				prop("h", nullable(IntType)),
			),
			events: events(event("click", t.click), event("change", t.change)),
		},
		{
			name: "TreeView",
			properties: props(
				prop("dock", nullable(t.dock)),
				prop("w", nullable(IntType)),
				prop("h", nullable(IntType)),
			),
		},
		{
			name: "Splitter",
			properties: props(
				prop("orientation", nullable(t.orientation)),
				prop("dock", nullable(t.dock)),
			),
		},
		{
			name: "ListView",
			properties: props(
				prop("items", listOf(StringType)),
				prop("dock", nullable(t.dock)),
				prop("w", nullable(IntType)),
				prop("h", nullable(IntType)),
				prop("item_height", IntType),
			),
		},
		{
			name:       "DatePicker",
			properties: props(prop("value", nullable(StringType))),
			events:     events(event("change", t.change)),
		},
		{
			name:       "ColorPicker",
			properties: props(prop("value", t.color)),
			events:     events(event("change", t.change)),
		},
		{
			name: "Notify",
			properties: props(
				prop("position", nullable(t.dock)),
				prop("message", StringType),
			),
		},
	}
}

func containerSpec(name string, t uiRegistryTypes) uiComponentSpec {
	return uiComponentSpec{
		name: name,
		properties: props(
			prop("gap", IntType),
			prop("dock", nullable(t.dock)),
			prop("w", nullable(IntType)),
			prop("h", nullable(IntType)),
		),
	}
}

func mustLookupType(scope *Scope, name string) Type {
	sym := scope.Lookup(name)
	if sym == nil {
		return AnyType
	}
	return sym.Type
}

func nullable(t Type) Type {
	return &NullableType{Inner: t}
}

func listOf(elem Type) Type {
	return &GenericType{Name: "List", Params: []Type{elem}}
}

func dictOf(key, value Type) Type {
	return &GenericType{Name: "Dict", Params: []Type{key, value}}
}

func prop(name string, typ Type) *ParamInfo {
	return &ParamInfo{Name: name, Type: typ}
}

func props(items ...*ParamInfo) map[string]*ParamInfo {
	out := make(map[string]*ParamInfo, len(items))
	for _, item := range items {
		out[item.Name] = item
	}
	return out
}

type eventSpec struct {
	name string
	typ  Type
}

func event(name string, typ Type) eventSpec {
	return eventSpec{name: name, typ: typ}
}

func events(items ...eventSpec) map[string]Type {
	out := make(map[string]Type, len(items))
	for _, item := range items {
		out[item.name] = item.typ
	}
	return out
}
