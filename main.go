package main

import (
	"fmt"
	"os"

	"github.com/roblillack/tack/core"
)

func Fatalf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "Error: "+format+"\n", args...)
	os.Exit(1)
}

func main() {
	fmt.Println("Hello World!")

	cwd, err := os.Getwd()
	if err != nil {
		Fatalf("Unable to determine working dir: %w", err)
	}

	tacker, err := core.NewTacker(cwd)
	if err != nil {
		Fatalf("Unable to initialize tacker: %w", err)
	}

	if err := tacker.Tack(); err != nil {
		Fatalf("Error tacking: %w", err)
	}
}
