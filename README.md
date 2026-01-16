# GoAnywhere

Generate language bindings for your Go libraries to use them from Python, C, and other languages.

GoAnywhere parses your Go packages and generates wrapper code that exposes your functions and structs to other languages through CGO shared libraries or Python ctypes bindings.

## Features

- Parse Go packages and extract exported functions, structs, and methods
- Generate CGO bindings for C-compatible shared libraries
- Generate Python ctypes bindings for direct Python integration
- Automatic type mapping between Go and target languages
- Handle complex types: slices, maps, pointers, and custom structs

## Installation

```bash
go install github.com/riceriley59/goanywhere/cmd/goanywhere@latest
```

Or build from source:

```bash
git clone https://github.com/riceriley59/goanywhere.git
cd goanywhere
make build
```

## Quick Start

```bash
# Generate CGO bindings (default)
goanywhere generate ./mypackage

# Generate Python bindings
goanywhere generate ./mypackage --plugin python

# Specify output file
goanywhere generate ./mypackage -o ./bindings/main.go
```

## Documentation

- [Usage Guide](docs/usage.md) - Detailed usage instructions and examples

## Supported Plugins

| Plugin | Description | Output |
|--------|-------------|--------|
| `cgo` | CGO/C bindings via shared library | `main.go` (build with `-buildmode=c-shared`) |
| `python` | Python ctypes bindings | `<package>.py` |

## Development

```bash
# Install dependencies
make install-deps

# Run tests
make test

# Run linter
make lint

# Run all CI checks
make ci
```

## License

MIT
