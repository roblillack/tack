module github.com/roblillack/tack

go 1.16

require (
	github.com/cbroglie/mustache v1.2.2
	github.com/fsnotify/fsnotify v1.4.9 // indirect
	github.com/jandelgado/gcov2lcov v1.0.5 // indirect
	github.com/stretchr/testify v1.7.0
	github.com/yuin/goldmark v1.3.7
	gopkg.in/yaml.v2 v2.4.0
)

// replace github.com/cbroglie/mustache => ../mustache
// replace github.com/roblillack/mustache => ../mustache
