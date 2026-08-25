package core

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateInitializesPage(t *testing.T) {
	base := writeFiles(t, map[string]string{
		"content/default.yaml":       "who: World",
		"templates/default.mustache": "Hello {{who}}!",
	})

	// Pages which have not been Init()ed yet are initialized on the fly.
	page := NewPage(&Tacker{BaseDir: base}, filepath.Join(base, ContentDir))
	assert.False(t, page.inited)
	assert.NoError(t, page.Generate())
	assert.True(t, page.inited)
	assert.Equal(t, "Hello World!", generated(t, base, "index.html"))
}

func TestInitOfNonexistantPage(t *testing.T) {
	tacker, base := tackSite(t, map[string]string{
		"content/default.yaml":       "",
		"templates/default.mustache": "Hi",
	})

	page := NewPage(tacker, filepath.Join(base, ContentDir, "nonexistant"))
	assert.Error(t, page.Init())
	assert.Error(t, page.Generate())
}

func TestPageWithInvalidMetadata(t *testing.T) {
	base := writeFiles(t, map[string]string{
		"content/default.yaml":       "\ttitle: [",
		"templates/default.mustache": "Hi",
	})

	_, err := NewTacker(base)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unable to process metadata")
}

func TestPageWithConflictingTemplates(t *testing.T) {
	// The metadata file's basename selects the template, the front matter of
	// the markup file asks for a different one.
	base := writeFiles(t, map[string]string{
		"content/a.yaml":             "",
		"content/b.md":               "---\ntemplate: c\n---\n\n# Hi",
		"templates/default.mustache": "Hi",
	})

	_, err := NewTacker(base)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "multiple templates requested")
}

func TestPageWithUnreadableMarkup(t *testing.T) {
	skipAsRoot(t)

	base := writeFiles(t, map[string]string{
		"content/body.md":            "# Hi",
		"templates/default.mustache": "{{{body}}}",
	})
	assert.NoError(t, os.Chmod(filepath.Join(base, ContentDir, "body.md"), 0200))

	_, err := NewTacker(base)
	assert.Error(t, err)
}

func TestGenerateWithMissingTemplate(t *testing.T) {
	base := writeFiles(t, map[string]string{
		"content/nonexistant.yaml":   "",
		"templates/default.mustache": "Hi",
	})
	tacker, err := NewTacker(base)
	assert.NoError(t, err)
	tacker.Logger = nil
	tacker.DebugLogger = nil

	err = tacker.Tack()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unable to load template 'nonexistant'")
}

func TestGenerateIntoReadOnlyDir(t *testing.T) {
	skipAsRoot(t)

	tacker, base := tackSite(t, map[string]string{
		"content/default.yaml":       "",
		"templates/default.mustache": "Hi",
	})

	out := filepath.Join(base, TargetDir)
	assert.NoError(t, os.Remove(filepath.Join(out, "index.html")))
	assert.NoError(t, os.Chmod(out, 0555))
	defer func() { assert.NoError(t, os.Chmod(out, 0755)) }()

	assert.Len(t, tacker.Pages, 1)
	assert.Error(t, tacker.Pages[0].Generate())
}

func TestGenerateWithUnreadableAsset(t *testing.T) {
	skipAsRoot(t)

	base := writeFiles(t, map[string]string{
		"content/default.yaml":       "",
		"content/asset.txt":          "Hello World!",
		"templates/default.mustache": "Hi",
	})
	tacker, err := NewTacker(base)
	assert.NoError(t, err)
	tacker.Logger = nil
	tacker.DebugLogger = nil

	assert.NoError(t, os.Chmod(filepath.Join(base, ContentDir, "asset.txt"), 0000))
	defer func() {
		assert.NoError(t, os.Chmod(filepath.Join(base, ContentDir, "asset.txt"), 0644))
	}()

	assert.Error(t, tacker.Tack())
}

func TestTitleWords(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		// Ordinary slugs: first letter of each word, hyphens left in place.
		{"hello-world", "Hello-World"},
		{"index", "Index"},
		{"a-b-c", "A-B-C"},
		{"hello world", "Hello World"},

		// The remainder of a word is never touched.
		{"hELLO-wORLD", "HELLO-WORLD"},
		{"MiXeD-CaSe", "MiXeD-CaSe"},
		{"HTML-and-CSS", "HTML-And-CSS"},

		// Digits are word-internal, so "2nd" does not become "2Nd".
		{"2nd-place", "2nd-Place"},
		{"3rd-time", "3rd-Time"},
		{"4k-video", "4k-Video"},
		{"post-2024-01-01", "Post-2024-01-01"},
		{"x1-y2", "X1-Y2"},

		// Apostrophes are word-internal too -- strings.Title got these wrong.
		{"it's-a-test", "It's-A-Test"},
		{"don't-panic", "Don't-Panic"},
		{"rock'n'roll", "Rock'n'roll"},
		{"don’t-worry", "Don’t-Worry"},

		// Non-ASCII letters, including a title-case digraph.
		{"über-uns", "Über-Uns"},
		{"café-au-lait", "Café-Au-Lait"},
		{"ǳ-digraph", "ǲ-Digraph"},

		// Non-ASCII punctuation and symbols separate words too, which
		// strings.Title did not do.
		{"a–b", "A–B"},
		{"rock—and—roll", "Rock—And—Roll"},
		{"foo·bar", "Foo·Bar"},
		{"«zitat»", "«Zitat»"},
		{"a\u00a0b", "A\u00a0B"},
		{"pi≈three", "Pi≈Three"},

		// Edge cases.
		{"", ""},
		{"foo--bar", "Foo--Bar"},
		{"-leading", "-Leading"},
		{"trailing-", "Trailing-"},
		{"_under_score", "_Under_Score"},
		{"foo_bar", "Foo_Bar"},
		{"  spaced  ", "  Spaced  "},
	}

	for _, tt := range tests {
		if got := titleWords(tt.in); got != tt.want {
			t.Errorf("titleWords(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}
