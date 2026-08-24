package core

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode"

	"github.com/yuin/goldmark"
	meta "github.com/yuin/goldmark-meta"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

var enumerationRegex = regexp.MustCompile(`^[0-9]+\.\s*`)
var dateRegex = regexp.MustCompile(`^([0-9]{4}-[0-9]{2}-[0-9]{2})[\.\-]\s*`)

// Page is the main structure holding page content. Some of the fields are
// only available after the page has been initialized using Init().
type Page struct {
	// available directly after construction.
	Slug     string
	Name     string
	DiskPath string
	Tacker   *Tacker
	Floating bool
	Date     time.Time

	inited bool
	// first available after call to Init()
	Parent        *Page
	SiblingsAndMe []*Page
	Children      []*Page
	Posts         []*Page
	Assets        map[string]struct{}
	Variables     map[string]interface{}
	Template      string
	addTagPages   bool
}

// NewPage creates a new page structure for the specified Tacker
// based on the given path of the page's directory. No file i/o
// will take place here just yet.
func NewPage(tacker *Tacker, realPath string) *Page {
	fn := filepath.Base(realPath)
	if realPath == filepath.Join(tacker.BaseDir, ContentDir) {
		fn = "index"
	}

	page := &Page{
		Tacker:   tacker,
		DiskPath: realPath,
		Slug:     fn,
		Floating: true,
	}

	if enumerationRegex.MatchString(fn) {
		page.Floating = false
		page.Slug = enumerationRegex.ReplaceAllLiteralString(fn, "")
	} else if m := dateRegex.FindStringSubmatch(fn); len(m) == 2 {
		if d, err := time.Parse("2006-01-02", m[1]); err == nil {
			page.Slug = dateRegex.ReplaceAllLiteralString(fn, "")
			page.Date = d
			page.Floating = false
		}
	}

	page.Name = strings.ReplaceAll(titleWords(page.Slug), "-", " ")

	return page
}

// titleWords upper-cases the first letter of every word, leaving the remainder
// of each word untouched. It replaces the deprecated strings.Title and keeps its
// behavior, except that an apostrophe no longer starts a new word: strings.Title
// turned "don't" into "Don'T".
func titleWords(s string) string {
	prev := ' '
	return strings.Map(func(r rune) rune {
		if isWordSeparator(prev) {
			prev = r
			return unicode.ToTitle(r)
		}
		prev = r
		return r
	}, s)
}

// isWordSeparator reports whether r separates two words. Letters and digits are
// word-internal, and so are apostrophes, so that the "s" in "it's" does not
// begin a word. Everything else -- whitespace, punctuation and symbols alike --
// separates.
//
// This follows the rule the deprecated strings.Title used, with three fixes: it
// treated apostrophes as separators ("don't" became "Don'T"); it separated on
// non-ASCII runes only when they were whitespace, so dashes such as the en dash
// never started a word ("a\u2013b" became "A\u2013b"); and it treated underscores as
// word-internal, so "foo_bar" was left alone rather than titled like "foo-bar".
func isWordSeparator(r rune) bool {
	if r == '\'' || r == '\u2019' {
		return false
	}
	if r <= 0x7F {
		switch {
		case '0' <= r && r <= '9':
			return false
		case 'a' <= r && r <= 'z':
			return false
		case 'A' <= r && r <= 'Z':
			return false
		}
		return true
	}
	if unicode.IsLetter(r) || unicode.IsDigit(r) {
		return false
	}
	return unicode.IsSpace(r) || unicode.IsPunct(r) || unicode.IsSymbol(r)
}

// Root determines if the current page is the root page of the website being
// tacked. The root page might be stored in the top-level content directory
// or a directory with the slug "index" just below the top level.
func (p *Page) Root() bool {
	return p.DiskPath == filepath.Join(p.Tacker.BaseDir, ContentDir) ||
		p.Slug == "index" && filepath.Dir(p.DiskPath) == filepath.Join(p.Tacker.BaseDir, ContentDir)
}

// Permalink return an absolute path to the current page based on its and it's
// ancestor pages' slugs. The Page must be Init()ed prior to calling this.
func (p *Page) Permalink() string {
	if p.Parent == nil {
		if p.Root() {
			return "/"
		}
		return "/" + p.Slug
	}

	return path.Join(p.Parent.Permalink(), p.Slug)
}

// TargetDir returns the (absolute) path to the directory which will contain
// this page's HTML and further assets. The Page must be Init()ed prior to
// calling this.
func (p *Page) TargetDir() []string {
	if p.Parent == nil {
		if p.Root() {
			return []string{}
		}
		return []string{p.Slug}
	}

	return append(p.Parent.TargetDir(), TagSlug(p.Slug))
}

// Ancestors returns a slice of all of this page's ancestors, starting with
// the immediate parent page and ending with the root page. The Page must be
// Init()ed prior to calling this.
func (p *Page) Ancestors() []*Page {
	r := []*Page{}

	for i := p.Parent; i != nil; i = i.Parent {
		r = append([]*Page{i}, r...)
	}

	return r
}

// Siblings returns a slice of all sibling pages of the current one. The Page
// must be Init()ed prior to calling this.
func (p *Page) Siblings() []*Page {
	r := []*Page{}

	for _, i := range p.SiblingsAndMe {
		if i != p {
			r = append(r, i)
		}
	}

	return r
}

// Post returns `true` if the current page has a post date defined as
// part of the content directory name.
func (p *Page) Post() bool {
	return !p.Date.IsZero()
}

