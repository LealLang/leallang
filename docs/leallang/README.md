# LealLang Language Design Reference

LealLang is a typed, declarative programming language for building desktop GUI applications. It uses indentation-based scoping, colon-delimited blocks, explicit UI component references, package-based modules, records, nullable types, generic collections, and Go-like explicit error handling.

The language is designed around a simple idea: UI structure should be declared clearly, while application behavior should remain explicit, typed, and predictable.

---

## Documentation Structure

### [Core Language](core/index.md)

- [Design Goals](core/index.md#design-goals)
- [Non-Goals](core/index.md#non-goals)
- [Lexical Structure](core/lexical.md)
- [Comments](core/lexical.md#comments)
- [Naming Conventions](core/lexical.md#naming-conventions)
- [Expressions and Operators](core/expressions.md)
- [Packages and Imports](core/packages.md)
- [Visibility](core/packages.md#visibility)
- [Primitive Types](core/types.md#primitive-types)
- [Nullable Types](core/types.md#nullable-types)
- [Variables](core/types.md#variables)
- [Constants](core/types.md#constants)
- [Records](core/types.md#records)
- [Assignment, Copying, ref, and clone](core/types.md#assignment-copying-ref-and-clone)
- [Collections](core/types.md#collections)
- [Strings and Interpolation](core/types.md#strings-and-interpolation)

### [Functions](functions/index.md)

- [Functions](functions/index.md#functions)
- [Arguments](functions/index.md#arguments)
- [Multiple Return Values](functions/index.md#multiple-return-values)
- [Error Handling](functions/index.md#error-handling)
- [Control Flow](functions/control-flow.md)

### [Components](components/index.md)

- [Windows](components/index.md#windows)
- [Components](components/index.md#components)
- [Component Arguments](components/index.md#component-arguments)
- [UI References with @](components/index.md#ui-references-with-)
- [Event Handlers](components/index.md#event-handlers)
- [Widgets](components/widgets.md)
- [Containers](components/containers.md)
- [Complex Components](components/complex.md)
- [Events](components/events.md)

### [Built-in Namespaces](builtins/index.md)

- [console](builtins/console.md)
- [file](builtins/file.md)
- [window](builtins/window.md)
- [msg](builtins/msg.md)
- [json](builtins/json.md)
- [constants](builtins/constants.md)

### [Plugins](plugins/index.md)

- [Plugins](plugins/index.md)

### [Examples](examples/index.md)

- [Application Structure](examples/application-structure.md)
- [Full Example](examples/full-example.md)
