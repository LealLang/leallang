# Packages and Imports

Each `.ll` file declares exactly one package.

```python
package app.settings
```

The package declaration is the first non-comment statement in a file.

```python
package app.main

import app.settings
import app.config
```

Imports use package paths.

```python
import app.settings
```

An imported package is accessed through its final path segment by default.

```python
settings.open()
```

Import aliases use `as`.

```python
import app.settings as settings
import plugin.my_uikit as ui

settings.open()
```

Aliases avoid name conflicts.

```python
import app.settings as app_settings
import plugin.settings as plugin_settings

app_settings.open()
plugin_settings.open()
```

Only `pub` symbols are accessible from other packages.

---

## Visibility

LealLang is private by default.

The following top-level declarations are private unless marked with `pub`:

- Functions
- Types
- Constants
- Variables
- Windows

Use `pub` to expose a top-level declaration across package boundaries.

```python
pub type Config:
    pub theme: string
    pub autosave: bool

pub func load(path: string) -> Config, Error?:
    text, err = file.read_text(path)
    if err != null:
        return Config(theme: "light", autosave: false), err

    config, parse_err = json.parse<Config>(text)
    if parse_err != null:
        return Config(theme: "light", autosave: false), parse_err

    return config, null

pub Window[settings]:
    title = "Settings"

    Button[close_button]:
        text = "Close"
```

Record fields are also private unless marked with `pub`.

```python
type Car:
    pub name: string
    model: string
    year: int
```

In another package:

```python
car.name   # valid
car.model  # invalid
car.year   # invalid
```

---

## Window and Component Visibility

Windows can be public.

```python
pub Window[settings]:
    title = "Settings"

    Button[close_button]:
        text = "Close"
```

Child components inside a window are not made public individually.

```python
Window[settings]:
    title = "Settings"

    pub Button[close_button]:
        text = "Close" # invalid
```

Components are private implementation details of their window. Other packages interact with a window through public functions.

```python
pub func open():
    window.open(@Window[settings])

pub func set_title(text: string):
    @Label[title_label].text = text
```

External packages call the public functions instead of directly mutating internal components.

```python
import app.settings

settings.open()
settings.set_title("Settings")
```
