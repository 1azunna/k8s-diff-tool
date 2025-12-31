package main

import (
	"fmt"
	"os"

	"github.com/1azunna/k8s-diff-tool/cmd/kdiff"
)

func main() {
	if err := kdiff.NewRootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
