package commands

import "flag"

var DebugMode bool

func init() {
	flag.BoolVar(&DebugMode, "d", false, "Print debugging information during site builds")
}
