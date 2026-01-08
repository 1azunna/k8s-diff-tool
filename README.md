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

### Install with Go
You can install the tool directly using `go install`:
```bash
go install github.com/1azunna/k8s-diff-tool/cmd/kdiff@latest
```
This will install the `kdiff` binary to your `$GOPATH/bin` directory.

### Build from source
```bash
git clone https://github.com/1azunna/k8s-diff-tool.git
cd k8s-diff-tool
go build -o bin/kdiff ./cmd/kdiff
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
- `-c, --cluster-mode`: Compare local files with live cluster resources.
- `--kube-context`: Specify the Kubernetes context to use (only for --cluster-mode).
- `-i, --include`: Only include specific resource Kinds (e.g., `-i Deployment,Service`).
- `-e, --exclude`: Exclude specific resource Kinds (e.g., `-e Namespace`).

### Examples

#### Compare two files
```bash
kdiff production/app.yaml staging/app.yaml
```

#### Compare local file with live cluster
```bash
kdiff production/app.yaml --cluster-mode
```

#### Compare two directories with secure masking
```bash
kdiff -d -s test/dir_a test/dir_b
```

#### Compare directories but only show Services
```bash
kdiff -d -i Service test/dir_a test/dir_b
```

## GitHub Action

You can use `k8s-diff-tool` as a GitHub Action in your CI/CD workflows to automatically compare Kubernetes manifests.

### Inputs

| Input | Description | Default | Required |
|-------|-------------|---------|----------|
| `base_path` | Base file or directory path to compare | | No |
| `head_path` | Head file or directory path to compare | | No |
| `directory` | Enable directory comparison mode | `"false"` | No |
| `secure_mode` | Enable secure mode to mask secrets | `"false"` | No |
| `cluster_mode` | Enable cluster comparison mode | `"false"` | No |
| `kube_context` | Kubernetes context to use (for cluster mode) | | No |
| `include` | Comma-separated list of Kinds to include | | No |
| `exclude` | Comma-separated list of Kinds to exclude | | No |

### Outputs

| Output | Description |
|--------|-------------|
| `diff` | The captured diff output string |

### Example Usage

```yaml
name: K8s Diff Check
on: [pull_request]

jobs:
  kdiff:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          path: head-branch
      - uses: actions/checkout@v4
        with:
          ref: master
          path: base-branch

      - name: Compare manifests
        id: diff-check
        uses: 1azunna/k8s-diff-tool@main
        with:
          base_path: 'base-branch/deploy/overlays/prod'
          head_path: 'head-branch/deploy/overlays/prod'
          directory: 'true'
          secure_mode: 'true'
          
      - name: Post diff as comment
        uses: mshick/add-pr-comment@v2
        with:
          message: |
            Diff Output:

            <details><summary>Change details</summary>

            ````````diff

            ${{ steps.diff-check.outputs.diff }}

            ````````
            </details>
```

## Development

### Running Tests
To run the project unit tests:
```bash
go test ./...
```

## License
Apache 2.0