// Init initializes the page content, by reading the content and metadata from
// the disk, resolving the used template and creating the necessary structures
// to reference other pages from this one.
func (p *Page) Init() error {
	parent := filepath.Dir(p.DiskPath)
	siblingsAndMe := []*Page{}
	children := []*Page{}
	posts := []*Page{}

	for _, i := range p.Tacker.Pages {
		if i.DiskPath == parent {
			p.Parent = i
		}
		if filepath.Dir(i.DiskPath) == parent && !i.Floating && !i.Post() {
			siblingsAndMe = append(siblingsAndMe, i)
		}
		if filepath.Dir(i.DiskPath) == p.DiskPath {
			if !i.Post() && !i.Floating {
				children = append(children, i)
			} else if i.Post() {
				posts = append(posts, i)
			}
		}
	}

	sort.Slice(siblingsAndMe, func(i, j int) bool {
		return strings.Compare(filepath.Base(siblingsAndMe[i].DiskPath), filepath.Base(siblingsAndMe[j].DiskPath)) == -1
	})
	p.SiblingsAndMe = siblingsAndMe
	sort.Slice(children, func(i, j int) bool {
		return strings.Compare(filepath.Base(children[i].DiskPath), filepath.Base(children[j].DiskPath)) == -1
	})
	p.Children = children
	sort.Slice(posts, func(i, j int) bool {
		return posts[i].Date.After(posts[j].Date)
	})
	p.Posts = posts
	p.Assets = map[string]struct{}{}
	p.Variables = map[string]interface{}{}

	allFiles, err := FindFiles(p.DiskPath)
	if err != nil {
		return err
	}
nextFile:
	for _, filename := range allFiles {
		for _, i := range p.Tacker.Pages {
			if i == p || strings.HasPrefix(p.DiskPath, i.DiskPath+string(os.PathSeparator)) {
				continue
			}
			if strings.HasPrefix(filename, i.DiskPath+string(os.PathSeparator)) {
				continue nextFile
			}
		}
		ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(filename)), ".")
		base := BasenameWithoutExtension(filename)
		switch ext {
		case "yml", "yaml":
			md, err := ProcessMetadata(filename)
			if err != nil {
				return fmt.Errorf("unable to process metadata for %s: %w", p.Permalink(), err)
			}
			md["template"] = base
			if err := p.addVariables(md); err != nil {
				return err
			}
		case "md", "mkd":
			markdown, err := os.ReadFile(filename)
			if err != nil {
				return err
			}
			buf := &bytes.Buffer{}
			engine := goldmark.New(goldmark.WithRendererOptions(html.WithUnsafe()), goldmark.WithExtensions(meta.Meta))
			context := parser.NewContext()
			if err := engine.Convert(markdown, buf, parser.WithContext(context)); err != nil {
				return err
			}
			if err := p.addVariables(meta.Get(context)); err != nil {
				return err
			}

			p.Variables[base] = buf.String()
		default:
			p.Assets[strings.TrimPrefix(filename, p.DiskPath)] = struct{}{}
		}
	}

	p.inited = true
	return nil
}

func (p *Page) addVariables(md map[string]interface{}) error {
	for k, v := range md {
		if k == "template" {
			newTemplate := fmt.Sprint(v)
			if p.Template != "" && p.Template != newTemplate {
				return fmt.Errorf("%s: multiple templates requested! %s vs. %s", p.DiskPath, p.Template, newTemplate)
			}
			p.Template = newTemplate
			continue
		}
		if k == "tags" {
			if bv, ok := v.(bool); ok && bv {
				p.addTagPages = true
				continue
			}

			if tags, ok := v.([]interface{}); ok {
				for _, i := range tags {
					s, ok := i.(string)
					if s == "" || !ok {
						continue
					}
					p.Tacker.addTag(s, p)
				}
			}
		}
		p.Variables[k] = v
	}

	return nil
}

// Generate renders the current page given all the content and metadata read
// from disk and the configured template. If not done already, calling this
// function will initialize the page using Init().
func (p *Page) Generate() error {
	if !p.inited {
		if err := p.Init(); err != nil {
			return err
		}
	}

	a := []string{}
	for _, i := range p.Ancestors() {
		a = append(a, i.Slug)
	}

	s := []string{}
	for _, i := range p.SiblingsAndMe {
		s = append(s, i.Slug)
	}

	destDir := filepath.Join(append([]string{p.Tacker.BaseDir, TargetDir}, p.TargetDir()...)...)

	p.Tacker.Debug("Generating %s", p.Slug)
	par := "-"
	if p.Parent != nil {
		par = p.Parent.DiskPath
	}
	p.Tacker.Debug(" - disk path: %s", p.DiskPath)
	p.Tacker.Debug(" - parent: %s", par)
	p.Tacker.Debug(" - permalink: %s", p.Permalink())
	p.Tacker.Debug(" - destdir: %s", destDir)
	p.Tacker.Debug(" - ancestors: %s", strings.Join(a, " << "))
	p.Tacker.Debug(" - siblings: %s", strings.Join(s, ", "))

	if err := os.MkdirAll(destDir, 0755); err != nil {
		return err
	}

	tpl, err := p.Tacker.FindTemplate(p.Template)
	if err != nil {
		return fmt.Errorf("unable to load template '%s' when rendering '%s': %s", p.Template, p.Permalink(), err)
	}

	f, err := os.OpenFile(filepath.Join(destDir, "index.html"), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	if err := tpl.Render(p, f); err != nil {
		return fmt.Errorf("unable to render template '%s' when rendering '%s': %s", p.Template, p.Permalink(), err)
	}

	for i := range p.Assets {
		p.Tacker.Debug("Copying ...%s", i)
		if err := os.MkdirAll(filepath.Dir(filepath.Join(destDir, i)), 0755); err != nil {
			return err
		}
		if err := CopyFile(filepath.Join(p.DiskPath, i), filepath.Join(destDir, i)); err != nil {
			return err
		}
	}

	return nil
}
