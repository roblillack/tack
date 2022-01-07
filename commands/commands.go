package commands

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/roblillack/tack/core"
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

func newTackerWithArgs(args ...string) (*core.Tacker, error) {
	if len(args) > 1 {
		return nil, errors.New("too many arguments")
	}

	dir := ""
	if len(args) == 1 {
		d, err := filepath.Abs(args[0])
		if err != nil {
			return nil, fmt.Errorf("unable to resolve directory %s: %s", args[0], err)
		}
		dir = d
	} else {
		cwd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("unable to determine working dir: %s", err)
		}
		dir = cwd
	}

	t, err := core.NewTacker(dir)
	if err != nil {
		return nil, err
	}

	if !DebugMode {
		t.DebugLogger = nil
	}

	return t, nil
}
