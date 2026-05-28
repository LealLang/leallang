# Complex Components

Complex components compose widgets and containers into higher-level UI structures.

## Tabs

Primary argument:

- none

Properties:

- `dock: Dock?`

Events:

- `change: ChangeEvent`

Children:

- `Panel` children used as tab pages

Example:

```python
Tabs[main_tabs](change: tab_changed):
    Panel[home_tab](label: "Home", dock: dock.fill):
        Label[home_label]("Welcome")
    Panel[settings_tab](label: "Settings", dock: dock.fill):
        Label[settings_label]("Settings here")
```

## MenuBar

Primary argument:

- none

Properties:

- `dock: Dock?`

Events:

- none

Children:

- `Menu`

Example:

```python
MenuBar[main_menu]:
    Menu[file_menu](label: "File"):
        MenuItem[open_item]("Open", click: open_clicked)
        MenuItem[save_item]("Save", click: save_clicked)
        MenuSeparator[file_separator]()
        MenuItem[exit_item]("Exit", click: exit_clicked)
    Menu[help_menu](label: "Help"):
        MenuItem[about_item]("About", click: about_clicked)
```

## Menu

Primary argument:

- none

Properties:

- `label: string`

Events:

- none

Children:

- `MenuItem`
- `MenuSeparator`

Example:

```python
Menu[file_menu](label: "File"):
    MenuItem[open_item]("Open", click: open_clicked)
```

## MenuItem

Primary argument:

- `text: string`

Properties:

- `enabled: bool = true`
- `shortcut: string?`

Events:

- `click: ClickEvent`

Children:

- none

Example:

```python
MenuItem[save_item]("Save", click: save_clicked)
```

## MenuSeparator

Primary argument:

- none

Properties:

- none

Events:

- none

Children:

- none

Example:

```python
MenuSeparator[file_separator]()
```

## ContextMenu

Primary argument:

- none

Properties:

- none

Events:

- none

Children:

- `MenuItem`
- `MenuSeparator`

Example:

```python
Button[target_button]("Right-click me", context_menu: @ContextMenu[action_menu])

ContextMenu[action_menu]:
    MenuItem[copy_item]("Copy", click: copy_clicked)
    MenuItem[paste_item]("Paste", click: paste_clicked)
```

## Toolbar

Primary argument:

- none

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
Toolbar[main_toolbar](dock: dock.top):
    Button[new_button]("New", click: new_clicked)
    Button[open_button]("Open", click: open_clicked)
    Line[toolbar_divider](orientation: orientation.vertical, thickness: 1, color: colors.black)
    Button[save_button]("Save", click: save_clicked)
```

## Modal

Primary argument:

- none

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
Modal[confirm_modal](title: "Confirm Action"):
    Label[confirm_label]("Are you sure?")
    Row[confirm_actions](dock: dock.bottom):
        Button[yes_button]("Yes", click: confirm_clicked)
        Button[no_button]("No", click: cancel_clicked)
```

Open and close:

```python
modal.open(@Modal[confirm_modal])
modal.close(@Modal[confirm_modal])
```

## Table

Primary argument:

- none

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

Table[users_table](data: rows, click: row_clicked)
```

## Shared Properties

Many visible components support common properties such as:

- `tooltip: string?`, optional text shown when the user hovers over the component.
- `dock: Dock?`, optional docking behavior when used inside a dock-based container.
- `w: int?`, optional width.
- `h: int?`, optional height.

Example:

```python
Button[help_button]("Help", tooltip: "Open help")
Panel[side_panel](dock: dock.left, w: 240)
```
