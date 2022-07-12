package core

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTacker(t *testing.T) {
	_, filename, _, _ := runtime.Caller(0)

	testSites, err := filepath.Glob(filepath.Join(filepath.Dir(filename), "tests", "*"))
	if err != nil {
		t.Fatal("unable to read test sites")
	}

	for _, site := range testSites {
		if strings.HasPrefix(filepath.Base(site), ".") {
			continue
		}
		if filepath.Base(site) != "minimal-blog-with-tags-below-index" {
			continue
		}
		passStrict := map[string]struct{}{
			"helloworld":                                 {},
			"helloworld-index-not-in-root":               {},
			"minimal":                                    {},
			"minimal-with-nav":                           {},
			"test-copying-assets":                        {},
			"test-different-file-extensions":             {},
			"test-page-variable-overrides-site-metadata": {},
			"test-page-variable-overrides-template":      {},
		}
		for _, strictMode := range []bool{false, true} {
			tacker, err := NewTacker(site)
			tacker.Strict = strictMode
			assert.NoError(t, err)

			err = tacker.Tack()
			if !strictMode && err != nil {
				t.Fatalf("Unable to tack site %s: %s", filepath.Base(site), err)
			}

			_, shouldPass := passStrict[filepath.Base(site)]
			if strictMode && shouldPass && err != nil {
				t.Fatalf("Site %s should be tackable in strict mode but is not: %s", filepath.Base(site), err)
			} else if strictMode && err == nil && !shouldPass {
				t.Fatalf("Site %s should not be tackable in strict mode but is!", filepath.Base(site))
			} else if strictMode && err != nil && !shouldPass {
				t.Logf("Got error tacking %s in strict mode as expected", filepath.Base(site))
				continue
			}

			AssertDirEquals(t, filepath.Join(site, "output.expected"), filepath.Join(site, "output"))
		}
	}
}

type LayoutTest struct {
	Paths      []string
	Permalinks []string
}

var layoutTests = []LayoutTest{
	{
		Paths: []string{
			"index/content.yml",
			"page-a/bla.md",
			"page-b/blub.yml",
		},
		Permalinks: []string{
			"/",
			"/page-a",
			"/page-b",
		},
	},
	{
		Paths: []string{
			"content.yml",
			"page-a/bla.md",
			"page-b/blub.yml",
		},
		Permalinks: []string{
			"/",
			"/page-a",
			"/page-b",
		},
	},
	{
		Paths: []string{
			"page-a/bla.md",
			"page-b/blub.yml",
		},
		Permalinks: []string{
			"/",
			"/page-a",
			"/page-b",
		},
	},
	{
		Paths: []string{
			"2.a/x.md",
			"1.b/x.yaml",
		},
		Permalinks: []string{"/", "/a", "/b"},
	},
	{
		Paths: []string{
			"0.index/x.mkd",
			"1.b/x.yaml",
			"2.a/x.md",
		},
		Permalinks: []string{"/", "/a", "/b"},
	},
	{
		Paths: []string{
			"1981-10-08-a/x.md",
		},
		Permalinks: []string{
			"/",
			"/a",
		},
	},
	{
		Paths: []string{
			"posts/1981-10-08-a/x.md",
		},
		Permalinks: []string{
			"/",
			"/posts",
			"/posts/a",
		},
	},
	{
		// Assets only? No pages, but okay ...
		Paths: []string{"a.png", "b.png"},
	},
}

func TestDirectoryLayouts(t *testing.T) {
	for _, i := range layoutTests {
		base, err := os.MkdirTemp(os.TempDir(), "tacktest")
		assert.NoError(t, err)
		assert.NoError(t, os.MkdirAll(filepath.Join(base, ContentDir), 0755))
		assert.NoError(t, os.MkdirAll(filepath.Join(base, TemplateDir), 0755))
		for _, p := range i.Paths {
			if strings.HasSuffix(p, "/") {
				assert.NoError(t, os.MkdirAll(filepath.Join(base, ContentDir, p), 0755))
			} else {
				assert.NoError(t, os.MkdirAll(filepath.Join(base, ContentDir, filepath.Dir(p)), 0755))
				assert.NoError(t, os.WriteFile(filepath.Join(base, ContentDir, p), []byte{}, 0644))
			}
		}
		tacker, err := NewTacker(base)
		assert.NoError(t, err)
		assert.Len(t, tacker.Pages, len(i.Permalinks))
		for _, p := range tacker.Pages {
			assert.Contains(t, i.Permalinks, p.Permalink())
		}
		assert.NoError(t, os.RemoveAll(base))
	}
}

func TestDirectoryLayoutErrors(t *testing.T) {
	for _, i := range [][]string{
		// multiple root pages
		{"x.md", "index/x.md", "a/x.md"},
		{"0.index/x.md", "1.index/x.md"},
		// same page, different templates
		{"a.yaml", "b.yaml"},
	} {
		base, err := os.MkdirTemp(os.TempDir(), "tacktest")
		assert.NoError(t, err)
		assert.NoError(t, os.MkdirAll(filepath.Join(base, ContentDir), 0755))
		assert.NoError(t, os.MkdirAll(filepath.Join(base, TemplateDir), 0755))
		for _, p := range i {
			if strings.HasSuffix(p, "/") {
				assert.NoError(t, os.MkdirAll(filepath.Join(base, ContentDir, p), 0755))
			} else {
				assert.NoError(t, os.MkdirAll(filepath.Join(base, ContentDir, filepath.Dir(p)), 0755))
				assert.NoError(t, os.WriteFile(filepath.Join(base, ContentDir, p), []byte{}, 0644))
			}
		}
		_, err = NewTacker(base)
		assert.Error(t, err)
		assert.NoError(t, os.RemoveAll(base))
	}
}
func AssertDirEquals(t *testing.T, expected string, result string) {
	expList, err := FindFiles(expected)
	if err != nil {
		t.Errorf("unable to read dir %s: %s", expected, err)
	}

	for idx, fn := range expList {
		expList[idx] = strings.TrimPrefix(fn, expected)
	}

	resList, err := FindFiles(result)
	if err != nil {
		t.Errorf("unable to read dir %s: %s", result, err)
	}

	for idx, fn := range resList {
		resList[idx] = strings.TrimPrefix(fn, result)
	}

	if !StringSliceEqual(expList, resList) {
		t.Errorf("File lists differ: %s <--> %s", expected, result)
	}

	assert.Equal(t, expList, resList)

	for _, fn := range expList {
		expContent, err := os.ReadFile(expected + fn)
		if err != nil {
			t.Errorf("unable to read file %s: %s", fn, err)
		}

		resContent, err := os.ReadFile(result + fn)
		if err != nil {
			t.Errorf("unable to read file %s: %s", fn, err)
		}
		assert.Equal(t, expContent, resContent,
			"file content does not match for %s", fn)
	}
}

func StringSliceEqual(a []string, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	for idx, val := range a {
		if b[idx] != val {
			return false
		}
	}

	return true
}
