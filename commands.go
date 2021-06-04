package main

import (
	"fmt"

	"github.com/roblillack/tack/core"
	"github.com/roblillack/tack/server"
)

type Executor func(args ...string) error

type Command struct {
	Name string
	Desc string
	Fn   Executor
}

var commands = []Command{
	{Name: "help", Desc: "jojo", Fn: Help},
	{Name: "serve", Desc: "jojo", Fn: server.Serve},
	{Name: "tack", Desc: "jojo", Fn: Tack},
}

func Help(args ...string) error {
	fmt.Println("Help")
	return nil
}

func Tack(args ...string) error {
	tacker, err := core.NewTackerWithArgs(args...)
	if err != nil {
		return err
	}

	return tacker.Tack()
}
