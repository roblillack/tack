package commands

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

func init() {
	RegisterCommand("serve", "Runs a minimal HTTP server", Serve)
}

func ServeError(w http.ResponseWriter, err error) {
	w.WriteHeader(500)
	w.Write([]byte(fmt.Sprintf("Internal Server Error: %s\n", err.Error())))
}

func Serve(args ...string) error {
	tacker, err := newTackerWithArgs(args...)
	if err != nil {
		return err
	}
	if err := tacker.Tack(); err != nil {
		log.Println(err)
	}
	tacker.Logger = nil

	var lastBuild time.Time
	var checkpoint *core.Checkpoint
	var mutex sync.Mutex

	htmlDir := filepath.Join(tacker.BaseDir, core.TargetDir)
	log.Printf("Serving from %s, listening on port 8080 …\n", htmlDir)
	server := http.FileServer(http.Dir(htmlDir))
	return http.ListenAndServe(":8080", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		mutex.Lock()
		defer mutex.Unlock()
		log.Printf("%s %s://%s%s%s\n", req.Method, "http", req.Host, req.URL.Port(), req.RequestURI)
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
				if !lastBuild.IsZero() {
					log.Printf("Changes detected. Re-tacked in %s.\n", time.Since(tackStart))
				}
				lastBuild = time.Now()
				checkpoint = newCheckpoint
			}
		}

		for k, v := range noCacheHeaders {
			w.Header().Set(k, v)
		}

		server.ServeHTTP(w, req)
	}))
}
