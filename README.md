# modz

[![Build Status](https://github.com/goosz/modz/actions/workflows/test.yaml/badge.svg?branch=main)](https://github.com/goosz/modz/actions/workflows/test.yaml)
[![CodeQL](https://github.com/goosz/modz/actions/workflows/github-code-scanning/codeql/badge.svg)](https://github.com/goosz/modz/actions/workflows/github-code-scanning/codeql)
[![codecov](https://codecov.io/github/goosz/modz/graph/badge.svg?token=E82FCLR1QI)](https://codecov.io/github/goosz/modz)
[![OpenSSF Scorecard](https://api.scorecard.dev/projects/github.com/goosz/modz/badge)](https://scorecard.dev/viewer/?uri=github.com/goosz/modz)
[![OpenSSF Best Practices](https://www.bestpractices.dev/projects/10838/badge)](https://www.bestpractices.dev/projects/10838)
[![Go Report Card](https://goreportcard.com/badge/github.com/goosz/modz)](https://goreportcard.com/report/github.com/goosz/modz)
[![Go Reference](https://pkg.go.dev/badge/github.com/goosz/modz)](https://pkg.go.dev/github.com/goosz/modz)
[![pre-commit](https://img.shields.io/badge/pre--commit-enabled-brightgreen?logo=pre-commit)](https://github.com/pre-commit/pre-commit)

---

**Modz** is a modular, type-safe dependency injection framework for Go applications.

> **Status:** Modz is in very early stages of development.
> The API is experimental and subject to change.

## Overview

Modz enables you to compose applications from loosely-coupled modules that declare what data they produce and consume. The framework builds a dependency graph, wires up modules, and ensures type safety for all shared data.

- **Assembly:** Orchestrates the construction and wiring of modules and their dependencies.
- **Module:** A self-contained component that declares what data it produces and consumes.
- **Singleton:** A marker interface that allows modules to be installed multiple times without error.
- **Data:** A type-safe key and contract for sharing values between modules.
- **Binder:** A controlled interface for modules to access and provide data during configuration.

The framework includes robust validation to prevent configuration errors:
- Module uniqueness enforcement using package path + module name signatures
- Automatic detection of duplicate producers for the same data key
- Data key signature clash detection to prevent conflicts between packages
- Validation that data keys are properly declared at package level
- Support for singleton modules that can be installed multiple times without error

For a detailed API reference and technical documentation, see the [pkg.go.dev documentation](https://pkg.go.dev/github.com/goosz/modz).

## Getting Started

To install Modz, use `go get`:

    go get github.com/goosz/modz

This will then make the following packages available to you:

    github.com/goosz/modz

## Staying up to date

To update Modz to the latest version, use `go get -u github.com/goosz/modz`

## Contributing

Contributions are welcome! Please open an issue to discuss your ideas before submitting a pull request. Bug reports and suggestions are also appreciated.

## License

This project is licensed under the terms of the [MIT License](LICENSE).
