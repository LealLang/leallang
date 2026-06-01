# Full Example

Below is one possible full-file example that shows how a small multi-window application could look. It is not the only valid style.

## src/main.ll

```python
package app.main

import app.settings

type Settings:
    pub theme: string
    pub autosave: bool

type AppError:
    pub message: string
    pub code: int

global_state: Settings = Settings(theme: "light", autosave: false)

Window[main_window]:
    title = "MyApp"
    w = 900
    h = 600
    resize = resized

    pub ui func apply_settings(new_settings: Settings) -> Error?:
        global_state = new_settings
        @Label[theme_label].text = new_settings.theme
        return null

    Panel[side_panel]:
        bg = colors.white
        dock = dock.left
        w = 240

        Label[app_title]:
            text = "MyApp"

        Button[settings_button]:
            text = "Settings"
            on click = settings_clicked
            on hover = settings_hovered
            on leave = settings_left

    Panel[content_area]:
        bg = colors.white
        dock = dock.fill

        Label[theme_label]:
            text = global_state.theme

        Button[ok_button]:
            text = "OK"
            on click = ok_clicked

func settings_clicked():
    settings.open()

func settings_hovered(e: HoverEvent):
    @Button[settings_button].text = "Open Settings"

func settings_left():
    @Button[settings_button].text = "Settings"

func resized(e: ResizeEvent):
    msg.info($"Window is now {e.width}x{e.height}")

func ok_clicked():
    msg.info("OK clicked")

func main():
    window.open(@Window[main_window])
```

## src/settings.ll

```python
package app.settings

import app.main
import app.config

draft_theme: string = ""
draft_autosave: bool = false

Window[settings_window]:
    title = "Settings"
    w = 600
    h = 400

    pub ui func load():
        config_value, err = app.config.load("./settings.json")
        if err != null:
            @Label[status_label].text = "Load failed"
            return

        @TextInput[theme_input].value = config_value.theme
        @Checkbox[autosave_box].checked = config_value.autosave
        @Label[status_label].text = "Settings loaded"

    Col[form_col]:
        dock = dock.fill
        gap = 8

        Label[status_label]:
            text = ""

        Row[theme_row]:
            gap = 8

            Label[theme_label]:
                text = "Theme"
            TextInput[theme_input]:
                placeholder = "light"
                on change = theme_changed
                on focus = on_focus

        Checkbox[autosave_box]:
            label = "Autosave"
            on change = autosave_changed

        Button[save_button]:
            text = "Save"
            on click = save_clicked

func theme_changed(e: ChangeEvent):
    draft_theme = @TextInput[theme_input].value

func on_focus(e: FocusEvent):
    console.print_ln("Focused theme input")

func autosave_changed(e: ChangeEvent):
    draft_autosave = @Checkbox[autosave_box].checked

func save_clicked():
    new_settings, err = collect_settings()
    if err != null:
        msg.error(err.message)
        return

    err = await @Window[main_window].apply_settings(new_settings)
    if err != null:
        msg.error(err.message)
        return

    window.close(@Window[settings_window])

func collect_settings() -> Settings, Error?:
    return Settings(theme: draft_theme, autosave: draft_autosave), null

pub func open():
    window.open(@Window[settings_window])
```

## src/config.ll

```python
package app.config

type Settings:
    pub theme: string
    pub autosave: bool

type AppError:
    pub message: string
    pub code: int

pub func load(path: string) -> Settings, Error?:
    text, err = file.read_text(path)
    if err != null:
        return Settings(theme: "light", autosave: false), err

    config, parse_err = json.parse<Settings>(text)
    if parse_err != null:
        return Settings(theme: "light", autosave: false), parse_err

    return config, null

pub func save(path: string, config: Settings) -> Error?:
    text, err = json.stringify<Settings>(config)
    if err != null:
        return err

    return file.write_text(path, text)
```

## Notes

- One package per file.
- Types are created with `TypeName(field: value, ...)`.
- Errors are checked explicitly with `if err != null`.
- `await` can be used inside event handlers.
- Windows are opened by calling public functions in other packages.
- In the first version, Main owns shared application state and Settings holds local draft state.
