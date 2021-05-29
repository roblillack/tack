package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/roblillack/tack/core"
)

func Fatalf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "Error: "+format+"\n", args...)
	os.Exit(1)
}

func main() {
	dir := ""
	if len(os.Args) == 2 {
		d, err := filepath.Abs(os.Args[1])
		if err != nil {
			Fatalf("Unable to resolve direcotry %s: %s", os.Args[1], err)
		}
		dir = d
	} else {
		cwd, err := os.Getwd()
		if err != nil {
			Fatalf("Unable to determine working dir: %s", err)
		}
		dir = cwd
	}

	tacker, err := core.NewTacker(dir)
	if err != nil {
		Fatalf("Unable to initialize tacker: %w", err)
	}

	if err := tacker.Tack(); err != nil {
		Fatalf("Error tacking: %s", err)
	}
}
