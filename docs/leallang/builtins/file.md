# file

The `file` namespace provides file system operations.

## read_text

Reads the contents of a file as a string.

```python
text, err = file.read_text("./settings.json")
if err != null:
    console.print_ln(err.message)
```

**Signature:**
```python
file.read_text(path: string) -> string, Error?
```

## write_text

Writes a string to a file.

```python
write_err = file.write_text("./settings.json", text)
if write_err != null:
    console.print_ln(write_err.message)
```

**Signature:**
```python
file.write_text(path: string, value: string) -> Error?
```

## exists

Checks if a file exists.

```python
if file.exists("./config.json"):
    console.print_ln("Config found")
```

**Signature:**
```python
file.exists(path: string) -> bool
```
