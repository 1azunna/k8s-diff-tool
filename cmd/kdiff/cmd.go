package main

import (
	"fmt"
	"path/filepath"
	"sort"

	"github.com/1azunna/k8s-diff-tool/internal/cluster"
	"github.com/1azunna/k8s-diff-tool/internal/differ"
	"github.com/1azunna/k8s-diff-tool/internal/loader"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type cliOptions struct {
	dirDiff      bool
	secureMode   bool
	clusterMode  bool
	kubeContext  string
	includeKinds []string
	excludeKinds []string
}

// Entrypoint creates the root command and encapsulates its flag state.
func Entrypoint() *cobra.Command {
	opts := &cliOptions{}

	cmd := &cobra.Command{
		Use:   "kdiff [path1] [path2]",
		Short: "A tool to diff Kubernetes manifests",
		Long: `kdiff is a tool for semantically comparing Kubernetes resources.
It supports calculating diffs for individual files or entire directories.`,
		Args:         cobra.RangeArgs(1, 2),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			pathA := args[0]
			var pathB string
			if len(args) > 1 {
				pathB = args[1]
			}

			diffOpts := differ.Options{
				SecureMode:   opts.secureMode,
				IncludeKinds: opts.includeKinds,
				ExcludeKinds: opts.excludeKinds,
			}

			if opts.clusterMode {
				if len(args) != 1 {
					return fmt.Errorf("cluster mode requires exactly 1 argument (local path)")
				}
				return runClusterDiff(pathA, diffOpts, opts.kubeContext)
			}

			if len(args) != 2 {
				return fmt.Errorf("requires exactly 2 arguments (path1 path2) for file mode")
			}

			if opts.dirDiff {
				// Explicit directory mode requested
				isDirA, err := loader.IsDir(pathA)
				if err != nil {
					return fmt.Errorf("invalid path %s: %w", pathA, err)
				}
				isDirB, err := loader.IsDir(pathB)
				if err != nil {
					return fmt.Errorf("invalid path %s: %w", pathB, err)
				}

				if !isDirA || !isDirB {
					return fmt.Errorf("both arguments must be directories when -d is used")
				}

				return runDirDiff(pathA, pathB, diffOpts)
			}

			// Default file mode
			isDirA, err := loader.IsDir(pathA)
			if err == nil && isDirA {
				return fmt.Errorf("%s is a directory; use -d to diff directories", pathA)
			}

			isDirB, err := loader.IsDir(pathB)
			if err == nil && isDirB {
				return fmt.Errorf("%s is a directory; use -d to diff directories", pathB)
			}

			return runFileDiff(pathA, pathB, diffOpts)
		},
	}

	cmd.Flags().BoolVarP(&opts.dirDiff, "dir", "d", false, "Compare two directories")
	cmd.Flags().BoolVarP(&opts.secureMode, "secure", "s", false, "Mask sensitive data in Secrets and ConfigMaps")
	cmd.Flags().BoolVarP(&opts.clusterMode, "cluster-mode", "c", false, "Compare local files with live cluster resources")
	cmd.Flags().StringVar(&opts.kubeContext, "kube-context", "", "Kubernetes context to use")
	cmd.Flags().StringSliceVarP(&opts.includeKinds, "include", "i", nil, "Filter resources by Kind (case-insensitive, comma-separated)")
	cmd.Flags().StringSliceVarP(&opts.excludeKinds, "exclude", "e", nil, "Exclude resources by Kind (case-insensitive, comma-separated)")

	return cmd
}

func runFileDiff(pathA, pathB string, opts differ.Options) error {
	dataA, err := loader.LoadFile(pathA)
	if err != nil {
		return err
	}

	dataB, err := loader.LoadFile(pathB)
	if err != nil {
		return err
	}

	output, err := differ.Diff(dataA, dataB, opts)
	if err != nil {
		return err
	}

	fmt.Println(output)
	return nil
}

func runDirDiff(dirA, dirB string, opts differ.Options) error {
	filesA, err := loader.ListYAMLFiles(dirA)
	if err != nil {
		return err
	}
	filesB, err := loader.ListYAMLFiles(dirB)
	if err != nil {
		return err
	}

	// Create a union of all filenames
	allFilesMap := make(map[string]bool)
	mapA := make(map[string]bool)
	for _, f := range filesA {
		allFilesMap[f] = true
		mapA[f] = true
	}
	mapB := make(map[string]bool)
	for _, f := range filesB {
		allFilesMap[f] = true
		mapB[f] = true
	}

	var allFiles []string
	for f := range allFilesMap {
		allFiles = append(allFiles, f)
	}
	sort.Strings(allFiles)

	for _, filename := range allFiles {
		var dataA, dataB []byte

		if mapA[filename] {
			fullPathA := filepath.Join(dirA, filename)
			dataA, err = loader.LoadFile(fullPathA)
			if err != nil {
				return fmt.Errorf("error reading %s: %w", fullPathA, err)
			}
		}

		if mapB[filename] {
			fullPathB := filepath.Join(dirB, filename)
			dataB, err = loader.LoadFile(fullPathB)
			if err != nil {
				return fmt.Errorf("error reading %s: %w", fullPathB, err)
			}
		}

		diff, err := differ.Diff(dataA, dataB, opts)
		if err != nil {
			return fmt.Errorf("error diffing %s: %w", filename, err)
		}

		// Requirement: Header should be the filename before diff is displayed
		fmt.Printf("# Diff for %s:\n", filename)
		fmt.Println(diff)
		// Add a separator for readability between files
		fmt.Println("# --------------------------------------------------")
	}

	return nil
}

