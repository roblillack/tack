package core

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// FindFiles returns a slice containing the absolute path of _all_ regular files below
// the given directory.
func FindFiles(dir string) ([]string, error) {
	result := []string{}
	walk := func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || !entry.Type().IsRegular() {
			return nil
		}

		result = append(result, path)
		return nil
	}

	return result, filepath.WalkDir(dir, walk)
}

func FindDirsWithFiles(dir string, extensions ...string) ([]string, error) {
	result := []string{}
	walk := func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !entry.IsDir() {
			return nil
		}

		if len(extensions) == 0 {
			result = append(result, path)
			return nil
		}

		for _, ext := range extensions {
			m, err := filepath.Glob(filepath.Join(path, "*."+ext))
			if err != nil {
				return err
			}
			if len(m) > 0 {
				result = append(result, path)
				return nil
			}
		}

		return nil
	}

	return result, filepath.WalkDir(dir, walk)
}

func CopyFile(src, dst string) error {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()

	_, err = io.Copy(destination, source)
	return err
}

func BasenameWithoutExtension(path string) string {
	b := filepath.Base(path)
	return strings.TrimSuffix(b, filepath.Ext(b))
}

func DirExists(path string) bool {
	s, err := os.Stat(path)
	return err == nil && s.IsDir()
}
