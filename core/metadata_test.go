package core

import (
	"bytes"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProcessMetadata(t *testing.T) {
	base := writeFiles(t, map[string]string{
		"empty.yaml":  "",
		"page.yaml":   "title: Hello\n:template: special\n",
		"invalid.yml": "\ttitle: [",
	})

	md, err := ProcessMetadata(filepath.Join(base, "empty.yaml"))
	assert.NoError(t, err)
	assert.Empty(t, md)

	// Keys may be prefixed with a colon to keep them from being interpreted as
	// YAML tags by other tools. The prefix is stripped when reading them.
	md, err = ProcessMetadata(filepath.Join(base, "page.yaml"))
	assert.NoError(t, err)
	assert.Equal(t, map[string]interface{}{"title": "Hello", "template": "special"}, md)

	_, err = ProcessMetadata(filepath.Join(base, "invalid.yml"))
	assert.Error(t, err)

	_, err = ProcessMetadata(filepath.Join(base, "nonexistant.yaml"))
	assert.Error(t, err)
}

func TestSiteMetadata(t *testing.T) {
	// All YAML files in the site's base directory are read in alphabetical
	// order, later ones overriding what earlier ones defined.
	_, base := tackSite(t, map[string]string{
		"a-site.yaml":                "author: Rob\ntitle: First\n",
		"b-more.yml":                 "title: Second\n",
		"not-metadata.txt":           "ignored",
		"content/default.yaml":       "",
		"templates/default.mustache": "{{author}}/{{title}}",
	})

	assert.Equal(t, "Rob/Second", generated(t, base, "index.html"))
}

func TestInvalidSiteMetadata(t *testing.T) {
	base := writeFiles(t, map[string]string{
		"site.yaml":                  "\ttitle: [",
		"content/default.yaml":       "",
		"templates/default.mustache": "Hi",
	})

	_, err := NewTacker(base)
	assert.Error(t, err)
}

func TestNewTackerNeedsSiteStructure(t *testing.T) {
	base := t.TempDir()

	_, err := NewTacker(base)
	assert.Error(t, err)

	// Only having a content directory is not enough ...
	assert.NoError(t, os.MkdirAll(filepath.Join(base, ContentDir), 0755))
	_, err = NewTacker(base)
	assert.Error(t, err)

	// ... templates are needed as well.
	assert.NoError(t, os.MkdirAll(filepath.Join(base, TemplateDir), 0755))
	_, err = NewTacker(base)
	assert.NoError(t, err)
}

func TestFindTemplate(t *testing.T) {
	tacker, _ := tackSite(t, map[string]string{
		"content/default.yaml":       "",
		"templates/default.mustache": "Hello {{who}}!",
		"templates/partial.mu":       "{{> default}}",
		"templates/invalid.stache":   "{{#unclosed}}",
	})

	// Not specifying a template name gets you the default one.
	tpl, err := tacker.FindTemplate("")
	assert.NoError(t, err)
	assert.NotNil(t, tpl)

	// All of the known template extensions are picked up, also for partials.
	tpl, err = tacker.FindTemplate("partial")
	assert.NoError(t, err)
	assert.NotNil(t, tpl)

	_, err = tacker.FindTemplate("nonexistant")
	assert.Error(t, err)

	_, err = tacker.FindTemplate("invalid")
	assert.Error(t, err)
}

func TestLogging(t *testing.T) {
	buf := &bytes.Buffer{}
	tacker := &Tacker{
		Logger:      log.New(buf, "", 0),
		DebugLogger: log.New(buf, "", 0),
	}

	tacker.Log("hello %s", "world")
	tacker.Debug("debug %d", 42)
	assert.Equal(t, "hello world\ndebug 42\n", buf.String())

	// Unsetting a logger silences the respective output.
	tacker.Logger = nil
	tacker.DebugLogger = nil
	tacker.Log("nope")
	tacker.Debug("nope")
	assert.Equal(t, "hello world\ndebug 42\n", buf.String())
}

func TestInvalidContentDirectoryName(t *testing.T) {
	// A directory name containing an unterminated character class cannot be
	// searched for content files.
	base := writeFiles(t, map[string]string{
		"content/oh[dear/body.md":    "",
		"templates/default.mustache": "Hi",
	})

	_, err := NewTacker(base)
	assert.Error(t, err)
}

func TestTackWithUnstatableOutputDir(t *testing.T) {
	base := writeFiles(t, map[string]string{
		"content/default.yaml":       "",
		"templates/default.mustache": "Hi",
	})
	tacker, err := NewTacker(base)
	assert.NoError(t, err)
	tacker.Logger = nil
	tacker.DebugLogger = nil

	// A symlink pointing to itself makes stat() fail with something else than
	// “does not exist”.
	out := filepath.Join(base, TargetDir)
	assert.NoError(t, os.Symlink(out, out))

	assert.Error(t, tacker.Tack())
}

func TestTackWithUndeletableOutputDir(t *testing.T) {
	skipAsRoot(t)

	base := writeFiles(t, map[string]string{
		"content/default.yaml":       "",
		"templates/default.mustache": "Hi",
		"output/index.html":          "stale",
	})
	tacker, err := NewTacker(base)
	assert.NoError(t, err)
	tacker.Logger = nil
	tacker.DebugLogger = nil

	assert.NoError(t, os.Chmod(base, 0555))
	defer func() { assert.NoError(t, os.Chmod(base, 0755)) }()

	assert.Error(t, tacker.Tack())
}

func TestTackWithUncreatableOutputDir(t *testing.T) {
	skipAsRoot(t)

	base := writeFiles(t, map[string]string{
		"content/default.yaml":       "",
		"templates/default.mustache": "Hi",
	})
	tacker, err := NewTacker(base)
	assert.NoError(t, err)
	tacker.Logger = nil
	tacker.DebugLogger = nil

	assert.NoError(t, os.Chmod(base, 0555))
	defer func() { assert.NoError(t, os.Chmod(base, 0755)) }()

	assert.Error(t, tacker.Tack())
}

func TestTackWithUnreadableAssetDir(t *testing.T) {
	skipAsRoot(t)

	base := writeFiles(t, map[string]string{
		"content/default.yaml":       "",
		"templates/default.mustache": "Hi",
		"public/asset.txt":           "",
	})
	tacker, err := NewTacker(base)
	assert.NoError(t, err)
	tacker.Logger = nil
	tacker.DebugLogger = nil

	assets := filepath.Join(base, AssetDir)
	assert.NoError(t, os.Chmod(assets, 0000))
	defer func() { assert.NoError(t, os.Chmod(assets, 0755)) }()

	assert.Error(t, tacker.Tack())
}
