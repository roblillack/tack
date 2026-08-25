package commands

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/roblillack/tack/core"
	"github.com/stretchr/testify/assert"
)

// writeSite creates a minimal tack-able site in a temporary directory. The
// files given are written in addition to (or on top of) the default ones.
func writeSite(t *testing.T, files map[string]string) string {
	t.Helper()

	base := t.TempDir()
	all := map[string]string{
		"content/default.yaml":       "who: World",
		"templates/default.mustache": "Hello {{who}}!",
	}
	for name, content := range files {
		all[name] = content
	}

	for name, content := range all {
		fn := filepath.Join(base, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(fn), 0755); err != nil {
			t.Fatalf("unable to create directory for %s: %s", name, err)
		}
		if err := os.WriteFile(fn, []byte(content), 0644); err != nil {
			t.Fatalf("unable to write %s: %s", name, err)
		}
	}

	return base
}

// writeBrokenSite creates a site which can be set up, but not tacked up: the
// template is selected by the metadata file's name, and that one is missing.
func writeBrokenSite(t *testing.T) string {
	t.Helper()

	base := writeSite(t, map[string]string{"content/nonexistant.yaml": ""})
	if err := os.Remove(filepath.Join(base, core.ContentDir, "default.yaml")); err != nil {
		t.Fatalf("unable to remove default metadata: %s", err)
	}

	return base
}

// fixSite turns a site created by writeBrokenSite into a working one.
func fixSite(t *testing.T, base string) {
	t.Helper()

	if err := os.Remove(filepath.Join(base, core.ContentDir, "nonexistant.yaml")); err != nil {
		t.Fatalf("unable to remove metadata: %s", err)
	}
	if err := os.WriteFile(filepath.Join(base, core.ContentDir, "default.yaml"), []byte("who: World"), 0644); err != nil {
		t.Fatalf("unable to write metadata: %s", err)
	}
}

// chdir switches the working directory for the duration of the test.
func chdir(t *testing.T, dir string) {
	t.Helper()

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("unable to determine working dir: %s", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("unable to change to %s: %s", dir, err)
	}
	t.Cleanup(func() { _ = os.Chdir(cwd) })
}

// captureStdout collects everything fn prints to stdout.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("unable to create pipe: %s", err)
	}

	orig := os.Stdout
	os.Stdout = w
	defer func() { os.Stdout = orig }()

	fn()
	w.Close()

	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("unable to read captured output: %s", err)
	}

	return string(out)
}

func commandNames() []string {
	names := []string{}
	for _, i := range List {
		names = append(names, i.Name)
	}

	return names
}

func TestRegisteredCommands(t *testing.T) {
	assert.Equal(t, []string{"help", "serve", "tack"}, commandNames())
}

func TestRegisterCommand(t *testing.T) {
	defer func(list []Command) { List = list }(List)

	List = nil
	RegisterCommand("serve", "Runs a server", Serve)
	assert.Equal(t, []string{"serve"}, commandNames())

	// Commands sorting after the first one are appended ...
	RegisterCommand("tack", "Tacks up everything", Tack)
	assert.Equal(t, []string{"serve", "tack"}, commandNames())

	// ... everything else is put in front.
	RegisterCommand("help", "Displays this help screen", Help)
	assert.Equal(t, []string{"help", "serve", "tack"}, commandNames())
	assert.Equal(t, "Displays this help screen", List[0].Description)
}

func TestHelp(t *testing.T) {
	out := captureStdout(t, func() { assert.NoError(t, Help()) })

	assert.Contains(t, out, "tack "+Version)
	assert.Contains(t, out, "usage: tack [<flags>] [<verb>] [parameters]")
	for _, i := range List {
		assert.Contains(t, out, i.Name)
		assert.Contains(t, out, i.Description)
	}
	// Global flags are listed as well.
	assert.Contains(t, out, "-p")
	assert.Contains(t, out, "Port to listen on")
}

func TestNewTackerWithArgs(t *testing.T) {
	defer func(debug, strict bool) { DebugMode, StrictMode = debug, strict }(DebugMode, StrictMode)

	base := writeSite(t, nil)

	tacker, err := newTackerWithArgs(base)
	assert.NoError(t, err)
	assert.Equal(t, base, tacker.BaseDir)
	// Debug logging and strict mode are off unless requested.
	assert.Nil(t, tacker.DebugLogger)
	assert.NotNil(t, tacker.Logger)
	assert.False(t, tacker.Strict)

	DebugMode = true
	StrictMode = true
	tacker, err = newTackerWithArgs(base)
	assert.NoError(t, err)
	assert.NotNil(t, tacker.DebugLogger)
	assert.True(t, tacker.Strict)

	// Only a single directory may be passed ...
	_, err = newTackerWithArgs(base, base)
	assert.Error(t, err)

	// ... and it needs to contain a site.
	_, err = newTackerWithArgs(filepath.Join(base, "nonexistant"))
	assert.Error(t, err)
}

