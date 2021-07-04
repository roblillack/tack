package commands

import "fmt"

func init() {
	RegisterCommand("help", "Displays this help screen", Help)
}

var Version = "0.0.0-dev"

func Help(args ...string) error {
	fmt.Printf(`tack %s

usage: tack [<verb>] [parameters]

Available verbs:
`, Version)
	for _, i := range List {
		fmt.Printf("    %-15s %s\n", i.Name, i.Description)
	}
	return nil
}
