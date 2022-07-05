package commands

import "flag"

var DebugMode bool
var StrictMode bool

func init() {
	flag.BoolVar(&DebugMode, "d", false, "Print debugging information during site builds")
	flag.BoolVar(&StrictMode, "s", false, "Enable strict mode (fails when trying to render undefined variables)")
}
