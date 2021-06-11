package core

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckpoints(t *testing.T) {
	_, filename, _, _ := runtime.Caller(0)

	testSites, err := filepath.Glob(filepath.Join(filepath.Dir(filename), "tests", "*"))
	if err != nil {
		t.Fatal("unable to read test sites")
	}

	for _, site := range testSites {
		if strings.HasPrefix(filepath.Base(site), ".") {
			continue
		}
		tacker, err := NewTacker(site)
		assert.NoError(t, err)
		changes, checkpoint, err := tacker.HasChanges(nil)
		assert.NoError(t, err)
		assert.True(t, changes)
		changes, checkpoint, err = tacker.HasChanges(checkpoint)
		assert.NoError(t, err)
		assert.False(t, changes)
		assert.NoError(t, tacker.Tack())
		changes, checkpoint, err = tacker.HasChanges(checkpoint)
		assert.NoError(t, err)
		assert.False(t, changes)
		assert.NoError(t, os.WriteFile(filepath.Join(site, "temp.yaml"), []byte{}, 0644))
		changes, _, err = tacker.HasChanges(checkpoint)
		assert.NoError(t, err)
		assert.True(t, changes)
		assert.NoError(t, os.Remove(filepath.Join(site, "temp.yaml")))
	}
}
