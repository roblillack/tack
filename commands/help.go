package commands

import (
	"flag"
	"fmt"
)

func init() {
	RegisterCommand("help", "Displays this help screen", Help)
	flag.Usage = func() { _ = Help() }
}

var Version = "0.0.0-dev"

func Help(args ...string) error {
	fmt.Printf(`tack %s

usage: tack [-d] [<verb>] [parameters]

Available verbs:
`, Version)
	for _, i := range List {
		fmt.Printf("    %-15s %s\n", i.Name, i.Description)
	}

	fmt.Println("\nAvailable global flags:")

	flag.CommandLine.VisitAll(func(fl *flag.Flag) {
		_, usage := flag.UnquoteUsage(fl)
		fmt.Printf("    -%-14s %s\n", fl.Name, usage)
	})

	return nil
}
