# Threading and Concurrency Design

## Context

LealLang is designed to make desktop GUI applications easier to build. The current language spec defines declarative windows, component references, event handlers, explicit `Error?` returns, and `window.open(@Window[id])`, but it does not yet define how windows and long-running actions avoid blocking the UI.

The first concurrency model should keep the language safe and approachable:

- Multiple windows can remain responsive at the same time.
- Slow work can run outside the UI actor that triggered it.
- UI updates stay on the owning window actor.
- Errors remain explicit values, not exceptions or hidden propagation.
- Shared mutable state is avoided in the first version.

## Goals

- Opening a second window must not block the existing window.
- Slow event-handler work should be expressible without freezing the window.
- Background work must not directly mutate UI components.
- Cross-window updates should use public UI-facing functions instead of direct component access.
- Task results and cancellation must fit LealLang's existing multiple-return and `Error?` style.

## Non-goals

- General-purpose threads exposed directly to application code.
- Locks, mutexes, channels, task groups, or shared mutable global state.
- Background tasks that directly read or mutate UI components.
- Hidden error propagation syntax such as `try`, `catch`, `throw`, or implicit awaits.
- A global app-state/store actor in the first version.

## Execution Model

Each `Window[...]` owns a UI actor. A UI actor is the serialized event queue for one window. It runs that window's event handlers, local UI updates, and `ui func` calls.

Opening a window starts that window's actor and returns immediately:

```python
func open_settings_clicked():
    window.open(@Window[settings])
```

If the Settings window runs slow synchronous code in one of its handlers, only the Settings actor is blocked. The Main window keeps processing its own events because it has a separate actor.

Slow work should still be placed in `async func`s. That keeps the originating window responsive too.

## Window-owned UI Functions

A `ui func` is a UI-facing function owned by a specific window. It is declared inside the owning `Window[...]` block.

```python
Window[main]("App"):
    Label[theme_label]("Light")
    Label[status_label]("Ready")

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

## Cross-window Calls

Cross-window calls use the existing UI reference form:

```python
@Window[main].apply_settings(settings)
```

This keeps window IDs as UI tree identifiers instead of turning them into normal variables.

Calling another window's `ui func` queues work onto that window actor. The call returns a UI task. Awaiting the task is optional.

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

If the target `ui func` returns values, awaiting yields those values plus the explicit `Error?`.

```python
Window[main]("App"):
    pub ui func current_theme() -> string, Error?:
        return app_settings.theme, null

func show_current_theme_clicked():
    task = @Window[main].current_theme()
    theme, err = await task
    if err != null:
        msg.error(err.message)
        return

    @Label[current_theme_label].text = theme
```

## Shared Application State

For the first version, the main window owns shared application state. Secondary windows keep local draft state and send validated changes to Main through `pub ui func`s.

```python
type Settings:
    pub theme: string
    pub autosave: bool

Window[main]("App"):
    pub ui func apply_settings(settings: Settings) -> Error?:
        app_settings = settings
        @Label[theme_label].text = settings.theme
        return null

Window[settings]("Settings"):
    Button[save_button]("Save", click: save_settings_clicked)

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

## Async Functions and Tasks

`async func` declares work that can run outside any window actor.

```python
async func load_settings(path: string) -> Settings, Error?:
    text, err = file.read_text(path)
    if err != null:
        return Settings(theme: "light", autosave: false), err

    settings, parse_err = json.parse<Settings>(text)
    return settings, parse_err
```

Calling an `async func` immediately returns a `Task`.

```python
task = load_settings("./settings.json")
settings, err = await task
```

`await` suspends only the current continuation. It does not block the whole window actor.

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

Task typing follows the successful return values of the async function.

```python
async func count_files(path: string) -> int, Error?:
    # ...

task: Task<int> = count_files("./data")
count, err = await task
```

For an async function that only returns `Error?`, awaiting yields just the error value.

```python
async func save_settings_to_disk(settings: Settings) -> Error?:
    text, err = json.stringify<Settings>(settings)
    if err != null:
        return err

    return file.write_text("./settings.json", text)

err = await save_settings_to_disk(settings)
```

## Error Handling

Concurrency keeps LealLang's explicit error model.

- `await` yields normal result values plus trailing `Error?`.
- Cancellation is represented as an `Error?`.
- Runtime task failures are represented as an `Error?`.
- There are no exceptions, thrown errors, or hidden propagation.

