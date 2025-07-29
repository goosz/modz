// Package modz provides a modular, type-safe dependency injection framework for Go applications.
//
// # Overview
//
// Modz enables you to compose applications from loosely-coupled modules that declare what data
// they produce and consume. The framework builds a dependency graph, wires up modules, and
// ensures type safety for all shared data.
//
// # Key Concepts
//
//   - [Assembly]: Orchestrates the construction and wiring of modules and their dependencies.
//   - [Module]: A self-contained component that declares what data it produces and consumes.
//   - [Data]: A type-safe key and contract for sharing values between modules.
//   - [Binder]: A controlled interface for modules to access and provide data during configuration.
//
// # Module Lifecycle
//
// Each [Module] goes through two main phases orchestrated by the [Assembly]:
//  1. The module's discovery phase: The [Assembly] inspects each [Module] to determine what [Data]
//     it produces and consumes, building the dependency graph.
//  2. The module's configuration phase: The [Assembly] provides each [Module] with a [Binder] to wire
//     up its dependencies and store produced values. The [Binder] provides data access through its
//     [DataReader] and [DataWriter] interfaces. All of these methods (Install, getData, and putData)
//     are only valid during this configuration phase; calling them outside this phase is strictly
//     enforced and will result in an error.
//
// # Assembly Lifecycle
//
// The [Assembly] is responsible for orchestrating the module lifecycle. It first builds the
// dependency graph by inspecting all [Module]s, then configures each [Module] in dependency order.
// The [Assembly] itself does not manage application runtime; it focuses on construction and wiring.
//
// The Build() method of [Assembly] can only be called once per Assembly instance; subsequent calls will return an error.
//
// # Intended Usage
//
// Modz is designed for applications that benefit from modularity, clear dependency management,
// and type-safe data sharing between components. It is suitable for both small and large Go
// projects where decoupling and testability are important.
//
// # Status
//
// Modz is in early development and its API is subject to change. It is not yet recommended for
// production use.
//
// For more details, see the project's [README] and individual type documentation.
//
// [README]: https://github.com/goosz/modz#readme
package modz
