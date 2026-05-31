# Events

Events are connected by passing function references to component event properties.

```python
Button[save_button]("Save", click: save_clicked, hover: save_hovered, leave: save_left)
```

An event handler can use either no event parameter or one matching event parameter.

```python
func save_clicked():
    @Button[save_button].text = "Saved"
```

```python
func save_clicked(e: ClickEvent):
    console.print_ln($"Clicked at ({e.mouse_x}, {e.mouse_y})")
```

The event parameter type must match the event's declared event type.

```python
Button[save_button]("Save", click: save_clicked)

func save_clicked(e: ResizeEvent): # invalid for click
    console.print_ln(e.width)
```

Event handlers return no value.

## Async Event Handlers

Event handlers may suspend at `await`. Awaiting does not block the window actor — other events for the same window continue processing.

```python
func load_clicked():
    @Label[status_label].text = "Loading..."

    task = load_data("./data.json")
    data, err = await task

    if err != null:
        @Label[status_label].text = "Load failed"
        return

    @Label[status_label].text = $"Loaded {data.count} items"
```

## Click Event

Type: `ClickEvent`

```python
func save_clicked(e: ClickEvent):
    console.print_ln($"Clicked at ({e.mouse_x}, {e.mouse_y})")
```

Without event data:

```python
func save_clicked():
    @Button[save_button].text = "Saved"
```

## Hover Event

Type: `HoverEvent`

```python
func save_hovered(e: HoverEvent):
    @Button[save_button].text = "Save now"
```

## Leave Event

Type: `LeaveEvent`

```python
func save_left():
    @Button[save_button].text = "Save"
```

## Scroll Event

Type: `ScrollEvent`

```python
func on_scroll(e: ScrollEvent):
    console.print_ln($"Scrolled to Y: {e.scroll_y}, delta: {e.delta_y}")
```

## Resize Event

Type: `ResizeEvent`

```python
func on_resize(e: ResizeEvent):
    msg.info($"Window is now {e.width}x{e.height}")
```

## Focus Event

Type: `FocusEvent`

```python
func on_focus(e: FocusEvent):
    msg.info($"Focused via {e.source}")
```

## Change Event

Type: `ChangeEvent`

```python
func name_changed(e: ChangeEvent):
    value = @TextInput[name_input].value
    console.print_ln(value)
```