func TestNewTackerWithoutArgs(t *testing.T) {
	base := writeSite(t, nil)
	chdir(t, base)

	// Without arguments the working directory is used.
	tacker, err := newTackerWithArgs()
	assert.NoError(t, err)
	assert.NotNil(t, tacker)
	assert.NoError(t, tacker.Tack())
	assert.FileExists(t, filepath.Join(base, core.TargetDir, "index.html"))
}

func TestTack(t *testing.T) {
	base := writeSite(t, nil)

	assert.NoError(t, Tack(base))
	content, err := os.ReadFile(filepath.Join(base, core.TargetDir, "index.html"))
	assert.NoError(t, err)
	assert.Equal(t, "Hello World!", string(content))

	// Errors setting up the site are passed on ...
	assert.Error(t, Tack(base, base))
	// ... as are errors while tacking it up.
	assert.Error(t, Tack(writeBrokenSite(t)))
}

func TestServeArgumentErrors(t *testing.T) {
	defer func(port int) { Port = port }(Port)

	base := writeSite(t, nil)

	assert.Error(t, Serve(base, base))

	Port = 65536
	assert.Error(t, Serve(base))
	Port = -1
	assert.Error(t, Serve(base))
}

func TestServePortInUse(t *testing.T) {
	defer func(port int) { Port = port }(Port)

	// Bind to the same wildcard address the server uses, so that it really
	// cannot get hold of the port.
	listener, err := net.Listen("tcp", ":0")
	assert.NoError(t, err)
	defer listener.Close()

	Port = listener.Addr().(*net.TCPAddr).Port
	assert.Error(t, Serve(writeSite(t, nil)))
}

func TestServeError(t *testing.T) {
	w := httptest.NewRecorder()
	ServeError(w, httptest.NewRequest("GET", "/some/page", nil), fmt.Errorf("nope"))

	assert.Equal(t, 500, w.Code)
	assert.Equal(t, "Error: nope\n", w.Body.String())
}

func TestServe(t *testing.T) {
	defer func(port int) { Port = port }(Port)

	// The site does not tack up cleanly at first – the server should come up
	// nevertheless.
	base := writeBrokenSite(t)
	Port = freePort(t)
	go func() { _ = Serve(base) }()

	// Tacking the site fails, so the error is reported to the client.
	status, body := get(t, Port)
	assert.Equal(t, 500, status)
	assert.Contains(t, body, "Error:")

	// Changes are picked up without restarting the server.
	fixSite(t, base)
	awaitRebuild(t)
	status, body = get(t, Port)
	assert.Equal(t, 200, status)
	assert.Equal(t, "Hello World!", body)

	// Responses are marked as not cacheable.
	resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/", Port))
	assert.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, "no-cache", resp.Header.Get("Pragma"))
}

func TestServeReportsBrokenSites(t *testing.T) {
	defer func(port int) { Port = port }(Port)

	base := writeSite(t, nil)
	Port = freePort(t)
	go func() { _ = Serve(base) }()

	status, body := get(t, Port)
	assert.Equal(t, 200, status)
	assert.Equal(t, "Hello World!", body)

	// Breaking the site metadata makes reloading it fail.
	assert.NoError(t, os.WriteFile(filepath.Join(base, "site.yaml"), []byte("\ttitle: ["), 0644))
	awaitRebuild(t)
	status, body = get(t, Port)
	assert.Equal(t, 500, status)
	assert.Contains(t, body, "Error:")
}

// freePort returns a port number nothing is listening on right now.
func freePort(t *testing.T) int {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("unable to find a free port: %s", err)
	}
	defer listener.Close()

	return listener.Addr().(*net.TCPAddr).Port
}

// awaitRebuild waits until the server checks for changes again.
func awaitRebuild(t *testing.T) {
	t.Helper()

	time.Sleep(3100 * time.Millisecond)
}

// get requests the site's index page, waiting for the server to come up.
func get(t *testing.T, port int) (int, string) {
	t.Helper()

	var lastErr error
	for i := 0; i < 100; i++ {
		resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/", port))
		if err != nil {
			lastErr = err
			time.Sleep(20 * time.Millisecond)
			continue
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("unable to read response: %s", err)
		}

		return resp.StatusCode, string(body)
	}

	t.Fatalf("server did not come up: %s", lastErr)
	return 0, ""
}
