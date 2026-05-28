# Full Example

A complete working application demonstrating config loading, settings management, and main window integration.

## src/config.ll

```python
package app.config

pub type Config:
    pub theme: string
    pub autosave: bool

pub func default_config() -> Config:
    return Config(theme: "light", autosave: false)

pub func load(path: string) -> Config, Error?:
    text, err = file.read_text(path)
    if err != null:
        return default_config(), err

    config, parse_err = json.parse<Config>(text)
    if parse_err != null:
        return default_config(), parse_err

    return config, null
```

## src/settings.ll

```python
package app.settings

import app.config

pub Window[settings]("MyApp - Settings", w: 600, h: 400):
    Panel[main_panel](bg: colors.white, dock: dock.fill):
        Label[title_label]("Settings")
        TextInput[theme_input](placeholder: "Theme", change: theme_changed)
        Checkbox[autosave_box](label: "Autosave", change: autosave_changed)
        Row[action_row](gap: 8):
            Button[save_button]("Save", click: save_clicked)
            Button[close_button]("Close", click: close_clicked)

pub func open():
    window.open(@Window[settings])

pub func load_settings():
    config_value, err = config.load("./settings.json")
    if err != null:
        msg.error(err.message)
        return

    @Window[settings].TextInput[theme_input].value = config_value.theme
    @Window[settings].Checkbox[autosave_box].checked = config_value.autosave

func theme_changed(e: ChangeEvent):
    value = @TextInput[theme_input].value
    console.print_ln($"Theme changed to {value}")

func autosave_changed(e: ChangeEvent):
    enabled = @Checkbox[autosave_box].checked
    console.print_ln($"Autosave: {enabled}")

func save_clicked():
    theme = @TextInput[theme_input].value
    autosave = @Checkbox[autosave_box].checked

    new_config = config.Config(theme: theme, autosave: autosave)
    text, err = json.stringify<config.Config>(new_config)
    if err != null:
        msg.error(err.message)
        return

    write_err = file.write_text("./settings.json", text)
    if write_err != null:
        msg.error(write_err.message)
        return

    toast.show(message: "Settings saved", duration: 3, type: toast_type.success)

func close_clicked():
    window.close(@Window[settings])
```

## src/main.ll

```python
package app.main

import app.settings

Window[main_window]("MyApp", w: 900, h: 600, resize: resized):
    Panel[side_panel](bg: colors.white, dock: dock.left, w: 240):
        Label[app_title]("MyApp")
        Button[settings_button]("Settings", click: settings_clicked, hover: settings_hovered, leave: settings_left)

    Panel[content_panel](bg: "#f5f5f5ff", dock: dock.fill):
        Label[welcome_label]("Welcome", dock: dock.center)

func settings_clicked():
    settings.open()
    settings.load_settings()

func settings_hovered(e: HoverEvent):
    @Button[settings_button].text = "Open Settings"

func settings_left():
    @Button[settings_button].text = "Settings"

func resized(e: ResizeEvent):
    msg.info($"Window resized to {e.width}x{e.height}")
```
