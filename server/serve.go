package server

import (
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"sync"
	"time"

	"github.com/roblillack/tack/core"
)

var noCacheHeaders = map[string]string{
	"Cache-Control": "no-cache, no-store, no-transform, must-revalidate, private, max-age=0",
	"Expires":       time.Unix(0, 0).Format(time.RFC1123),
	"Pragma":        "no-cache",
}

func ServeError(w http.ResponseWriter, err error) {
	w.WriteHeader(500)
	w.Write([]byte(fmt.Sprintf("Internal Server Error: %s\n", err.Error())))
}

func Serve(args ...string) error {
	tacker, err := core.NewTackerWithArgs(args...)
	if err != nil {
		return err
	}
	tacker.Logger = nil

	var lastBuild time.Time
	var checkpoint *core.Checkpoint
	var mutex sync.Mutex
	log.Println("Listening on :8080...")
	server := http.FileServer(http.Dir(filepath.Join(tacker.BaseDir, core.TargetDir)))
	return http.ListenAndServe(":8080", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		mutex.Lock()
		defer mutex.Unlock()
		log.Println(req.RequestURI)
		if time.Since(lastBuild) >= 3*time.Second {
			rebuild, newCheckpoint, err := tacker.HasChanges(checkpoint)
			if err != nil {
				log.Println(err)
				ServeError(w, err)
				return
			}
			if rebuild {
				tackStart := time.Now()
				if err := tacker.Reload(); err != nil {
					log.Println(err)
					ServeError(w, err)
					return
				}
				if err := tacker.Tack(); err != nil {
					log.Println(err)
					ServeError(w, err)
					return
				}
				lastBuild = time.Now()
				checkpoint = newCheckpoint
				log.Printf("Rebuilt in %s.\n", time.Since(tackStart))
			}
		}

		for k, v := range noCacheHeaders {
			w.Header().Set(k, v)
		}

		server.ServeHTTP(w, req)
	}))
}
