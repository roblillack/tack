package commands

import (
	"strings"
)

type Executor func(args ...string) error

type Command struct {
	Name        string
	Description string
	Fn          Executor
}

var List []Command

func RegisterCommand(name string, desc string, fn Executor) {
	cmd := Command{Name: name, Description: desc, Fn: fn}

	if len(List) == 0 {
		List = []Command{cmd}
		return
	}

	if strings.Compare(List[0].Name, name) < 0 {
		List = append(List, cmd)
	} else {
		List = append([]Command{cmd}, List...)

	}
}
