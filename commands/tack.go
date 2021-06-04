package commands

import "github.com/roblillack/tack/core"

func init() {
	RegisterCommand("tack", "Tacks up everything", Tack)
}

func Tack(args ...string) error {
	tacker, err := core.NewTackerWithArgs(args...)
	if err != nil {
		return err
	}

	return tacker.Tack()
}
