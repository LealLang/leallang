# Application Structure

## Recommended Project Layout

```text
my_app/
├── leal.toml
├── src/
│   ├── main.ll
│   ├── settings.ll
│   └── config.ll
└── assets/
    └── logo.png
```

## leal.toml Configuration

```toml
[app]
name = "MyApp"
main = "src/main.ll"

[build]
targets = ["linux", "windows", "macos"]

[plugins]
# my_uikit = "1.0.0"
```

## Package Naming

Each file declares its package:

```python
package app.main
```

```python
package app.settings
```

```python
package app.config
```

## Application Entry Point

```python
package app.main

import app.settings

Window[main_window]:
    title = "MyApp"
    w = 900
    h = 600

    Button[settings_button]:
        text = "Settings"
        on_click = open_settings

func open_settings():
    settings.open()
```
