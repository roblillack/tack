package commands

import (
	"fmt"

	"github.com/roblillack/tack/core"
)

func init() {
	RegisterCommand("list", "Lists all pages of the site", ListCmd)
}

func ListCmd(args ...string) error {
	tacker, err := newTackerWithArgs(args...)
	if err != nil {
		return err
	}

	listChildren(tacker, nil, 0)

	return nil
}

func listChildren(t *core.Tacker, parent *core.Page, level int) {
	for _, p := range t.Pages {
		if p.Parent != parent {
			continue
		}
		for i := 0; i < level; i++ {
			fmt.Printf("  ")
		}
		fmt.Printf("- %s (%s --> %s)\n", p.Name, p.DiskPath, p.Permalink())
		listChildren(t, p, level+1)
	}
}
