# Usage Guide

## Commands

GoAnywhere provides two main commands:

- `generate` - Generate language binding source code
- `build` - Generate and compile bindings into distributable packages

## Generate Command

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

## Build Command

The `build` command generates binding code and compiles it into ready-to-use packages.

```bash
goanywhere build <input-directory> [flags]
```

### Flags

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--output` | `-o` | Output directory for built artifacts | `<input>/<plugin>_build` |
| `--import-path` | `-i` | Import path for the target package | Auto-detected from go.mod |
| `--plugin` | `-p` | Plugin type (`cgo`, `python`) | `cgo` |
| `--build-system` | | Python build system (`setuptools`, `hatch`, `poetry`, `uv`) | `setuptools` |
| `--lib-name` | | Override the default library name | `lib<package>` |
| `--verbose` | `-v` | Show build progress and details | `false` |

### CGO Build

Build a shared library from your Go package:

```bash
goanywhere build ./mypackage --plugin cgo
```

This generates CGO bindings and compiles them into a shared library (`.so` on Linux, `.dylib` on macOS, `.dll` on Windows).

### Python Build

Build a complete Python package with shared library:

```bash
goanywhere build ./mypackage --plugin python
```

This creates a distributable Python package structure:

```
mypackage/python_build/
├── mypackage/
│   ├── __init__.py
│   ├── bindings.py
│   └── lib/
│       └── libmypackage.so
├── cgo_plugin/
│   └── main.go
├── libmypackage.so
└── pyproject.toml
```

### Python Build Systems

Choose your preferred Python build system:

```bash
# setuptools (default)
goanywhere build ./mypackage --plugin python --build-system setuptools

# hatch
goanywhere build ./mypackage --plugin python --build-system hatch

# poetry
goanywhere build ./mypackage --plugin python --build-system poetry

# uv
goanywhere build ./mypackage --plugin python --build-system uv
```

After building, install the package:

```bash
cd mypackage/python_build
pip install -e .
```

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

> **Tip:** Use `goanywhere build` to automate these steps. See [Build Command](#build-command).

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

### Using Generate (Manual Build)

1. Write your Go library with exported functions and structs
2. Run `goanywhere generate ./mypackage`
3. Build the shared library with `go build -buildmode=c-shared`
4. Use the library from C, Python, or other languages

### Using Build (Automated)

1. Write your Go library with exported functions and structs
2. Run `goanywhere build ./mypackage --plugin python`
3. Install the package with `pip install -e ./mypackage/python_build`
4. Import and use in Python
