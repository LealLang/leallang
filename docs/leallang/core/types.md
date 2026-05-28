# Types and Data

## Primitive Types

LealLang includes these primitive types:

```text
char
string
int
float
bool
null
any
```

### char

A single character.

```python
letter: char = 'a'
```

### string

A sequence of characters.

```python
name: string = "LealLang"
```

### int

A whole number.

```python
count: int = 10
```

### float

A decimal number.

```python
price: float = 9.99
```

### bool

A boolean value.

```python
is_ready: bool = true
```

### null

The absence of a value.

`null` can only be assigned to nullable types.

### any

`any` is a dynamic escape hatch.

```python
value: any = "hello"
value = 42
```

Type-specific operations on `any` require explicit narrowing or checked extraction.

```python
if value is string:
    console.print_ln(value)
```

Checked extraction uses `.as<T>()` and returns the extracted value with a success flag.

```python
number, ok = value.as<int>()
if ok:
    console.print_ln(number)
```

Assigning from `any` to a concrete type requires an explicit check or conversion.

```python
name: string = value # invalid

if value is string:
    name: string = value # valid
```

`any` does not silently coerce values between unrelated types.

---

## Nullable Types

LealLang supports explicit nullable types with `?`.

```python
name: string? = null
err: Error? = null
```

A non-nullable type cannot receive `null`.

```python
name: string = null   # invalid
```

A nullable type can receive either a value or `null`.

```python
name: string? = "Ana"
name = null
```

### Null Checks and Type Narrowing

After a null check, the checked value narrows inside the checked block.

```python
err: Error? = file_error()

if err != null:
    console.print_ln(err.message)
```

Before the check:

```text
err: Error?
```

Inside the `if` block:

```text
err: Error
```

This makes field access safe after checking for `null`.

---

## Variables

Variables can be declared with explicit types.

```python
name: string = "Ana"
count: int = 3
```

Local variable types can be inferred from initializers.

```python
name = "Ana"
count = 3
```

Variables are mutable by default.

```python
count = 1
count = 2
```

A variable cannot be assigned a value of an incompatible type.

```python
count: int = "three" # invalid
```

---

## Constants

Constants are immutable values.

```python
const app_name: string = "MyApp"
const max_retry_count = 3
```

Constants cannot be reassigned.

```python
const max_count = 10
max_count = 20 # invalid
```

Public constants use `pub`.

```python
pub const default_theme = "light"
```

---

## Records

Records define structured data.

```python
type Config:
    pub theme: string
    pub autosave: bool
```

Record fields are private by default. Use `pub` to make a field accessible from other packages.

```python
type Car:
    pub name: string
    model: string
    year: int
```

### Record Construction

Record construction requires named fields.

```python
config = Config(theme: "light", autosave: false)
```

Positional record construction is invalid.

```python
config = Config("light", false) # invalid
```

All non-nullable fields must be initialized. Nullable fields can be omitted and default to `null`.

```python
type User:
    pub name: string
    pub email: string?

user = User(name: "Ana", email: null) # valid
user = User(name: "Ana")              # valid, email defaults to null
user = User(email: null)               # invalid, name is required
```

Unknown fields are invalid.

```python
user = User(name: "Ana", age: 30) # invalid
```

Repeated fields are invalid.

```python
user = User(name: "Ana", name: "Bob") # invalid
```

Qualified record construction uses the imported package name or alias.

```python
import app.config

settings = config.Config(theme: "light", autosave: false)
```

### Record Mutation

Records are mutable.

```python
car = Car(name: "Fiesta", model: "SE", year: 2018)
car.year = 2020
```

Private fields can only be accessed within the same package.

---

## Assignment, Copying, ref, and clone

LealLang distinguishes between value-like records and reference-like collections.

### Record Assignment

Records are mutable value types.

Assigning a record copies it.

```python
car = Car(name: "Fiesta", model: "SE", year: 2018)
b = car

car.year = 2020

console.print_ln(car.year) # 2020
console.print_ln(b.year)   # 2018
```

### Function Parameters

Function parameters receive copies by default.

```python
func update_year(car: Car):
    car.year = 2020

car = Car(name: "Fiesta", model: "SE", year: 2018)
update_year(car)

console.print_ln(car.year) # 2018
```

Use `ref` to pass a mutable alias to the caller's variable.

```python
func update_year(ref car: Car):
    car.year = 2020

car = Car(name: "Fiesta", model: "SE", year: 2018)
update_year(car)

console.print_ln(car.year) # 2020
```

Only variables can be passed to `ref` parameters.

```python
update_year(car)                                      # valid
update_year(Car(name: "Fiesta", model: "SE", year: 2018)) # invalid
update_year(get_car())                                # invalid
```

### Collection Assignment

`List<T>` and `Dict<K, V>` are reference types.

Assigning a list or dictionary copies the reference, not the collection.

```python
numbers: List<int> = [1, 2, 3]
other = numbers

numbers.push(4)

console.print_ln(other.count()) # 4
```

### clone

Use `clone()` to explicitly copy collections.

```python
numbers: List<int> = [1, 2, 3]
other = numbers.clone()

numbers.push(4)

console.print_ln(numbers.count()) # 4
console.print_ln(other.count())   # 3
```

---

## Collections

LealLang supports generic collections.

### List

A list is an ordered collection of values of the same type.

```python
numbers: List<int> = [1, 2, 3]
names: List<string> = ["Ana", "Bob"]
```

List element types can be inferred from the literal.

```python
numbers = [1, 2, 3]
```

Common list methods:

```python
numbers.push(4)        # [1, 2, 3, 4]
last = numbers.pop()   # 4
numbers.insert(1, 10)  # [1, 10, 2, 3]
numbers.remove(10)     # true
index = numbers.index_of(2)
has = numbers.has_index(5)
count = numbers.count()
numbers.clear()
```

List indexing:

```python
first = numbers[0]
```

Use an exclusive upper-bound range when indexing over a list.

```python
loop i in 0..<numbers.count():
    console.print_ln(numbers[i])
```

### Dict

A dictionary maps keys to values.

```python
scores: Dict<string, int> = {
    "Ana": 10,
    "Bob": 8
}
```

Mixed values use `any` explicitly.

```python
metadata: Dict<string, any> = {
    "name": "LealLang",
    "version": 1
}
```

Common dictionary methods:

```python
score = scores["Ana"]
has = scores.has_key("Ana")
added = scores.try_add("Max", 7)
changed = scores.try_set("Ana", 11)
removed = scores.try_remove("Bob")
count = scores.count()
scores.clear()
```

Direct dictionary access assumes the key exists.

```python
score = scores["Ana"]
```

If the key is missing, direct access fails at runtime.

Safe dictionary access uses `.get`.

```python
score, err = scores.get("Ana")
if err != null:
    console.print_ln("Missing score")
```

Use `_` to discard an unneeded returned value.

```python
value, _ = cache.get("theme")
```

---

## Strings and Interpolation

String interpolation uses `$"..."`.

```python
name = "Ana"
console.print_ln($"Hello, {name}")
```

Expressions can be used inside interpolation braces.

```python
console.print_ln($"Total: {price * quantity}")
```

Interpolation works in component declarations.

```python
Label[welcome_label]($"Welcome, {user_name}")
```
