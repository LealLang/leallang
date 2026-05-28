# Control Flow

## if / else if / else

```python
if is_ready:
    console.print_ln("Ready")
else if has_error:
    console.print_ln("Error")
else:
    console.print_ln("Waiting")
```

Parentheses around conditions are optional.

```python
if (is_ready):
    console.print_ln("Ready")
```

## switch Expression

Switch expressions evaluate to a value. Expression switch arms use `->`.

```python
color = switch color_name:
    "blue" -> colors.blue
    "red" -> colors.red
    "purple" | "violet" -> colors.purple
    _ -> colors.white
```

Returned directly:

```python
func get_color_by_text(color_name: string) -> Color:
    return switch color_name:
        "blue" -> colors.blue
        "red" -> colors.red
        _ -> colors.white
```

## switch Statement

Switch statements execute blocks. Statement switch arms use `:`.

```python
switch color_name:
    "blue":
        console.print_ln("Blue selected")
    "red":
        console.print_ln("Red selected")
    _:
        console.print_ln("Default selected")
```

## Range Loop

```python
loop i in 0..10:
    console.print_ln(i)
```

The `..` range is inclusive.

```python
loop i in 0..3:
    console.print_ln(i)

# 0, 1, 2, 3
```

The `..<` range excludes the upper bound.

```python
loop i in 0..<3:
    console.print_ln(i)

# 0, 1, 2
```

Exclusive upper-bound ranges are useful for collection indexing.

```python
loop i in 0..<items.count():
    console.print_ln(items[i])
```

## Step Loop

`step` controls how values are produced.

```python
loop i in 0..10 step 2:
    console.print_ln(i)

# 0, 2, 4, 6, 8, 10
```

Descending:

```python
loop i in 10..0 step -1:
    console.print_ln(i)
```

## Collection Iteration

```python
numbers: List<int> = [1, 2, 3]

loop item in numbers:
    console.print_ln(item)
```

With step:

```python
loop item in numbers step 2:
    console.print_ln(item)
```

## Dictionary Iteration

Dictionary iteration yields key-value records.

```python
scores: Dict<string, int> = {
    "Ana": 10,
    "Bob": 8
}

loop kv in scores:
    console.print_ln($"Key: {kv.key}, Value: {kv.value}")
```

## Loop Conditions

A loop can include a `while` condition. `while` stops the loop when the condition becomes false.

```python
loop i in 0..10 step 2 while i < 8:
    console.print_ln(i)
```

A loop can include an `if` filter. `if` skips values that do not match.

```python
loop i in 0..10 step 2 while i < 8 if i != 4:
    console.print_ln(i)

# 0, 2, 6
```

When multiple loop modifiers appear, the order is `step`, then `while`, then `if`.

## Infinite Loop

```python
loop true:
    console.print_ln("Running")

    if should_stop:
        break
```

## break and continue

```python
loop i in 0..10:
    if i == 3:
        continue

    if i == 7:
        break

    console.print_ln(i)
```