Standard concurrency-related error codes should include:

| Code | Meaning |
| ---- | ------- |
| `window_closed` | The target or owning window was closed before the awaited work completed. |
| `task_cancelled` | The task was cancelled before completion. |
| `task_failed` | The task failed at runtime before returning normal values. |

Example:

```python
err = await @Window[main].apply_settings(settings)
if err != null:
    if err.code == "window_closed":
        return

    msg.error(err.message)
```

## Safety and Sendability

`async func` is background-safe. It cannot depend on UI or shared mutable state.

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
- direct mutation of another task or window's state

Values crossing into or out of an async task must be sendable. Lists and dictionaries remain reference types inside one actor, but task boundaries deep-copy or freeze them so background work cannot race UI code.

```python
func start_import_clicked():
    files = selected_files()

    # `files` is copied into the task boundary.
    task = import_files(files)

    # Mutating local UI/window state later does not race the task.
    files.clear()

    count, err = await task
```

## Window Lifecycle and Cancellation

Closing a window cancels work owned by that window actor.

When a window closes:

- the actor stops accepting new events
- queued `ui func` calls targeting that window are cancelled
- suspended handlers owned by that window are cancelled
- cancelled handlers do not resume into destroyed UI
- awaiters receive an explicit `Error?` with code `window_closed`

Background tasks may finish internally, but their result is discarded if the awaiting window is gone.

```python
func save_clicked():
    task = save_settings_to_disk(settings)
    err = await task

    if err != null:
        msg.error(err.message)
        return

    err = await @Window[main].apply_settings(settings)
    if err != null:
        if err.code == "window_closed":
            return

        msg.error(err.message)
```

If Settings closes while `save_settings_to_disk` is running, `save_clicked` is cancelled and does not resume into the closed Settings window. If Main closes before `apply_settings` runs, awaiting that UI task returns `window_closed`.

## Example Flow

```python
type Settings:
    pub theme: string
    pub autosave: bool

async func save_settings_to_disk(settings: Settings) -> Error?:
    text, err = json.stringify<Settings>(settings)
    if err != null:
        return err

    return file.write_text("./settings.json", text)

Window[main]("LealLang App"):
    Label[theme_label]("Light")
    Button[open_settings]("Settings", click: open_settings_clicked)

    pub ui func apply_settings(settings: Settings) -> Error?:
        @Label[theme_label].text = settings.theme
        return null

func open_settings_clicked():
    window.open(@Window[settings])

Window[settings]("Settings"):
    TextInput[theme_input](placeholder: "Theme")
    Button[save_button]("Save", click: save_settings_clicked)

func save_settings_clicked():
    settings = Settings(
        theme: @TextInput[theme_input].value,
        autosave: true,
    )

    err = await save_settings_to_disk(settings)
    if err != null:
        msg.error(err.message)
        return

    err = await @Window[main].apply_settings(settings)
    if err != null:
        if err.code != "window_closed":
            msg.error(err.message)
        return

    window.close(@Window[settings])
```

Main and Settings run on separate window actors. Disk saving runs in a background task. Main is updated only through its public `ui func`.

## Spec Integration Points

Future spec updates should integrate this design into these sections:

- `docs/leallang/functions/index.md`
  - `async func`
  - `ui func`
  - `await`
  - explicit task errors
- `docs/leallang/core/types.md`
  - `Task<T>`
  - sendable values
  - task-boundary copy/freeze rules for collections
- `docs/leallang/core/expressions.md`
  - `await` expressions
  - calls to async functions
  - calls to `@Window[id].ui_func(...)`
- `docs/leallang/components/index.md`
  - window-owned actors
  - window-owned `ui func` declarations
  - cross-window communication rules
- `docs/leallang/components/events.md`
  - event handlers may suspend at `await`
  - awaiting does not block the whole window actor
- `docs/leallang/builtins/window.md`
  - non-blocking `window.open`
  - close cancellation behavior

## Implementation Notes

The current interpreter is synchronous. Implementing this design later will require changes across lexer/parser, AST, checker, and runtime:

- add `async`, `await`, and `ui` syntax
- represent `Task<T>` in the type system
- type-check async return and await result shapes
- enforce background-safety and sendability rules
- add per-window actor queues for the UI bridge
- isolate async task environments from UI/runtime mutable cells
- deep-copy or freeze values at task boundaries
- resume awaited continuations on the owning window actor

These implementation details are intentionally out of scope for this design document.
