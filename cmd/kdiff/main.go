package main

import (
	"fmt"
	"os"
)

func main() {
	if err := Entrypoint().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
