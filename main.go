package main

import (
	"fmt"
	"os"
)

func Fatalf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "Error: "+format+"\n", args...)
	os.Exit(1)
}

func main() {
	cmd := Tack
	args := []string{}

	if len(os.Args) >= 2 {
		found := false
		for _, i := range commands {
			if os.Args[1] == i.Name {
				cmd = i.Fn
				found = true
				break
			}
		}
		if found {
			args = os.Args[2:]
		} else {
			args = os.Args[1:]
		}
	}

	if err := cmd(args...); err != nil {
		Fatalf(err.Error())
	}
}
