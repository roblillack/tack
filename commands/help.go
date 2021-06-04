package commands

import "fmt"

func init() {
	RegisterCommand("help", "Displays this help screen", Help)
}

func Help(args ...string) error {
	fmt.Println(`tack.

usage: tack <verb> [parameters]

Available verbs:`)
	for _, i := range List {
		fmt.Printf("    %-15s %s\n", i.Name, i.Description)
	}
	return nil
}
