package core

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFindFiles(t *testing.T) {
	base := writeFiles(t, map[string]string{
		"a.txt":       "",
		"sub/b.txt":   "",
		"sub/c/d.txt": "",
	})

	files, err := FindFiles(base)
	assert.NoError(t, err)
	assert.Equal(t, []string{
		filepath.Join(base, "a.txt"),
		filepath.Join(base, "sub", "b.txt"),
		filepath.Join(base, "sub", "c", "d.txt"),
	}, files)

	_, err = FindFiles(filepath.Join(base, "nonexistant"))
	assert.Error(t, err)
}

func TestFindDirsWithFiles(t *testing.T) {
	base := writeFiles(t, map[string]string{
		"markup/a.md":    "",
		"metadata/b.yml": "",
		"assets/c.png":   "",
	})

	// Without any extension given, all directories are returned.
	dirs, err := FindDirsWithFiles(base)
	assert.NoError(t, err)
	assert.Equal(t, []string{
		base,
		filepath.Join(base, "assets"),
		filepath.Join(base, "markup"),
		filepath.Join(base, "metadata"),
	}, dirs)

	dirs, err = FindDirsWithFiles(base, "md", "yml")
	assert.NoError(t, err)
	assert.Equal(t, []string{
		filepath.Join(base, "markup"),
		filepath.Join(base, "metadata"),
	}, dirs)

	_, err = FindDirsWithFiles(filepath.Join(base, "nonexistant"), "md")
	assert.Error(t, err)
}

func TestFindDirsWithFilesInvalidPattern(t *testing.T) {
	// A directory name containing an unterminated character class cannot be
	// turned into a valid glob pattern.
	base := writeFiles(t, map[string]string{"oh[dear/a.md": ""})

	_, err := FindDirsWithFiles(base, "md")
	assert.Equal(t, filepath.ErrBadPattern, err)
}

func TestCopyFile(t *testing.T) {
	base := writeFiles(t, map[string]string{"source.txt": "Hello World!"})
	src := filepath.Join(base, "source.txt")

	assert.NoError(t, CopyFile(src, filepath.Join(base, "target.txt")))
	content, err := os.ReadFile(filepath.Join(base, "target.txt"))
	assert.NoError(t, err)
	assert.Equal(t, "Hello World!", string(content))

	// The source needs to exist ...
	assert.Error(t, CopyFile(filepath.Join(base, "nonexistant"), filepath.Join(base, "x.txt")))
	// ... and be a regular file.
	assert.Error(t, CopyFile(base, filepath.Join(base, "x.txt")))
	// The target's directory needs to exist, too.
	assert.Error(t, CopyFile(src, filepath.Join(base, "nonexistant", "x.txt")))
}

func TestCopyUnreadableFile(t *testing.T) {
	skipAsRoot(t)

	base := writeFiles(t, map[string]string{"source.txt": "Hello World!"})
	src := filepath.Join(base, "source.txt")
	assert.NoError(t, os.Chmod(src, 0200))

	assert.Error(t, CopyFile(src, filepath.Join(base, "target.txt")))
}

func TestBasenameWithoutExtension(t *testing.T) {
	assert.Equal(t, "index", BasenameWithoutExtension("/a/b/index.html"))
	assert.Equal(t, "default", BasenameWithoutExtension("default.yaml"))
	assert.Equal(t, "no-extension", BasenameWithoutExtension("/a/no-extension"))
	assert.Equal(t, "two.dots", BasenameWithoutExtension("two.dots.md"))
}

func TestDirExists(t *testing.T) {
	base := writeFiles(t, map[string]string{"sub/file.txt": ""})

	assert.True(t, DirExists(base))
	assert.True(t, DirExists(filepath.Join(base, "sub")))
	assert.False(t, DirExists(filepath.Join(base, "sub", "file.txt")))
	assert.False(t, DirExists(filepath.Join(base, "nonexistant")))
}

func TestFirstFileWithExtension(t *testing.T) {
	base := writeFiles(t, map[string]string{
		"page.mu":        "",
		"page.mustache":  "",
		"plain":          "",
		"directory/x.md": "",
	})

	// Extensions are tried in the order given, not in the order found on disk.
	assert.Equal(t, filepath.Join(base, "page.mustache"),
		FirstFileWithExtension(base, "page", "mustache", "mu"))
	assert.Equal(t, filepath.Join(base, "page.mu"),
		FirstFileWithExtension(base, "page", "mu", "mustache"))

	// Without any extension given, the basename is used as-is.
	assert.Equal(t, filepath.Join(base, "plain"), FirstFileWithExtension(base, "plain"))

	// Directories and missing files do not count.
	assert.Equal(t, "", FirstFileWithExtension(base, "directory"))
	assert.Equal(t, "", FirstFileWithExtension(base, "page", "stache"))
	assert.Equal(t, "", FirstFileWithExtension(base, "nonexistant"))
}
