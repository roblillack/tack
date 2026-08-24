package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTagPages(t *testing.T) {
	tacker, base := tackSite(t, map[string]string{
		// Tag index pages keep their own name and may use a dedicated template
		// for the generated tag pages.
		"content/1.index/default.yaml": "tags: true\nname: All Tags\ntemplate_tags: tag\n",
		// Non-strings and empty tags are ignored.
		"content/2012-08-24.first/body.md":  "---\ntags: [\"Cars\", \"boats\", \"\", 42]\n---\n",
		"content/2019-07-31.second/body.md": "---\ntags: [\"cars\"]\n---\n",
		"content/2021-06-06.third/body.md":  "---\ntags: [\"Cars\", \"anchors\"]\n---\n",
		"templates/default.mustache":        "{{name}}|{{#tags}}{{name}}:{{count}} {{/tags}}",
		"templates/tag.mustache":            "{{name}} ({{count}}):{{#posts}} {{name}}{{/posts}}",
	})

	// Tags are sorted by usage first, then alphabetically. The most commonly
	// used spelling of a tag wins.
	assert.Equal(t, "All Tags|Cars:3 anchors:1 boats:1 ", generated(t, base, "index.html"))
	assert.Equal(t, "Cars (3): First Second Third", generated(t, base, "cars", "index.html"))
	assert.Equal(t, "boats (1): First", generated(t, base, "boats", "index.html"))
	assert.Equal(t, "First|Cars:3 boats:1 ", generated(t, base, "first", "index.html"))

	assert.Equal(t, Tag{Name: "Cars", Slug: "cars", Count: 3, Permalink: "/cars"}, tacker.Tag("Cars"))
	// Unknown tags are reported as unused.
	assert.Equal(t, Tag{Name: "Trains", Slug: "trains"}, tacker.Tag("Trains"))
}

func TestTagsAreOnlyCountedOncePerPage(t *testing.T) {
	tacker, _ := tackSite(t, map[string]string{
		"content/default.yaml":       "tags: [\"Cars\", \"cars\"]\n",
		"templates/default.mustache": "{{#tags}}{{name}}:{{count}}{{/tags}}",
	})

	assert.Equal(t, Tag{Name: "Cars", Slug: "cars", Count: 1}, tacker.Tag("cars"))
}

func TestMultipleTagIndexPages(t *testing.T) {
	base := writeFiles(t, map[string]string{
		"content/1.a/default.yaml":   "tags: true",
		"content/2.b/default.yaml":   "tags: true",
		"templates/default.mustache": "Hi",
	})

	_, err := NewTacker(base)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "multiple tag index pages")
}

func TestPostsOfTopLevelPages(t *testing.T) {
	// The root page lives in its own directory here, so the posts next to it
	// have no parent page of their own ...
	_, base := tackSite(t, map[string]string{
		"content/0.index/default.yaml": "posts_limit: 2\n",
		"content/2012-08-24.a/body.md": "",
		"content/2019-07-31.b/body.md": "",
		"content/2021-06-06.c/body.md": "",
		"templates/default.mustache":   "{{name}}:{{#posts}}{{name}} {{/posts}}",
	})

	// ... they still show up on the root page, newest first and limited to the
	// number of posts requested.
	assert.Equal(t, "Index:C B ", generated(t, base, "index.html"))
}

func TestPostsAreInheritedFromParent(t *testing.T) {
	_, base := tackSite(t, map[string]string{
		"content/1.posts/default.yaml":         "",
		"content/1.posts/1.about/default.yaml": "",
		"content/1.posts/2012-08-24.a/body.md": "",
		"templates/default.mustache":           "{{name}}:{{#posts}}{{name}} {{/posts}}",
	})

	assert.Equal(t, "Posts:A ", generated(t, base, "posts", "index.html"))
	// The “about” page has no posts of its own, so it lists its parent's.
	assert.Equal(t, "About:A ", generated(t, base, "posts", "about", "index.html"))
	// The root page has neither its own posts nor a parent to ask.
	assert.Equal(t, "Index:", generated(t, base, "index.html"))
}
