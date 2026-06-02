# Widgets

Widgets are leaf UI components. They do not accept child components unless their entry states otherwise.

## Label

Properties:

- `text: string`
- `dock: Dock?`
- `w: int?`
- `h: int?`
- `color: Color`
- `tooltip: string?`

Events:

- none

Children:

- none

Example:

```python
Label[greeting_label]:
    text = "Hello, world"

Label[welcome_label]:
    text = $"Welcome, {username}"
    dock = dock.center
```

## Image

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
Image[logo]:
    src = "./logo.png"
    w = 200
    h = 100

Image[remote_image]:
    src = "https://example.com/image.png"
    dock = dock.fill
```

## ProgressBar

Properties:

- `value: float`
- `min: float`
- `max: float`
- `w: int?`
- `h: int?`

Events:

- none

Children:

- none

Example:

```python
ProgressBar[load_bar]:
    value = 0.75
    min = 0.0
    max = 1.0
```

## Button

Properties:

- `text: string`
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
Button[save_button]:
    text = "Save"
    on_click = save_clicked

Button[help_button]:
    text = "Help"
    tooltip = "Open help"
```

## TextInput

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
TextInput[name_input]:
    placeholder = "Enter your name"
    on_change = name_changed
```

Access current value:

```python
value = @TextInput[name_input].value
```

## TextArea

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
TextArea[notes_area]:
    placeholder = "Write notes here"
    w = 400
    h = 600
    on_change = notes_changed
```

Access current value:

```python
notes = @TextArea[notes_area].value
```

## Checkbox

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
Checkbox[agree_box]:
    label = "I agree"
    on_change = agreement_changed
```

Access checked state:

```python
checked = @Checkbox[agree_box].checked
```

## RadioButton

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

RadioButton[difficulty_radio]:
    options = options
    on_change = difficulty_changed
```

Access selection:

```python
text = @RadioButton[difficulty_radio].selected_text
value = @RadioButton[difficulty_radio].selected_value
```

## Toggle

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
Toggle[dark_mode_toggle]:
    label = "Dark Mode"
    on = false
    on_change = dark_mode_changed
```

Access state:

```python
is_on = @Toggle[dark_mode_toggle].on
```

## Slider

Properties:

- `value: float`
- `min: float`
- `max: float`
- `step: float`

Events:

- `change: ChangeEvent`

Children:

- none

Example:

```python
Slider[volume_slider]:
    value = 0.5
    min = 0.0
    max = 1.0
    step = 0.1
    on_change = volume_changed
```

Access value:

```python
volume = @Slider[volume_slider].value
```

## Dropdown

Properties:

- `options: List<string>`
- `selected: int`

Events:

- `change: ChangeEvent`

Children:

- none

Example:

```python
choices: List<string> = ["Light", "Dark", "System"]
Dropdown[theme_dropdown]:
    options = choices
    on_change = theme_changed
```

## NumberInput

Properties:

- `value: float`
- `min: float?`
- `max: float?`
- `step: float`

Events:

- `change: ChangeEvent`

Children:

- none

Example:

```python
NumberInput[retry_count]:
    value = 1.0
    min = 0.0
    max = 10.0
    step = 1.0
    on_change = retry_changed
```

## Shared Widget Properties

Many widgets can use layout and sizing properties such as `dock`, `w`, and `h`.

```python
Button[save_button]:
    text = "Save"
    dock = dock.right
    w = 120
```

Many widgets can use `tooltip`.

```python
Button[help_button]:
    text = "Help"
    tooltip = "Open help"
```

Color properties accept typed color constants.

```python
Label[title_label]:
    text = "Settings"
    color = colors.black

Label[warning_label]:
    text = "Warning"
    color = colors.red
```
