# window

The `window` namespace manages window lifecycle.

## open

Opens a window.

```python
window.open(@Window[settings])
```

## close

Closes a window.

```python
window.close(@Window[settings])
```

## Usage Example

```python
pub func open_settings():
    window.open(@Window[settings])

func close_settings():
    window.close(@Window[settings])
```
