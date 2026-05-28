# UI Constants and Helper Namespaces

This page documents common typed UI constants, such as `colors.white` and `dock.fill`, as well as helper namespaces such as `modal` and `toast`.

LealLang built-in constants are typed values, not arbitrary strings.

## colors

Color constants have type `Color` and are used for UI styling.

```python
colors.blue
colors.red
colors.white
colors.black
colors.purple
```

Examples:

- `colors.white` has type `Color`.
- `colors.blue` has type `Color`.

Color properties accept named color constants or hex color literal strings.

```python
Panel[main_panel](bg: colors.white)
Panel[content_panel](bg: "#f5f5f5ff")
```

## dock

Docking positions have type `Dock`.

```python
dock.left
dock.right
dock.center
dock.top
dock.bottom
dock.fill
```

Example:

- `dock.fill` has type `Dock`.

## orientation

Layout orientations have type `Orientation`.

```python
orientation.horizontal
orientation.vertical
```

Example:

- `orientation.horizontal` has type `Orientation`.

## modal

Modal dialog management.

```python
modal.open(@Modal[confirm_modal])
modal.close(@Modal[confirm_modal])
```

## toast

Toast notification system.

```python
toast.show(message: "File saved", duration: 3)
toast.show(message: "Something went wrong", duration: 5, type: toast_type.error)
toast.show(message: "Upload complete", duration: 3, type: toast_type.success)
```

## toast_type

Toast notification types have type `ToastType`.

```python
toast_type.error
toast_type.success
```

Example:

- `toast_type.success` has type `ToastType`.
