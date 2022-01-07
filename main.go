package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/roblillack/tack/commands"
	"github.com/roblillack/tack/core"
)

func Fatalf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "Error: "+format+"\n", args...)
	os.Exit(1)
}

func main() {
	cmd := commands.Tack
	args := []string{}

	flag.Parse()

	if len(flag.Args()) >= 1 {
		found := false
		for _, i := range commands.List {
			if flag.Arg(0) == i.Name {
				cmd = i.Fn
				found = true
				break
			}
		}
		if found {
			args = flag.Args()[1:]
		} else if core.DirExists(flag.Arg(0)) {
			args = flag.Args()
		} else {
			fmt.Fprintf(os.Stderr, "Not a known verb or site directory: %s\n\n", flag.Arg(0))
			_ = commands.Help()
			os.Exit(1)
		}
	}

	if err := cmd(args...); err != nil {
		Fatalf(err.Error())
	}
}
