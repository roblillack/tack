package commands

import "flag"

var DebugMode bool
var StrictMode bool
var Port int

func init() {
	flag.BoolVar(&DebugMode, "d", false, "Print debugging information during site builds")
	flag.BoolVar(&StrictMode, "s", false, "Enable strict mode (fails when trying to render undefined variables)")
	flag.IntVar(&Port, "p", 8080, "Port to listen on when serving the site (0 picks a free one)")
}
