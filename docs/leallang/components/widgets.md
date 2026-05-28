# Widgets

Widgets are leaf UI components. They do not accept child components unless their entry states otherwise.

> [!NOTE]
> Access examples using local `@Component[id]` assume they are inside an event handler for the window containing that component. Outside that context, use `@Window[window_id].ComponentType[component_id]`.

## Label

Primary argument:

- `text: string`

Properties:

- `dock: Dock?`
- `w: int?`
- `h: int?`
- `color`: a `Color` value or hex color literal string
- `tooltip: string?`

Events:

- none

Children:

- none

Example:

```python
Label[greeting_label]("Hello, world")
Label[welcome_label]($"Welcome, {username}", dock: dock.center)
```

## Image

Primary argument:

- none

Properties:

- `src: string`
- `w: int?`
- `h: int?`
- `dock: Dock?`
- `tooltip: string?`

Events:

- none

Children:

- none

Example:

```python
Image[logo](src: "./logo.png", w: 200, h: 100)
Image[remote_image](src: "https://example.com/image.png", dock: dock.fill)
```

## ProgressBar

Primary argument:

- none

Properties:

- `value: int`
- `min: int = 0`
- `max: int = 100`
- `w: int?`
- `h: int?`

Events:

- none

Children:

- none

Example:

```python
ProgressBar[load_bar](value: 75, min: 0, max: 100)
```

## Button

Primary argument:

- `text: string`

Properties:

- `enabled: bool = true`
- `tooltip: string?`
- `dock: Dock?`
- `w: int?`
- `h: int?`

Events:

- `click: ClickEvent`
- `hover: HoverEvent`
- `leave: LeaveEvent`

Children:

- none

Example:

```python
Button[save_button]("Save", click: save_clicked)
Button[help_button]("Help", tooltip: "Open help")
```

## TextInput

Primary argument:

- none

Properties:

- `value: string = ""`
- `placeholder: string?`
- `enabled: bool = true`
- `w: int?`
- `h: int?`

Events:

- `change: ChangeEvent`
- `focus: FocusEvent`

Children:

- none

Example:

```python
TextInput[name_input](placeholder: "Enter your name", change: name_changed)
```

Access current value:

```python
value = @TextInput[name_input].value
```

## TextArea

Primary argument:

- none

Properties:

- `value: string = ""`
- `placeholder: string?`
- `w: int?`
- `h: int?`

Events:

- `change: ChangeEvent`
- `focus: FocusEvent`

Children:

- none

Example:

```python
TextArea[notes_area](placeholder: "Write notes here", w: 400, h: 600, change: notes_changed)
```

Access current value:

```python
notes = @TextArea[notes_area].value
```

## Checkbox

Primary argument:

- none

Properties:

- `label: string`
- `checked: bool = false`
- `enabled: bool = true`

Events:

- `change: ChangeEvent`

Children:

- none

Example:

```python
Checkbox[agree_box](label: "I agree", change: agreement_changed)
```

Access checked state:

```python
checked = @Checkbox[agree_box].checked
```

## RadioButton

Primary argument:

- none

Properties:

- `options: Dict<string, string>`
- `selected_text: string?`
- `selected_value: string?`

Events:

- `change: ChangeEvent`

Children:

- none

Example:

```python
options: Dict<string, string> = {
    "Easy": "easy",
    "Medium": "medium",
    "Hard": "hard"
}

RadioButton[difficulty_radio](options: options, change: difficulty_changed)
```

Access selection:

```python
text = @RadioButton[difficulty_radio].selected_text
value = @RadioButton[difficulty_radio].selected_value
```

## Toggle

Primary argument:

- none

Properties:

- `label: string`
- `on: bool = false`
- `enabled: bool = true`

Events:

- `change: ChangeEvent`

Children:

- none

Example:

```python
Toggle[dark_mode_toggle](label: "Dark Mode", change: dark_mode_changed)
```

Access state:

```python
is_on = @Toggle[dark_mode_toggle].on
```

## Slider

Primary argument:

- none

Properties:

- `value: int?`
- `min: int`
- `max: int`
- `step: int = 1`

Events:

- `change: ChangeEvent`

Children:

- none

Example:

```python
Slider[volume_slider](min: 0, max: 100, step: 1, change: volume_changed)
```

Access value:

```python
volume = @Slider[volume_slider].value
```

## Dropdown

Primary argument:

- none

Properties:

- `options: List<string>`
- `selected: string?`

Events:

- `change: ChangeEvent`

Children:

- none

Example:

```python
choices: List<string> = ["Light", "Dark", "System"]
Dropdown[theme_dropdown](options: choices, change: theme_changed)
```

## NumberInput

Primary argument:

- none

Properties:

- `value: int?`
- `min: int?`
- `max: int?`
- `step: int = 1`

Events:

- `change: ChangeEvent`

Children:

- none

Example:

```python
NumberInput[retry_count](min: 0, max: 10, step: 1, change: retry_changed)
```

## Shared Widget Properties

Many widgets can use layout and sizing properties such as `dock`, `w`, and `h`.

```python
Button[save_button]("Save", dock: dock.right, w: 120)
```

Many widgets can use `tooltip`.

```python
Button[help_button]("Help", tooltip: "Open help")
```

Color properties accept typed color constants or hex color literal strings.

```python
Label[title_label]("Settings", color: colors.black)
Label[warning_label]("Warning", color: "#ff0000ff")
```
