# Contributing to GoAnywhere

Thank you for your interest in contributing to GoAnywhere! This document provides guidelines and instructions for contributing.

## Getting Started

1. Fork the repository
2. Clone your fork:
   ```bash
   git clone https://github.com/YOUR_USERNAME/goanywhere.git
   cd goanywhere
   ```
3. Install dependencies:
   ```bash
   make install-deps
   ```
4. Create a branch for your changes:
   ```bash
   git checkout -b feature/your-feature-name
   ```

## Development Workflow

### Building

```bash
make build
```

### Running Tests

```bash
# Run all tests
make test

# Run tests with coverage report
make coverage-html
```

### Code Quality

Before submitting a PR, ensure your code passes all checks:

```bash
# Format code
make format

# Run linter
make lint

# Fix lint issues automatically (where possible)
make lint-fix

# Run all CI checks
make ci
```

## Code Style

- Follow standard Go conventions and idioms
- Use `gofmt` for formatting (run `make format`)
- Write meaningful commit messages
- Add tests for new functionality
- Update documentation as needed

## Pull Request Process

1. Ensure all CI checks pass
2. Update the README.md if you've added new features or changed behavior
3. Add tests for any new functionality
4. Keep PRs focused - one feature or fix per PR
5. Write a clear PR description explaining the changes

## Reporting Issues

When reporting issues, please include:

- Go version (`go version`)
- Operating system
- Steps to reproduce
- Expected vs actual behavior
- Relevant error messages or logs

## Adding New Plugins

If you're adding a new language plugin:

1. Create a new package under `internal/plugins/<language>/`
2. Implement the `core.Plugin` interface
3. Register the plugin in `init()` using `factory.Register()`
4. Add comprehensive tests
5. Update documentation in `docs/`

## Questions?

Feel free to open an issue for questions or discussions about contributing.
