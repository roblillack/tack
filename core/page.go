package core

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/renderer/html"
)

var enumerationRegex = regexp.MustCompile(`^[0-9]+\.\s*`)
var dateRegex = regexp.MustCompile(`^([0-9]{4}-[0-9]{2}-[0-9]{2})[\.\-]\s*`)

type Page struct {
	// available directly after construction.
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
}

func NewPage(tacker *Tacker, realPath string) *Page {
	fn := filepath.Base(realPath)

	page := &Page{
		Tacker:   tacker,
		DiskPath: realPath,
		Name:     fn,
		Floating: true,
	}

	if enumerationRegex.MatchString(fn) {
		page.Floating = false
		page.Name = enumerationRegex.ReplaceAllLiteralString(fn, "")
	} else if m := dateRegex.FindStringSubmatch(fn); len(m) == 2 {
		if d, err := time.Parse("2006-01-02", m[1]); err == nil {
			page.Name = dateRegex.ReplaceAllLiteralString(fn, "")
			page.Date = d
			page.Floating = false
		}
	}

	return page
}

func (p *Page) Root() bool {
	return p.DiskPath == filepath.Join(p.Tacker.BaseDir, ContentDir) || p.Name == "index" && filepath.Dir(p.DiskPath) == filepath.Join(p.Tacker.BaseDir, ContentDir)
}

func (p *Page) Permalink() string {
	if p.Parent == nil {
		if p.Root() {
			return "/"
		}
		return "/" + p.Name
	}

	return p.Parent.Permalink() + "/" + p.Name
}

func (p *Page) TargetDir() []string {
	if p.Parent == nil {
		if p.Root() {
			return []string{}
		}
		return []string{p.Name}
	}

	return append(p.Parent.TargetDir(), p.Name)
}

func (p *Page) Ancestors() []*Page {
	r := []*Page{}

	for i := p.Parent; i != nil; i = i.Parent {
		r = append([]*Page{i}, r...)
	}

	return r
}

func (p *Page) Siblings() []*Page {
	r := []*Page{}

	for _, i := range p.SiblingsAndMe {
		if i != p {
			r = append(r, i)
		}
	}

	return r
}

func (p *Page) Init() error {
	parent := filepath.Dir(p.DiskPath)
	siblingsAndMe := []*Page{}
	children := []*Page{}
	posts := []*Page{}

	for _, i := range p.Tacker.Pages {
		if i.DiskPath == parent {
			p.Parent = i
		}
		if filepath.Dir(i.DiskPath) == parent && !i.Floating {
			siblingsAndMe = append(siblingsAndMe, i)
		}
		if filepath.Dir(i.DiskPath) == p.DiskPath {
			if i.Date.IsZero() {
				children = append(children, i)
			} else {
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

	metadata := map[string]interface{}{}
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
		if ext == "yml" || ext == "yaml" {
			if p.Template != "" {
				return fmt.Errorf("multiple templates requested! %s vs. %s", p.Template, base)
			}
			p.Template = base
			md, err := ProcessMetadata(filename)
			if err != nil {
				return fmt.Errorf("unable to process metadata for %s: %w", p.Permalink(), err)
			}
			for k, v := range md {
				metadata[k] = v
			}
		} else if ext == "md" || ext == "mkd" {
			markdown, err := os.ReadFile(filename)
			if err != nil {
				return err
			}
			buf := &bytes.Buffer{}
			parser := goldmark.New(goldmark.WithRendererOptions(html.WithUnsafe()))
			if err := parser.Convert(markdown, buf); err != nil {
				return err
			}
			metadata[base] = buf.String()
		} else {
			p.Assets[strings.TrimPrefix(filename, p.DiskPath)] = struct{}{}
		}
	}

	p.Variables = metadata
	p.inited = true
	return nil
}

func (p *Page) Generate() error {
	if !p.inited {
		if err := p.Init(); err != nil {
			return err
		}
	}

	a := []string{}
	for _, i := range p.Ancestors() {
		a = append(a, i.Name)
	}

	s := []string{}
	for _, i := range p.SiblingsAndMe {
		s = append(s, i.Name)
	}

	destDir := filepath.Join(append([]string{p.Tacker.BaseDir, TargetDir}, p.TargetDir()...)...)

	p.Tacker.Log("Generating %s", p.Name)
	p.Tacker.Log(" - permalink: %s", p.Permalink())
	p.Tacker.Log(" - destdir: %s", destDir)
	p.Tacker.Log(" - ancestors: %s", strings.Join(a, " << "))
	p.Tacker.Log(" - siblings: %s", strings.Join(s, ", "))

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
		p.Tacker.Log("Copying ...%s", i)
		if err := os.MkdirAll(filepath.Dir(filepath.Join(destDir, i)), 0755); err != nil {
			return err
		}
		if err := CopyFile(filepath.Join(p.DiskPath, i), filepath.Join(destDir, i)); err != nil {
			return err
		}
	}

	return nil
}
