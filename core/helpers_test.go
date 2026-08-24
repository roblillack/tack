package core

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

// writeFiles creates a fresh temporary directory containing all the files
// given. The map keys are slash-separated paths relative to that directory,
// intermediate directories are created as needed.
func writeFiles(t *testing.T, files map[string]string) string {
	t.Helper()

	base := t.TempDir()
	for name, content := range files {
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

// tackSite writes a site to a temporary directory, tacks it up, and returns
// the Tacker used as well as the site's base directory.
func tackSite(t *testing.T, files map[string]string) (*Tacker, string) {
	t.Helper()

	base := writeFiles(t, files)
	tacker, err := NewTacker(base)
	if err != nil {
		t.Fatalf("unable to set up site: %s", err)
	}
	tacker.Logger = nil
	tacker.DebugLogger = nil

	if err := tacker.Tack(); err != nil {
		t.Fatalf("unable to tack site: %s", err)
	}

	return tacker, base
}

// generated returns the contents of a file below the site's output directory.
func generated(t *testing.T, base string, path ...string) string {
	t.Helper()

	content, err := os.ReadFile(filepath.Join(append([]string{base, TargetDir}, path...)...))
	assert.NoError(t, err)

	return string(content)
}

// skipAsRoot skips tests which rely on file permissions being enforced.
func skipAsRoot(t *testing.T) {
	t.Helper()

	if os.Geteuid() == 0 {
		t.Skip("permissions are not enforced when running as root")
	}
}
