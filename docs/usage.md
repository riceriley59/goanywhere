# Usage Guide

## Command Overview

```bash
goanywhere generate <input-directory> [flags]
```

### Flags

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--output` | `-o` | Output file path | `<input>/<plugin>_plugin/main.go` |
| `--import-path` | `-i` | Import path for the target package | Auto-detected from go.mod |
| `--plugin` | `-p` | Plugin type (`cgo`, `python`) | `cgo` |
| `--verbose` | `-v` | Show parsed constructs and skipped items | `false` |

## Examples

### Basic Usage

Generate CGO bindings for a package:

```bash
goanywhere generate ./mypackage
```

This creates `./mypackage/cgo_plugin/main.go`.

### Python Bindings

Generate Python ctypes bindings:

```bash
goanywhere generate ./mypackage --plugin python
```

This creates `./mypackage/python_plugin/mypackage.py`.

### Custom Output Path

```bash
goanywhere generate ./mypackage -o ./bindings/wrapper.go
```

### Specifying Import Path

If your package is outside a Go module or the import path cannot be auto-detected:

```bash
goanywhere generate ./mypackage --import-path github.com/user/project/mypackage
```

### Verbose Output

See what functions and structs are being processed:

```bash
goanywhere generate ./mypackage -v
```

## Building the Generated Code

### CGO Shared Library

After generating CGO bindings, build a shared library:

```bash
# Linux
CGO_ENABLED=1 go build -buildmode=c-shared -o libmypackage.so ./mypackage/cgo_plugin/main.go

# macOS
CGO_ENABLED=1 go build -buildmode=c-shared -o libmypackage.dylib ./mypackage/cgo_plugin/main.go

# Windows
CGO_ENABLED=1 go build -buildmode=c-shared -o mypackage.dll ./mypackage/cgo_plugin/main.go
```

### Using Python Bindings

The generated Python file loads the shared library automatically:

```python
from mypackage import add, greet, Point

# Call functions
result = add(1, 2)
message = greet("World")

# Use structs
p = Point()
p.x = 10
p.y = 20
distance = p.distance()
```

## Supported Go Types

| Go Type | CGO Mapping | Python Mapping |
|---------|-------------|----------------|
| `int`, `int8-64` | `C.longlong`, `C.int8_t`, etc. | `c_longlong`, `c_int8`, etc. |
| `uint`, `uint8-64` | `C.ulonglong`, `C.uint8_t`, etc. | `c_ulonglong`, `c_uint8`, etc. |
| `float32`, `float64` | `C.float`, `C.double` | `c_float`, `c_double` |
| `bool` | `C.bool` | `c_bool` |
| `string` | `*C.char` | `c_char_p` |
| `error` | `**C.char` (out param) | Error string |
| `*Struct` | `C.uintptr_t` (handle) | `c_size_t` (handle) |
| `[]T` | Pointer + length | `POINTER(T)` |
| `map[K]V` | `C.uintptr_t` (handle) | `c_size_t` (handle) |

## Limitations

The following Go constructs are not supported:

- Channels (`chan`)
- Function types (`func`)
- Non-empty interfaces
- Variadic functions (skipped with warning)
- Unexported functions and types

## Example Project Structure

```
mypackage/
├── mypackage.go          # Your Go code
├── cgo_plugin/
│   └── main.go           # Generated CGO bindings
└── python_plugin/
    └── mypackage.py      # Generated Python bindings
```

## Workflow

1. Write your Go library with exported functions and structs
2. Run `goanywhere generate ./mypackage`
3. Build the shared library with `go build -buildmode=c-shared`
4. Use the library from C, Python, or other languages
