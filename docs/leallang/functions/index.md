# Functions

Functions use the `func` keyword.

```python
func greet(name: string):
    console.print_ln($"Hello, {name}")
```

Public functions use `pub`.

```python
pub func open():
    window.open(@Window[settings])
```

A function without `-> Type` returns no value.

```python
func greet(name: string):
    console.print_ln($"Hello, {name}")
```

A function that returns a value must declare its return type.

```python
func add(a: int, b: int) -> int:
    return a + b
```

Returning a value without a declared return type is invalid.

```python
func add(a: int, b: int):
    return a + b # invalid
```

`return` by itself exits a no-value function.

```python
func load_settings():
    config_value, err = config.load("./settings.json")
    if err != null:
        msg.error(err.message)
        return

    @Window[settings].TextInput[theme_input].value = config_value.theme
```

## Arguments

Positional arguments must come before named arguments.

```python
toast.show("Saved", duration: 3)           # valid
toast.show(message: "Saved", duration: 3)  # valid
toast.show(message: "Saved", 3)            # invalid
```

After the first named argument, all remaining arguments must be named.

Component declarations have stricter positional argument rules. See [Components](../components/index.md).

## ref Parameters

Use `ref` to pass a mutable alias to a variable.

```python
func increment(ref value: int):
    value = value + 1

count = 1
increment(count)
console.print_ln(count) # 2
```

Only variables can be passed to `ref` parameters.

```python
increment(count)     # valid
increment(1)         # invalid
increment(get_num()) # invalid
```

---

## Multiple Return Values

Functions can return multiple values.

The return signature must explicitly declare every returned type.

```python
pub func return_zero() -> int:
    return 0

pub func return_with_error() -> int, Error?:
    return 0, null

pub func return_more_with_error() -> int, bool, Error?:
    return 0, true, null

pub func return_more_with_no_error() -> int, float, bool:
    return 0, 0.5, false
```

The number of returned values must match the function declaration.

```python
func bad() -> int, Error?:
    return 0 # invalid
```

Types must match positionally.

```python
func bad() -> int, bool:
    return true, 0 # invalid
```

Multiple returned values can be assigned together.

```python
value, err = file.read_text("./settings.json")
```

`_` discards a returned value.

```python
_, err = file.read_text("./settings.json")
value, _ = cache.get("theme")
```

Values assigned to `_` cannot be read later.

---

## Error Handling

LealLang uses explicit error values instead of exceptions.

The built-in `Error` type is:

```python
type Error:
    pub message: string
    pub code: string
```

Functions that can fail return an `Error?` as one of their return values, usually last.

```python
func read_config(path: string) -> string, Error?:
    content, err = file.read_text(path)
    if err != null:
        return "", err

    return content, null
```

Callers handle errors explicitly.

```python
value, err = read_config("./settings.json")
if err != null:
    console.print_ln(err.message)
    return

console.print_ln(value)
```

Multiple fallible steps keep their error values explicit.

```python
func load_config(path: string) -> Config, Error?:
    text, err = file.read_text(path)
    if err != null:
        return default_config(), err

    config, parse_err = json.parse<Config>(text)
    if parse_err != null:
        return default_config(), parse_err

    return config, null
```

There is no `try`, `catch`, `finally`, `throw`, or hidden error propagation syntax.
