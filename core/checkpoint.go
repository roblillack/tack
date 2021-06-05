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

type Checkpoint struct {
	files []fileInfo
}

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
