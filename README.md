# k8s-diff-tool

`k8s-diff-tool` is a powerful CLI utility designed to semantically compare Kubernetes resource configurations. Unlike generic text diff tools, it understands the structure of Kubernetes manifests, allowing it to filter out noise and protect sensitive information while highlighting meaningful changes.

## Features

- **Semantic Understanding (Not Yet Implemented)**: Automatically ignores irrelevant metadata fields like `managedFields`, `resourceVersion`, `creationTimestamp`, and `uid`.
- **Sensitive Data Masking**: Securely masks values in `Secrets` and `ConfigMaps` using a length-preserving hash-suffix method (enabled with `-s`).
- **Resource Filtering**: Include (`-i`) or exclude (`-e`) specific resource Kinds from the comparison.
- **Directory Support**: Compare two directories of YAML files (`-d`) to see differences across an entire stack.
- **Multi-Document Support**: Handles YAML files containing multiple Kubernetes resources separated by `---`.

## Installation

### Prerequisites
- [Go](https://golang.org/doc/install) 1.21+ (recommended)

### Build from source
```bash
git clone https://github.com/1azunna/k8s-diff-tool.git
cd k8s-diff-tool
go build -o bin/kdiff .
```

To make `kdiff` accessible globally, add the `bin` directory to your PATH:
```bash
export PATH="$HOME/path/to/k8s-diff-tool/bin:$PATH"
```

## Usage

```bash
kdiff [path1] [path2] [flags]
```

### Flags
- `-d, --dir`: Compare all matching YAML files in two directories.
- `-s, --secure`: Mask sensitive data in `Secrets` and `ConfigMaps`.
- `-i, --include`: Only include specific resource Kinds (e.g., `-i Deployment,Service`).
- `-e, --exclude`: Exclude specific resource Kinds (e.g., `-e Namespace`).

### Examples

#### Compare two files
```bash
kdiff production/app.yaml staging/app.yaml
```

#### Compare two directories with secure masking
```bash
kdiff -d -s test/dir_a test/dir_b
```

#### Compare directories but only show Services
```bash
kdiff -d -i Service test/dir_a test/dir_b
```

## Development

### Running Tests
To run the project unit tests:
```bash
go test ./...
```

## License
Apache 2.0
