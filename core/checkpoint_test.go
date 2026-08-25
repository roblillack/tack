package core

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

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

func TestCheckpointEquals(t *testing.T) {
	now := time.Now()
	checkpoint := &Checkpoint{files: []fileInfo{
		{Name: "a", ModTime: now},
		{Name: "b", ModTime: now},
	}}

	assert.True(t, checkpoint.Equals(&Checkpoint{files: []fileInfo{
		{Name: "a", ModTime: now},
		{Name: "b", ModTime: now},
	}}))

	// A different number of files ...
	assert.False(t, checkpoint.Equals(&Checkpoint{files: []fileInfo{
		{Name: "a", ModTime: now},
	}}))

	// ... different file names ...
	assert.False(t, checkpoint.Equals(&Checkpoint{files: []fileInfo{
		{Name: "a", ModTime: now},
		{Name: "c", ModTime: now},
	}}))

	// ... and different modification times all mean “not equal”.
	assert.False(t, checkpoint.Equals(&Checkpoint{files: []fileInfo{
		{Name: "a", ModTime: now},
		{Name: "b", ModTime: now.Add(time.Second)},
	}}))
}

func TestCheckpointErrors(t *testing.T) {
	tacker := &Tacker{BaseDir: filepath.Join(t.TempDir(), "nonexistant")}

	_, err := tacker.Checkpoint()
	assert.Error(t, err)

	// Without a previous checkpoint there are always changes, even if creating
	// the new checkpoint failed.
	changes, _, err := tacker.HasChanges(nil)
	assert.True(t, changes)
	assert.Error(t, err)

	changes, _, err = tacker.HasChanges(&Checkpoint{})
	assert.False(t, changes)
	assert.Error(t, err)
}