func runClusterDiff(path string, opts differ.Options, kubeContext string) error {
	isDir, err := loader.IsDir(path)
	if err != nil {
		return fmt.Errorf("invalid path %s: %w", path, err)
	}

	client, err := cluster.NewClient(kubeContext)
	if err != nil {
		return fmt.Errorf("failed to create cluster client: %w", err)
	}

	if isDir {
		return runClusterDirDiff(client, path, opts)
	}
	return runClusterFileDiff(client, path, opts)
}

func runClusterFileDiff(client *cluster.Client, path string, opts differ.Options) error {
	data, err := loader.LoadFile(path)
	if err != nil {
		return err
	}
	return diffLocalWithCluster(client, data, path, opts)
}

func runClusterDirDiff(client *cluster.Client, dir string, opts differ.Options) error {
	files, err := loader.ListYAMLFiles(dir)
	if err != nil {
		return err
	}
	sort.Strings(files)

	for _, filename := range files {
		fullPath := filepath.Join(dir, filename)
		data, err := loader.LoadFile(fullPath)
		if err != nil {
			return fmt.Errorf("error reading %s: %w", fullPath, err)
		}

		// Header is handled in diffLocalWithCluster? No, let's keep it here for structure consistency
		// BUT diffLocalWithCluster might output multiple resources if the file handles multiple docs.
		// So we should let it handle printing.
		if err := diffLocalWithCluster(client, data, filename, opts); err != nil {
			return err
		}
	}
	return nil
}

func diffLocalWithCluster(client *cluster.Client, localData []byte, filename string, opts differ.Options) error {
	resources, err := cluster.ParseResources(localData)
	if err != nil {
		return fmt.Errorf("failed to parse %s: %w", filename, err)
	}

	for _, localRes := range resources {
		gvk := localRes.GroupVersionKind()
		name := localRes.GetName()
		namespace := localRes.GetNamespace()
		if namespace == "" {
			namespace = "default" // fallback or rely on client default. Client handles empty anyway?
			// client.GetResource uses default if empty.
			// But creating consistent identifier for output
		}

		// Check if resource exists mainly to distinguish between "create" and "update" logic,
		// but SSA handles both. However, for diffing, we want to verify connectivity?
		// Actually, standard kubectl diff simply does SSA dry-run.

		// 1. Fetch live resource (current state)
		liveRes, err := client.GetResource(gvk.GroupVersion().String(), gvk.Kind, name, namespace)
		if err != nil && !errors.IsNotFound(err) {
			return fmt.Errorf("failed to fetch resource %s/%s: %w", namespace, name, err)
		}

		// 2. Compute Dry-Run Apply (Predicted state)
		//    If liveRes is missing, SSA Dry-Run will show the Creation result (defaults applied).
		//    If liveRes exists, SSA Dry-Run will show the Merged result.
		dryRunRes, err := client.ServerSideApplyDryRun(localRes)
		if err != nil {
			return fmt.Errorf("failed to server-side dry-run apply: %w", err)
		}

		// 3. Prepare bytes for diffing.
		//    Option A: Diff Live vs DryRun.
		//       - Default assumption for "what will change".
		//       - If Live missing: Diff Empty vs DryRun (Creation).
		//       - If Live exists: Diff Live vs DryRun (Update).

		var liveBytes []byte
		if liveRes != nil {
			// Clean live resource too?
			// The DryRun result comes from the API server and contains full metadata (creationTimestamp, uid, etc of the Live object).
			// If we diff Live vs DryRun, those fields match perfectly.
			// But managedFields might change.
			// We should strip managedFields from BOTH to avoid noise, as SSA updates managedFields.
			unstructured.RemoveNestedField(liveRes.Object, "metadata", "managedFields")
			liveBytes, _ = yaml.Marshal(liveRes.Object)
		}

		// For the "Target" (Predicted), we use the DryRun result.
		// We also strip managedFields because "kdiff" manager entry will be new.
		unstructured.RemoveNestedField(dryRunRes.Object, "metadata", "managedFields")
		predictedBytes, _ := yaml.Marshal(dryRunRes.Object)

		// We use the Differ.
		// Wait, if we use Diff(Live, Predicted), we check "What changes?".
		// This is correct.
		// But earlier we used Diff(Live, Local).
		// Local was "Target".
		// Now Predicted is "Target".
		localBytes := predictedBytes

		diff, err := differ.Diff(liveBytes, localBytes, opts)
		if err != nil {
			return err
		}

		fmt.Printf("# Diff for %s (Cluster vs Local) [%s %s/%s]:\n", filename, gvk.Kind, namespace, name)
		fmt.Println(diff)
		fmt.Println("# --------------------------------------------------")
	}
	return nil
}
