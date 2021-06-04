package core

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func FindFiles(dir string, extensions ...string) ([]string, error) {
	result := []string{}
	walk := func(path string, entry fs.DirEntry, err error) error {
		if entry.IsDir() || !entry.Type().IsRegular() {
			return nil
		}
		if len(extensions) > 0 {
			ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(path)), ".")
			valid := false
			for _, i := range extensions {
				if ext == i {
					valid = true
					break
				}
			}
			if !valid {
				return nil
			}
		}
		if err != nil {
			return err
		}
		result = append(result, path)
		return nil
	}

	return result, filepath.WalkDir(dir, walk)
}

func FindDirsWithFiles(dir string, extensions ...string) ([]string, error) {
	result := []string{}
	walk := func(path string, entry fs.DirEntry, err error) error {
		if !entry.IsDir() {
			return nil
		}

		if err != nil {
			return err
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

func LatestModTime(path string, curr *time.Time) error {
	stat, err := os.Stat(path)
	if err != nil {
		return err
	}

	ts := stat.ModTime()
	if ts.After(*curr) {
		*curr = ts
	}

	return nil
}
