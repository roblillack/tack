package core

import (
	"io/fs"
	"path/filepath"
	"time"
)

type fileInfo struct {
	Name    string
	ModTime time.Time
}

// Checkpoint is a full listing of all files and their respective modification
// timestamps of a specific directory.
type Checkpoint struct {
	files []fileInfo
}

// Equals compares this checkpoint to another one. If an empty Checkpoint to
// compare with is given, or both structures do not share the exact same files,
// or any of the files has a different modification timestamp, they will not
// be regarded as “equal.”
func (c *Checkpoint) Equals(o *Checkpoint) bool {
	if len(c.files) != len(o.files) {
		return false
	}

	for idx, val := range c.files {
		if o.files[idx].Name != val.Name || !o.files[idx].ModTime.Equal(val.ModTime) {
			return false
		}
	}

	return true
}

// Checkpoint stats all the files within the source directory and stores the
// names and modification timestamps so we're able to compare this list with a
// future checkpoint without the need to set up file watchers.
func (t *Tacker) Checkpoint() (*Checkpoint, error) {
	outputDir := filepath.Join(t.BaseDir, TargetDir)
	checkpoint := &Checkpoint{}
	if err := filepath.Walk(t.BaseDir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if path == t.BaseDir {
			return nil
		}
		if path == outputDir {
			return filepath.SkipDir
		}
		checkpoint.files = append(checkpoint.files, fileInfo{
			Name:    path,
			ModTime: info.ModTime(),
		})
		return nil
	}); err != nil {
		return nil, err
	}

	return checkpoint, nil
}

// HasChanges will create a fresh Checkpoint for the current Tacker and
// compare it to the previous one. The functions returns if there are
// changes between the checkpoints (or none was provided in the first)
// place, as well the new Checkpoint, and any error which might have
// occurred while creating it.
func (t *Tacker) HasChanges(prev *Checkpoint) (bool, *Checkpoint, error) {
	if prev == nil {
		c, err := t.Checkpoint()
		return true, c, err
	}

	now, err := t.Checkpoint()
	if err != nil {
		return false, nil, err
	}

	return !now.Equals(prev), now, nil
}
