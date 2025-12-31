# Agent Guidelines for k8s-diff-tool

This document outlines the operational guidelines, coding standards, and best practices for agents (AI or human) contributing to the `k8s-diff-tool` project.

## Project Overview
`k8s-diff-tool` is a CLI utility written in Go designed to compare Kubernetes cluster resource configurations and highlight differences. It aims to provide deep, semantic diffs rather than simple text comparison.

## Coding Standards (Go)

### 1. Style & Formatting
- **Format**: Always run `gofmt` (or `goimports`) on your code before submitting.
- **Linters**: Adhere to `golangci-lint` rules.
- **Naming**:
    - Use `CamelCase` for public members and `camelCase` for private members.
    - Variable names should be short but descriptive (e.g., `ctx` for Context, `k8sClient` for Kubernetes client).
    - Avoid stuttering (e.g., `config.ConfigParser` -> `config.Parser`).

### 2. Project Structure
We follow a modular Go layout:
- `main.go`: Application entrypoint (root).
- `/cmd/kdiff`: CLI command definitions and flags.
- `/internal/loader`: Logic for loading and parsing YAML files.
- `/internal/differ`: Core diffing, filtering, and masking engine.
- `/test`: Test data and fixtures for local comparison testing.
- `/bin`: Local directory for compiled binaries (ignored by git).

### 3. Error Handling
- **Wrap Errors**: Use `fmt.Errorf("...: %w", err)` to provide context when returning errors.
- **Don't Panic**: Handle errors gracefully. Only panic during initialization if the app cannot function at all.

### 4. Concurrency
- Use `context.Context` for cancellation and timeouts.
- Prefer channels and `sync` primitives over shared state where possible.
- Avoid goroutine leaks; ensure all goroutines have a stopping mechanism.

### 5. Testing
- **Table-Driven Tests**: Use table-driven tests for comprehensive coverage.
- **Mocks**: Use interfaces to make code testable. Mock external K8s calls.
- **Coverage**: Aim for high test coverage, especially in diff logic.

## Contribution Workflow

1.  **Understand the Context**: Read `GEMINI.md` and related documentation.
2.  **Implementation**:
    - Create necessary directories if they don't exist.
    - Implement core logic in `/internal` or `/pkg`.
    - Wire up the CLI in `/cmd`.
3.  **Verification**: ensure code compiles and tests pass.

## Tools
- `kubectl`: May be used for local testing validation.
- `kind` or `minikube`: For local cluster testing.
- `dyff` or `go-cmp`: Potential libraries for diffing logic.