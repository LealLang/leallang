# Complex Components

Complex components compose widgets and containers into higher-level UI structures.

## Tabs

Properties:

- `dock: Dock?`

Events:

- `change: ChangeEvent`

Children:

- `Panel` children used as tab pages

Example:

```python
Tabs[main_tabs]:
    on change = tab_changed

    Panel[home_tab]:
        label = "Home"
        dock = dock.fill

        Label[home_label]:
            text = "Welcome"
    Panel[settings_tab]:
        label = "Settings"
        dock = dock.fill

        Label[settings_label]:
            text = "Settings here"
```

## MenuBar

Properties:

- `dock: Dock?`

Events:

- none

Children:

- `Menu`

Example:

```python
MenuBar[main_menu]:
    Menu[file_menu]:
        label = "File"

        MenuItem[open_item]:
            text = "Open"
            on click = open_clicked
        MenuItem[save_item]:
            text = "Save"
            on click = save_clicked
        MenuSeparator[file_separator]
        MenuItem[exit_item]:
            text = "Exit"
            on click = exit_clicked
    Menu[help_menu]:
        label = "Help"

        MenuItem[about_item]:
            text = "About"
            on click = about_clicked
```

## Menu

Properties:

- `label: string`

Events:

- none

Children:

- `MenuItem`
- `MenuSeparator`

Example:

```python
Menu[file_menu]:
    label = "File"

    MenuItem[open_item]:
        text = "Open"
        on click = open_clicked
```

## MenuItem

Properties:

- `text: string`
- `enabled: bool = true`
- `shortcut: string?`

Events:

- `click: ClickEvent`

Children:

- none

Example:

```python
MenuItem[save_item]:
    text = "Save"
    on click = save_clicked
```

## MenuSeparator

Properties:

- none

Events:

- none

Children:

- none

Example:

```python
MenuSeparator[file_separator]
```

## ContextMenu

Properties:

- none

Events:

- none

Children:

- `MenuItem`
- `MenuSeparator`

Example:

```python
Button[target_button]:
    text = "Right-click me"
    context_menu = @ContextMenu[action_menu]

ContextMenu[action_menu]:
    MenuItem[copy_item]:
        text = "Copy"
        on click = copy_clicked
    MenuItem[paste_item]:
        text = "Paste"
        on click = paste_clicked
```

## Toolbar

Properties:

- `dock: Dock?`

Events:

- none

Children:

- `Button`
- `Line`
- other toolbar components

Example:

```python
Toolbar[main_toolbar]:
    dock = dock.top

    Button[new_button]:
        text = "New"
        on click = new_clicked
    Button[open_button]:
        text = "Open"
        on click = open_clicked
    Line[toolbar_divider]:
        orientation = orientation.vertical
        thickness = 1
        color = colors.black
    Button[save_button]:
        text = "Save"
        on click = save_clicked
```

## Modal

Properties:

- `title: string`
- `w: int?`
- `h: int?`

Events:

- none

Children:

- any component

Example:

```python
Modal[confirm_modal]:
    title = "Confirm Action"

    Label[confirm_label]:
        text = "Are you sure?"
    Row[confirm_actions]:
        dock = dock.bottom

        Button[yes_button]:
            text = "Yes"
            on click = confirm_clicked
        Button[no_button]:
            text = "No"
            on click = cancel_clicked
```

Open and close:

```python
modal.open(@Modal[confirm_modal])
modal.close(@Modal[confirm_modal])
```

## Table

Properties:

- `data: List<Dict<string, string>>`
- `columns: List<string>?`

Events:

- `click: ClickEvent`
- `change: ChangeEvent`

Children:

- none

Example:

```python
rows: List<Dict<string, string>> = [
    {"name": "Alice", "age": "30"},
    {"name": "Bob", "age": "25"}
]

Table[users_table]:
    data = rows
    on click = row_clicked
```

## Shared Properties

Many visible components support common properties such as:

- `tooltip: string?`, optional text shown when the user hovers over the component.
- `dock: Dock?`, optional docking behavior when used inside a dock-based container.
- `w: int?`, optional width.
- `h: int?`, optional height.

Example:

```python
Button[help_button]:
    text = "Help"
    tooltip = "Open help"

Panel[side_panel]:
    dock = dock.left
    w = 240
```
