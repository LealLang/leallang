# Components Overview

LealLang UI is declared as an indented tree of windows and components.

## Syntax Forms

Component declarations create UI nodes using a block syntax with indented properties.

```python
Button[save_button]:
    text = "Save"
```

UI references access existing UI nodes.

```python
@Button[save_button].text = "Saved"
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

Windows are declared declaratively using a block with indented properties.

```python
Window[main_window]:
    title = "MyApp"
    w = 900
    h = 600

    Label[title_label]:
        text = "Hello"
```

Window IDs are unique within a package.

A window can be public.

```python
pub Window[settings]:
    title = "Settings"
    w = 600
    h = 400

    Button[close_button]:
        text = "Close"
        on click = close_clicked
```

Opening and closing windows is done through the `window` namespace.

```python
window.open(@Window[settings])
window.close(@Window[settings])
```

Opening a window starts that window's actor and returns immediately. It does not block the calling window.

Windows are usually exposed through public functions.

```python
pub func open():
    window.open(@Window[settings])
```

---

## Window Actors

Each `Window[...]` owns a UI actor. A UI actor is the serialized event queue for one window. It runs that window's event handlers, local UI updates, and `ui func` calls.

If one window runs slow synchronous code in one of its handlers, only that window's actor is blocked. Other windows keep processing their own events because they have separate actors.

```python
Window[main]:
    title = "Main"

    Button[open_settings]:
        text = "Settings"
        on click = open_settings_clicked

Window[settings]:
    title = "Settings"

    Button[save_button]:
        text = "Save"
        on click = save_settings_clicked
```

If `save_settings_clicked` runs slow code, only the Settings window is blocked. The Main window keeps responding.

---

## Window-owned ui func

A `ui func` is a UI-facing function owned by a specific window. It is declared inside the owning `Window[...]` block.

```python
Window[main]:
    title = "App"

    Label[theme_label]:
        text = "Light"

    pub ui func apply_settings(settings: Settings) -> Error?:
        @Label[theme_label].text = settings.theme
        return null
```

Rules:

- A `ui func` runs on its owning window actor.
- A `ui func` may use local UI refs for its owning window, such as `@Label[theme_label]`.
- Public cross-window UI updates should be exposed through `pub ui func`.
- Other windows should not directly mutate another window's components.

### Cross-window Calls

Cross-window calls use the `@Window[id].ui_func(args)` form:

```python
@Window[main].apply_settings(settings)
```

This queues work onto the target window actor. The call returns a `Task`. Awaiting is optional.

Fire-and-forget:

```python
func save_clicked():
    settings, err = collect_settings()
    if err != null:
        msg.error(err.message)
        return

    @Window[main].apply_settings(settings)
    window.close(@Window[settings])
```

Awaiting completion:

```python
func save_clicked():
    settings, err = collect_settings()
    if err != null:
        msg.error(err.message)
        return

    task = @Window[main].apply_settings(settings)
    err = await task
    if err != null:
        msg.error(err.message)
        return

    window.close(@Window[settings])
```

---

## Cross-window Communication

For the first version, the main window owns shared application state. Secondary windows keep local draft state and send validated changes to Main through `pub ui func`s.

```python
type Settings:
    pub theme: string
    pub autosave: bool

Window[main]:
    title = "App"

    pub ui func apply_settings(settings: Settings) -> Error?:
        app_settings = settings
        @Label[theme_label].text = settings.theme
        return null

Window[settings]:
    title = "Settings"

    Button[save_button]:
        text = "Save"
        on click = save_settings_clicked

func save_settings_clicked():
    settings, err = collect_settings()
    if err != null:
        msg.error(err.message)
        return

    err = await @Window[main].apply_settings(settings)
    if err != null:
        msg.error(err.message)
        return

    window.close(@Window[settings])
```

This avoids global mutable state shared by multiple windows.

---

## Components

Components are declared inside windows or other components using block syntax.

```python
Window[main_window]:
    title = "MyApp"

    Panel[side_panel]:
        bg = colors.white
        dock = dock.left
        w = 300

        Button[save_button]:
            text = "Save"
            on click = save_clicked
```

Component IDs are unique per window.

This is valid because the IDs belong to different windows:

```python
Window[main]:
    title = "Main"

    Button[save_button]:
        text = "Save"

Window[settings]:
    title = "Settings"

    Button[save_button]:
        text = "Save"
```

## Component Properties

Properties are declared as indented `key = value` lines inside the component block.

```python
Button[save_button]:
    text = "Save"
    enabled = true
    on click = save_clicked

Grid[button_grid]:
    cols = 3
    rows = 2
    gap = 8
```

Properties are checked by name and type.

A component property must exist.

```python
Button[save_button]:
    text = "Save"
    unknown = true # invalid
```

A property value must match its declared type.

```python
Button[save_button]:
    text = "Save"
    enabled = "yes" # invalid
```

---

## UI References with @

The `@` operator is used only to access UI windows and components.

Declarations do not use `@`.

```python
Window[settings]:
    title = "Settings"

    Button[close_button]:
        text = "Close"
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

UI references use the `@ComponentType[component_id]` form.

```python
@ComponentType[component_id]
```

UI references are valid inside a function that is being used as an event handler for a specific window context.

```python
Window[settings]:
    title = "Settings"

    Button[close_button]:
        text = "Close"
        on click = close_clicked

func close_clicked():
    @Button[close_button].text = "Closing"
```

UI references can be used as argument values.

```python
Button[target_button]:
    text = "Right-click me"
    context_menu = @ContextMenu[action_menu]
```

### Invalid UI References

Using a missing component ID is invalid.

```python
@Button[missing_button].text = "Save" # invalid
```

Using the wrong component type for an ID is invalid.

```python
Window[settings]:
    title = "Settings"

    TextInput[theme_input]:
        placeholder = "Theme"

@Button[theme_input].text = "Theme" # invalid
```

---

## Event Handlers

Events are declared with the `on event = handler` syntax inside the component block.

```python
Button[save_button]:
    text = "Save"
    on click = save_clicked
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
Button[save_button]:
    text = "Save"
    on click = save_clicked

func save_clicked(e: ResizeEvent): # invalid for click
    console.print_ln(e.width)
```

Event handlers return no value unless a specific component event explicitly documents otherwise.

---

## Component Type Checking

Event handler signatures are checked against event types.

```python
Button[save_button]:
    text = "Save"
    on click = save_clicked

func save_clicked(e: ClickEvent):
    console.print_ln(e.button)
```
