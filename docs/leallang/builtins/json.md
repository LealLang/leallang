# json

The `json` namespace provides JSON parsing and serialization.

## parse

Parses a JSON string into a typed value.

```python
config, err = json.parse<Config>(text)
if err != null:
    console.print_ln(err.message)
```

**Signature:**
```python
json.parse<T>(value: string) -> T, Error?
```

## stringify

Serializes a value to a JSON string.

```python
text, err = json.stringify<Config>(config)
if err != null:
    console.print_ln(err.message)
```

**Signature:**
```python
json.stringify<T>(value: T) -> string, Error?
```
