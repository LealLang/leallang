# json

The `json` namespace provides JSON parsing and serialization.

## parse

Parses a JSON string into an `any` value.

```python
result = json.parse(text)
```

**Signature:**
```python
json.parse(value: string) -> (any, Error?)
```

## stringify

Serializes a value to a JSON string.

```python
result = json.stringify(config)
```

**Signature:**
```python
json.stringify(value: any) -> (string, Error?)
```
