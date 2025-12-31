# GEMINI.md - Context & Instructions for AI Assistants

## Identity
You are an intelligent coding assistant helping to build `k8s-diff-tool`. Your goal is to write high-quality, idiomatic Go code to create a tool that semantically compares Kubernetes resources.

## Project Vision
To build a tool superior to generic text diffs for Kubernetes. It should semantic understanding of K8s resources (ignoring irrelevant fields like `managedFields`, `resourceVersion`, `creationTimestamp` unless requested).

## Tech Stack
- **Language**: Go (Latest stable version).
- **Core Libraries**:
    - `k8s.io/client-go`: For interacting with clusters.
    - `k8s.io/apimachinery`: For handling K8s objects and schemes.
    - `github.com/spf13/cobra`: For CLI command structure.
    - `github.com/spf13/viper`: For configuration management.
    - `github.com/google/go-cmp`: For deep equality checks and diff generation.

## Development Principles

### 1. Semantic Diffing
- The core value proposition is *context-aware* diffing.
- **Ignore Noise**: Automatically filter out metadata fields that change frequently but don't affect state (e.g., `generation`, `uid`, `resourceVersion`, `managedFields`).
- **Structure**: Compare specific resources structures (Deployment, Service) intelligently.
- **Filtering**: Supports inclusion and exclusion of resources by Kind.
- **Masking**: Sensitive data in `Secrets` and `ConfigMaps` is masked using a length-preserving hash-suffix method to allow diff visibility without leaking content.

### 2. Implementation Strategy
- **CLI First**: Application is structured using Cobra.
- **Modular Internal**:
    - `internal/loader`: Responsible for reading and parsing Kubernetes YAMLs.
    - `internal/differ`: Core logic for comparison, filtering, and masking.
- **Output Formats**: Supports human-readable colorized diffs.

### 3. User Experience
- The tool involves complex diff outputs; prioritize clarity.
- Use colors (red/green) for removals/additions.
- Use length-preserving masks for sensitive data to maintain visual context.

## Project Status
- [x] Initialize the project with `go mod`.
- [x] Set up the directory structure (`internal/`, `cmd/`, `main.go`).
- [x] Create the main entry point with Cobra.
- [x] Implement YAML loader with support for multi-document files.
- [x] Implement semantic diffing engine using `google/go-cmp`.
- [x] Add resource filtering by Kind.
- [x] Implement sensitive data masking for Secrets and ConfigMaps.

## Next Steps
- [ ] Add support for live cluster comparison using `client-go`.
- [ ] Implement Helm chart rendering and comparison.
- [ ] Add configurable masking rules via config file.
- [ ] Enhance diff output with summary tables.

## Knowledge Base
- **K8s API**: Familiarity with `apiVersion`, `kind`, and resource-specific schemas.
- **Go idioms**: Use of functional options for `Diff` configuration. Modular package design.

## Rules for AI
- **Always** verify file paths before writing.
- **Always** run `go mod tidy` if new dependencies are added.
- **Check** for existing code to avoid duplication.
