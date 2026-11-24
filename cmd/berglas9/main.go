// berglas9 — TUI for Berglas secrets in a Cloud Storage bucket.
package main

import (
	"fmt"
	"os"
)

const version = "v0.1.0"

func main() {
	cfg, err := parseLaunch(os.Args[1:])
	if err != nil {
		switch err {
		case errHelp:
			fmt.Println(usageDoc)
			os.Exit(0)
		case errVersion:
			fmt.Println(version)
			os.Exit(0)
		default:
			fmt.Fprintln(os.Stderr, err)
			fmt.Fprintln(os.Stderr, usageDoc)
			os.Exit(1)
		}
	}

	store, err := newBerglasStore(cfg.Bucket, cfg.Key)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}

	if err := runTUI(store, cfg.Bucket); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}
