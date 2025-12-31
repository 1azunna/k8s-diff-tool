package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/1azunna/k8s-diff-tool/internal/differ"
	"github.com/1azunna/k8s-diff-tool/internal/loader"
	"github.com/spf13/cobra"
)

type cliOptions struct {
	dirDiff      bool
	secureMode   bool
	includeKinds []string
	excludeKinds []string
}

func main() {
	if err := NewRootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// NewRootCmd creates the root command and encapsulates its flag state.
func NewRootCmd() *cobra.Command {
	opts := &cliOptions{}

	cmd := &cobra.Command{
		Use:   "kdiff [path1] [path2]",
		Short: "A tool to diff Kubernetes manifests",
		Long: `kdiff is a tool for semantically comparing Kubernetes resources.
It supports calculating diffs for individual files or entire directories.`,
		Args:         cobra.ExactArgs(2),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			pathA := args[0]
			pathB := args[1]

			diffOpts := differ.Options{
				SecureMode:   opts.secureMode,
				IncludeKinds: opts.includeKinds,
				ExcludeKinds: opts.excludeKinds,
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

	// Create a map for quick lookup of files in B
	mapB := make(map[string]bool)
	for _, f := range filesB {
		mapB[f] = true
	}

	// Intersection: Iterate over A and check if present in B
	for _, filename := range filesA {
		if _, exists := mapB[filename]; exists {
			fullPathA := filepath.Join(dirA, filename)
			fullPathB := filepath.Join(dirB, filename)

			dataA, err := loader.LoadFile(fullPathA)
			if err != nil {
				return fmt.Errorf("error reading %s: %w", fullPathA, err)
			}
			dataB, err := loader.LoadFile(fullPathB)
			if err != nil {
				return fmt.Errorf("error reading %s: %w", fullPathB, err)
			}

			diff, err := differ.Diff(dataA, dataB, opts)
			if err != nil {
				return fmt.Errorf("error diffing %s: %w", filename, err)
			}

			// Requirement: Header should be the filename before diff is displayed
			fmt.Printf("Diff for %s:\n", filename)
			fmt.Println(diff)
			// Add a separator for readability between files
			fmt.Println("--------------------------------------------------")
		}
	}

	return nil
}
