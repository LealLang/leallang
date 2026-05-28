# Components Overview

LealLang UI is declared as an indented tree of windows and components.

## Syntax Forms

Component declarations create UI nodes.

```python
Button[save_button]("Save")
```

UI references access existing UI nodes.

```python
@Button[save_button].text = "Saved"
```

Window-qualified UI references access components through a specific window.

```python
@Window[settings].Button[close_button].text = "Close"
```

Collection indexing uses the same `[]` characters on values.

```python
numbers[0]
scores["Ana"]
```

Generic types use angle brackets.

```python
List<string>
Dict<string, int>
```

Component IDs are not normal variables. They are identifiers inside the UI tree and are accessed through `@` references.

---

## Windows

Windows are declared declaratively.

```python
Window[main_window]("MyApp", w: 900, h: 600):
    Label[title_label]("Hello")
```

Window IDs are unique within a package.

A window can be public.

```python
pub Window[settings]("Settings", w: 600, h: 400):
    Button[close_button]("Close", click: close_clicked)
```

Opening and closing windows is done through the `window` namespace.

```python
window.open(@Window[settings])
window.close(@Window[settings])
```

Windows are usually exposed through public functions.

```python
pub func open():
    window.open(@Window[settings])
```

---

## Components

Components are declared inside windows or other components.

```python
Window[main_window]("MyApp"):
    Panel[side_panel](bg: colors.white, dock: dock.left, w: 300):
        Button[save_button]("Save", click: save_clicked)
```

Component IDs are unique per window.

This is valid because the IDs belong to different windows:

```python
Window[main_window]("Main"):
    Button[save_button]("Save")

Window[settings]("Settings"):
    Button[save_button]("Save")
```

## Component Arguments

A component can use one positional argument only for its obvious primary value.

```python
Button[save_button]("Save")
Label[title_label]("Settings")
Window[main_window]("MyApp", w: 900, h: 600)
```

Style, layout, sizing, behavior, and event arguments must be named.

```python
Panel[side_panel](bg: colors.white, dock: dock.left, w: 240)
Button[save_button]("Save", click: save_clicked)
Grid[button_grid](cols: 3, rows: 2, gap: 8)
```

Unclear positional component arguments are invalid.

```python
Panel[side_panel]("#ffffffff", dock.left, w: 400) # invalid
Panel[side_panel](bg: "#ffffffff", dock: dock.left, w: 400) # valid
```

---

## UI References with @

The `@` operator is used only to access UI windows and components.

Declarations do not use `@`.

```python
Window[settings]("Settings"):
    Button[close_button]("Close")
```

Access uses `@`.

```python
settings_window: Window = @Window[settings]
```

Each built-in component name is also the type of references to that component. For example, `@Window[settings]` has type `Window`, and `@Button[save_button]` has type `Button`.

```python
settings_window: Window = @Window[settings]
save_button: Button = @Button[save_button]
```

LealLang has two official UI reference forms.

Local UI reference:

```python
@ComponentType[component_id]
```

Window-qualified UI reference:

```python
@Window[window_id].ComponentType[component_id]
```

Local UI references are valid inside a function that is being used as an event handler for a specific window context.

```python
Window[settings]("Settings"):
    Button[close_button]("Close", click: close_clicked)

func close_clicked():
    @Button[close_button].text = "Closing"
```

Outside such a context, use window-qualified references.

```python
func load_settings():
    @Window[settings].TextInput[theme_input].value = "dark"
```

UI references can be used as argument values.

```python
Button[target_button]("Right-click me", context_menu: @ContextMenu[action_menu])
```

### Invalid UI References

Using a missing component ID is invalid.

```python
@Button[missing_button].text = "Save" # invalid
```

Using the wrong component type for an ID is invalid.

```python
Window[settings]("Settings"):
    TextInput[theme_input](placeholder: "Theme")

@Button[theme_input].text = "Theme" # invalid
```

---

## Event Handlers

Events are assigned by passing function references to event properties.

```python
Button[save_button]("Save", click: save_clicked)
```

An event handler can omit the event parameter.

```python
func save_clicked():
    @Button[save_button].text = "Saved"
```

An event handler can include one event parameter.

```python
func save_clicked(e: ClickEvent):
    console.print_ln($"Clicked at ({e.mouse_x}, {e.mouse_y})")
```

The event parameter type must match the event's declared event type.

```python
Button[save_button]("Save", click: save_clicked)

func save_clicked(e: ResizeEvent): # invalid for click
    console.print_ln(e.width)
```

Event handlers return no value unless a specific component event explicitly documents otherwise.

---

## Component Type Checking

Component properties are checked by name and type.

```python
Button[save_button]("Save", enabled: true, click: save_clicked)
```

A component property must exist.

```python
Button[save_button]("Save", unknown: true) # invalid
```

A property value must match its declared type.

```python
Button[save_button]("Save", enabled: "yes") # invalid
```

Event handler signatures are checked against event types.

```python
Button[save_button]("Save", click: save_clicked)

func save_clicked(e: ClickEvent):
    console.print_ln(e.button)
```
