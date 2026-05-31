# window

The `window` namespace manages window lifecycle.

## open

Opens a window. Opening starts the window's actor and returns immediately — it does not block the calling window.

```python
window.open(@Window[settings])
```

## close

Closes a window. Closing cancels work owned by that window actor.

```python
window.close(@Window[settings])
```

When a window closes:

- the actor stops accepting new events
- queued `ui func` calls targeting that window are cancelled
- suspended handlers owned by that window are cancelled
- cancelled handlers do not resume into destroyed UI
- awaiters receive an explicit `Error?` with code `window_closed`

```python
err = await @Window[main].apply_settings(settings)
if err != null:
    if err.code == "window_closed":
        return

    msg.error(err.message)
```

## Concurrency Error Codes

| Code | Meaning |
| ---- | ------- |
| `window_closed` | The target or owning window was closed before the awaited work completed. |
| `task_cancelled` | The task was cancelled before completion. |
| `task_failed` | The task failed at runtime before returning normal values. |

## Usage Example

```python
pub func open_settings():
    window.open(@Window[settings])

func close_settings():
    window.close(@Window[settings])
```
