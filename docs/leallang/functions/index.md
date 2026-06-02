# Functions

Functions use the `func` keyword.

```python
func greet(name: string):
    console.print_ln($"Hello, {name}")
```

Public functions use `pub`.

```python
pub func open():
    window.open(@Window[settings])
```

A function without `-> Type` returns no value.

```python
func greet(name: string):
    console.print_ln($"Hello, {name}")
```

A function that returns a value must declare its return type.

```python
func add(a: int, b: int) -> int:
    return a + b
```

Returning a value without a declared return type is invalid.

```python
func add(a: int, b: int):
    return a + b # invalid
```

`return` by itself exits a no-value function.

```python
func load_settings():
    config_value, err = config.load("./settings.json")
    if err != null:
        msg.error(err.message)
        return

    @Window[settings].apply_config(config_value)
```

## Arguments

Positional arguments must come before named arguments.

```python
toast.show("Saved", duration: 3)           # valid
toast.show(message: "Saved", duration: 3)  # valid
toast.show(message: "Saved", 3)            # invalid
```

After the first named argument, all remaining arguments must be named.

Component declarations have stricter positional argument rules. See [Components](../components/index.md).

## ref Parameters

Use `ref` to pass a mutable alias to a variable.

```python
func increment(ref value: int):
    value = value + 1

count = 1
increment(count)
console.print_ln(count) # 2
```

Only variables can be passed to `ref` parameters.

```python
increment(count)     # valid
increment(1)         # invalid
increment(get_num()) # invalid
```

---

## Multiple Return Values

Functions can return multiple values.

The return signature must explicitly declare every returned type.

```python
pub func return_zero() -> int:
    return 0

pub func return_with_error() -> int, Error?:
    return 0, null

pub func return_more_with_error() -> int, bool, Error?:
    return 0, true, null

pub func return_more_with_no_error() -> int, float, bool:
    return 0, 0.5, false
```

The number of returned values must match the function declaration.

```python
func bad() -> int, Error?:
    return 0 # invalid
```

Types must match positionally.

```python
func bad() -> int, bool:
    return true, 0 # invalid
```

Multiple returned values can be assigned together.

```python
value, err = file.read_text("./settings.json")
```

`_` discards a returned value.

```python
_, err = file.read_text("./settings.json")
value, _ = cache.get("theme")
```

Values assigned to `_` cannot be read later.

---

## Error Handling

LealLang uses explicit error values instead of exceptions.

The built-in `Error` type is:

```python
type Error:
    pub message: string
    pub code: string
```

Functions that can fail return an `Error?` as one of their return values, usually last.

```python
func read_config(path: string) -> string, Error?:
    content, err = file.read_text(path)
    if err != null:
        return "", err

    return content, null
```

Callers handle errors explicitly.

```python
value, err = read_config("./settings.json")
if err != null:
    console.print_ln(err.message)
    return

console.print_ln(value)
```

Multiple fallible steps keep their error values explicit.

```python
func load_config(path: string) -> Config, Error?:
    text, err = file.read_text(path)
    if err != null:
        return default_config(), err

    config, parse_err = json.parse(text)
    if parse_err != null:
        return default_config(), parse_err

    return config, null
```

There is no `try`, `catch`, `finally`, `throw`, or hidden error propagation syntax.

---

## async func

`async func` declares work that can run outside any window actor.

```python
async func load_settings(path: string) -> Settings, Error?:
    text, err = file.read_text(path)
    if err != null:
        return Settings(theme: "light", autosave: false), err

    settings, parse_err = json.parse(text)
    return settings, parse_err
```

Calling an `async func` immediately returns a `Task`.

```python
task = load_settings("./settings.json")
settings, err = await task
```

Allowed inside `async func`:

- primitives: `char`, `string`, `int`, `float`, `bool`, `null`
- records whose fields are sendable
- `List<T>` and `Dict<K, V>` when their contents are sendable
- pure/helper functions that do not touch UI
- task-safe built-ins such as `file`, `json`, and selected `system` calls

Forbidden inside `async func`:

- UI refs such as `@Button[id]` or `@Window[id]`
- UI-affine namespaces such as `window`, `msg`, `modal`, and `toast`
- direct component reads or writes
- `ref` parameters
- shared mutable captures from a window or event handler

---

## ui func

A `ui func` is a UI-facing function owned by a specific window. It is declared inside the owning `Window[...]` block.

```python
Window[main]:
    title = "App"

    Label[theme_label]:
        text = "Light"

    Label[status_label]:
        text = "Ready"

    pub ui func apply_settings(settings: Settings) -> Error?:
        @Label[theme_label].text = settings.theme
        @Label[status_label].text = "Settings saved"
        return null
```

Rules:

- A `ui func` runs on its owning window actor.
- A `ui func` may use local UI refs for its owning window, such as `@Label[theme_label]`.
- Public cross-window UI updates should be exposed through `pub ui func`.
- Other windows should not directly mutate another window's components.
- A `ui func` should not do slow blocking work directly. It should call or await an `async func` for slow work.

Cross-window calls use the existing UI reference form:

```python
@Window[main].apply_settings(settings)
```

This call queues work onto the target window actor and returns a `Task`. Awaiting is optional.

---

## func inside Window

Regular (non-UI) functions can also be declared inside a `Window[...]` block. These are local helper functions that are not UI-facing and do not run on the window actor. They are useful for event handlers and internal logic.

```python
Window[main]:
    title = "App"
    on_resize = handle_resize

    ui func handle_save():
        @Label[title].text = "Saved!"

    func handle_resize():
        console.print_ln("Window resized")
```

Rules:

- A bare `func` inside a Window block is private to that window.
- It does not run on the window actor (unlike `ui func`).
- It cannot use `@ComponentRef[id]` syntax to mutate UI components.
- It can be used as an event handler via `on event = handler_name`.

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

## await

`await` suspends the current continuation without blocking the window actor.

```python
func load_clicked():
    @Label[status_label].text = "Loading..."

    task = load_settings("./settings.json")
    settings, err = await task

    if err != null:
        msg.error(err.message)
        @Label[status_label].text = "Load failed"
        return

    @Window[main].apply_settings(settings)
    @Label[status_label].text = "Loaded"
```

`await` yields normal result values plus trailing `Error?`.

```python
task: Task<int> = count_files("./data")
count, err = await task
```

For an `async func` that only returns `Error?`, awaiting yields just the error value.

```python
async func save_settings_to_disk(settings: Settings) -> Error?:
    text, err = json.stringify(settings)
    if err != null:
        return err

    return file.write_text("./settings.json", text)

err = await save_settings_to_disk(settings)
```

There are no exceptions, thrown errors, or hidden propagation.
