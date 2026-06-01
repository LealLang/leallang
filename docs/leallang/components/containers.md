# Containers

Containers organize child components. They use indented properties for layout, sizing, and behavior.

## Panel

Properties:

- `bg`: a `Color` value or hex color literal string
- `dock: Dock?`
- `w: int?`
- `h: int?`
- `label: string?`

Events:

- none

Children:

- any component

Example:

```python
Panel[side_panel]:
    bg = colors.white
    dock = dock.left
    w = 400

    Label[side_title]:
        text = "Menu"
```

## ResizablePanel

Properties:

- `min_w: int?`
- `max_w: int?`
- `min_h: int?`
- `max_h: int?`
- `dock: Dock?`

Events:

- `resize: ResizeEvent`

Children:

- any component

Example:

```python
ResizablePanel[main_area]:
    min_w = 200
    max_w = 800
    min_h = 100

    Label[content_label]:
        text = "Content"
```

## Row

Properties:

- `gap: int = 0`
- `dock: Dock?`
- `w: int?`
- `h: int?`

Events:

- none

Children:

- any component

Example:

```python
Row[action_row]:
    gap = 8

    Button[ok_button]:
        text = "OK"
    Button[cancel_button]:
        text = "Cancel"
```

## Col

Properties:

- `gap: int = 0`
- `dock: Dock?`
- `w: int?`
- `h: int?`

Events:

- none

Children:

- any component

Example:

```python
Col[form_col]:
    gap = 8

    Label[name_label]:
        text = "Name"
    TextInput[name_input]:
        placeholder = "Enter your name"
```

## Grid

Properties:

- `cols: int`
- `rows: int?`
- `gap: int = 0`
- `dock: Dock?`

Events:

- none

Children:

- any component

Example:

```python
Grid[button_grid]:
    cols = 3
    rows = 2
    gap = 8

    Button[a_button]:
        text = "A"
    Button[b_button]:
        text = "B"
    Button[c_button]:
        text = "C"
```

## ScrollPanel

Properties:

- `dock: Dock?`
- `w: int?`
- `h: int?`

Events:

- `scroll: ScrollEvent`

Children:

- any component

Example:

```python
ScrollPanel[scroll_area]:
    dock = dock.fill
    on scroll = on_scroll

    Label[item_label]:
        text = "Scrollable content"
```

## StackLayout

`StackLayout` overlays children in declaration order.

Properties:

- `dock: Dock?`
- `w: int?`
- `h: int?`

Events:

- none

Children:

- any component

Example:

```python
StackLayout[hero_stack]:
    Image[background_image]:
        src = "./bg.png"
        dock = dock.fill
    Label[overlay_label]:
        text = "Overlay text"
        dock = dock.center
```

The name `StackLayout` is used instead of `Stack` to avoid confusion with stack data structures.

## Line

Properties:

- `orientation: Orientation`
- `thickness: int = 1`
- `color`: a `Color` value or hex color literal string
- `size: int?`

Events:

- none

Children:

- none

Example:

```python
Line[divider]:
    orientation = orientation.horizontal
    thickness = 1
    color = colors.black
    size = 600

Line[vertical_divider]:
    orientation = orientation.vertical
    thickness = 2
    color = "#ccccccff"
    size = 500
```
