package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/roblillack/tack/core"
)

type Executor func(args ...string) error

type Command struct {
	Name string
	Desc string
	Fn   Executor
}

var commands = []Command{
	{Name: "help", Desc: "jojo", Fn: Help},
	{Name: "serve", Desc: "jojo", Fn: Serve},
	{Name: "tack", Desc: "jojo", Fn: Tack},
}

var noCacheHeaders = map[string]string{
	"Expires":       time.Unix(0, 0).Format(time.RFC1123),
	"Cache-Control": "no-cache, no-store, no-transform, must-revalidate, private, max-age=0",
	"Pragma":        "no-cache",
}

func Serve(args ...string) error {
	tacker, err := createTacker(args...)
	if err != nil {
		return err
	}

	var lastBuild time.Time
	var mutex sync.Mutex
	fmt.Println("Listening on :8080...")
	server := http.FileServer(http.Dir(filepath.Join(tacker.BaseDir, core.TargetDir)))
	return http.ListenAndServe(":8080", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		mutex.Lock()
		if lastBuild.Add(5 * time.Second).Before(time.Now()) {
			now := time.Now()
			err := tacker.Reload()
			reloadDuration := time.Since(now)
			var tackDuration = time.Duration(0)
			if err != nil {
				fmt.Println(err)
			} else {
				now := time.Now()
				if err := tacker.Tack(); err != nil {
					fmt.Println(err)
				}
				tackDuration = time.Since(now)
			}
			lastBuild = time.Now()
			fmt.Printf("Rebuilt in %s / %s.\n", reloadDuration, tackDuration)
		}

		for k, v := range noCacheHeaders {
			w.Header().Set(k, v)
		}

		server.ServeHTTP(w, req)
		mutex.Unlock()
	}))
}

func Help(args ...string) error {
	fmt.Println("Help")
	return nil
}

func Tack(args ...string) error {
	tacker, err := createTacker(args...)
	if err != nil {
		return err
	}

	return tacker.Tack()
}

func createTacker(args ...string) (*core.Tacker, error) {
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

	return core.NewTacker(dir)
}
