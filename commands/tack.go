package commands

func init() {
	RegisterCommand("tack", "Tacks up everything", Tack)
}

func Tack(args ...string) error {
	tacker, err := newTackerWithArgs(args...)
	if err != nil {
		return err
	}

	return tacker.Tack()
}
