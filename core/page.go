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

	"github.com/yuin/goldmark"
	meta "github.com/yuin/goldmark-meta"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

var enumerationRegex = regexp.MustCompile(`^[0-9]+\.\s*`)
var dateRegex = regexp.MustCompile(`^([0-9]{4}-[0-9]{2}-[0-9]{2})[\.\-]\s*`)

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

	page.Name = strings.Replace(strings.Title(page.Slug), "-", " ", -1)

	return page
}

func (p *Page) Root() bool {
	return p.DiskPath == filepath.Join(p.Tacker.BaseDir, ContentDir) || p.Slug == "index" && filepath.Dir(p.DiskPath) == filepath.Join(p.Tacker.BaseDir, ContentDir)
}

func (p *Page) Permalink() string {
	if p.Parent == nil {
		if p.Root() {
			return "/"
		}
		return "/" + p.Slug
	}

	return path.Join(p.Parent.Permalink(), p.Slug)
}

func (p *Page) TargetDir() []string {
	if p.Parent == nil {
		if p.Root() {
			return []string{}
		}
		return []string{p.Slug}
	}

	return append(p.Parent.TargetDir(), TagSlug(p.Slug))
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

func (p *Page) Post() bool {
	return !p.Date.IsZero()
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
		if ext == "yml" || ext == "yaml" {
			md, err := ProcessMetadata(filename)
			if err != nil {
				return fmt.Errorf("unable to process metadata for %s: %w", p.Permalink(), err)
			}
			md["template"] = base
			if err := p.addVariables(md); err != nil {
				return err
			}
		} else if ext == "md" || ext == "mkd" {
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
		} else {
			p.Assets[strings.TrimPrefix(filename, p.DiskPath)] = struct{}{}
		}
	}

	p.inited = true
	return nil
}

func (p *Page) addVariables(md map[string]interface{}) error {
	for k, v := range md {
		if k == "template" {
			if p.Template != "" {
				return fmt.Errorf("multiple templates requested! %s vs. %s", p.Template, v)
			}
			p.Template = fmt.Sprint(v)
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

	if p.DiskPath != "" {
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
	}

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
		p.Tacker.Debug(" - copying %s", i)
		if err := os.MkdirAll(filepath.Dir(filepath.Join(destDir, i)), 0755); err != nil {
			return err
		}
		if err := CopyFile(filepath.Join(p.DiskPath, i), filepath.Join(destDir, i)); err != nil {
			return err
		}
	}

	if dict, ok := p.Variables["extra_files"].(map[interface{}]interface{}); ok {
		for k, v := range dict {
			fn := k.(string)
			templateName := v.(string)
			if fn == "" || templateName == "" {
				continue
			}

			p.Tacker.Log(" - rendering %s", fn)
			tpl, err := p.Tacker.FindTemplate(templateName)
			if err != nil {
				return fmt.Errorf("unable to load template '%s' for extra file '%s' when rendering '%s': %s", s, fn, p.Permalink(), err)
			}

			f, err := os.OpenFile(filepath.Join(destDir, fn), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
			if err != nil {
				return err
			}
			defer f.Close()

			if err := tpl.Render(p, f); err != nil {
				return fmt.Errorf("unable to render template '%s' for extra file '%s' when rendering '%s': %s", s, fn, p.Permalink(), err)
			}
		}
	}

	return nil
}
