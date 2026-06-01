# Plugins

Plugins allow custom UI components to be added to LealLang.

Plugins are not part of the core language semantics, but they should respect the same component property and event checking model.

## Package Structure

A plugin package uses `.llpkg`.

```text
my_plugin.llpkg
├── manifest.json
├── color_picker.lua
├── color_picker.html
└── color_picker.css
```

## Manifest

```json
{
  "name": "MyUIKit",
  "id": "com.example.myuikit",
  "version": "1.0.0",
  "author": "Example Author",
  "description": "A collection of custom UI components",
  "components": [
    {
      "name": "ColorPicker",
      "script": "color_picker.lua",
      "template": "color_picker.html",
      "styles": "color_picker.css",
      "properties": [
        { "name": "value", "type": "string", "default": "#ffffffff" },
        { "name": "disabled", "type": "bool", "default": false }
      ],
      "events": [
        { "name": "change", "event_type": "ChangeEvent" },
        { "name": "close" }
      ]
    }
  ]
}
```

## HTML Template

```html
<div class="color-picker">
  <input
    type="color"
    id="picker"
    onchange="LealLang.trigger('change', this.value)"
  />
  <span id="value_display"></span>
</div>
```

## Lua Component Script

```lua
function on_init(props)
    Component.SetHtml("value_display", props.value)
    Component.SetAttr("picker", "value", props.value)
end

function on_property_change(name, value)
    if name == "value" then
        Component.SetHtml("value_display", value)
        Component.SetAttr("picker", "value", value)
    end
end

function on_change(new_value)
    Component.SetProperty("value", new_value)
    Component.FireEvent("change")
end
```

## Plugin Usage

```python
package app.main

import plugin.my_uikit

Window[main_window]:
    title = "Plugin Example"

    ColorPicker[color_picker]:
        value = "#ff0000ff"
        on change = color_changed

func color_changed():
    console.print_ln(@ColorPicker[color_picker].value)
```

Plugin components provide property and event metadata so LealLang can type-check usage.
