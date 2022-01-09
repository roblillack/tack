package core

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/cbroglie/mustache"
	"gopkg.in/yaml.v2"
)

const ContentDir = "content"
const TemplateDir = "templates"
const TargetDir = "output"
const AssetDir = "public"

var TemplateExtensions = []string{"mustache", "mu", "stache"}
var MetadataExtensions = []string{"yaml", "yml"}
var MarkupExtensions = []string{"md", "mkd"}

type Tacker struct {
	BaseDir     string
	Metadata    map[string]interface{}
	Pages       []*Page
	Navigation  []*Page
	Posts       []*Page
	Tags        map[string][]*Page
	TagNames    map[string]map[string]int
	TagIndex    *Page
	Logger      *log.Logger
	DebugLogger *log.Logger
}

func NewTacker(dir string) (*Tacker, error) {
	mustache.AllowMissingVariables = true

	if !DirExists(dir) {
		return nil, fmt.Errorf("directory does not exist: %s", dir)
	}

	if !DirExists(filepath.Join(dir, ContentDir)) || !DirExists(filepath.Join(dir, TemplateDir)) {
		return nil, fmt.Errorf("does not look like a Tack-able site directory: %s", dir)
	}

	logger := log.New(os.Stdout, "", 0)

	t := &Tacker{
		BaseDir:     dir,
		Logger:      logger,
		DebugLogger: logger,
	}

	if err := t.Reload(); err != nil {
		return nil, err
	}

	return t, nil
}

func (t *Tacker) Reload() error {
	t.TagIndex = nil
	t.Tags = nil
	t.TagNames = nil

	if err := t.loadSiteMetadata(); err != nil {
		return err
	}
	if err := t.findAllPages(); err != nil {
		return err
	}

	navi := []*Page{}
	posts := []*Page{}
	for _, i := range t.Pages {
		if err := i.Init(); err != nil {
			return err
		}
		if i.Parent == nil && !i.Floating && !i.Post() {
			navi = append(navi, i)
		}
		if i.Post() {
			posts = append(posts, i)
		}
	}
	sort.Slice(navi, func(i, j int) bool {
		return strings.Compare(filepath.Base(navi[i].DiskPath), filepath.Base(navi[j].DiskPath)) == -1
	})
	t.Navigation = navi

	sort.Slice(posts, func(i, j int) bool {
		return posts[i].Date.After(posts[j].Date)
	})
	t.Posts = posts

	for _, i := range t.Pages {
		if !i.addTagPages {
			continue
		}
		if t.TagIndex != nil {
			return fmt.Errorf("multiple tag index pages detected: %s <-> %s", t.TagIndex.DiskPath, i.DiskPath)
		}
		t.TagIndex = i
		for slug, taggedPages := range t.Tags {
			tag := t.Tag(slug)

			template := ""
			if i.Template != "" {
				template = i.Template
			}

			vars := map[string]interface{}{}
			for k, v := range i.Variables {
				if k == "name" {
					continue
				}
				if s, ok := v.(string); ok && k == "template_tags" {
					template = s
					continue
				}
				vars[k] = v
			}
			vars["count"] = tag.Count

			page := &Page{
				inited:    true,
				Tacker:    t,
				DiskPath:  "",
				Slug:      tag.Slug,
				Name:      tag.Name,
				Floating:  true,
				Parent:    i,
				Posts:     taggedPages,
				Template:  template,
				Variables: vars,
			}
			t.Pages = append(t.Pages, page)
			i.Children = append(i.Children, page)
		}
	}

	return nil
}

func (t *Tacker) Log(format string, args ...interface{}) {
	if t.Logger == nil {
		return
	}
	t.Logger.Printf(format+"\n", args...)
}

func (t *Tacker) Debug(format string, args ...interface{}) {
	if t.DebugLogger == nil {
		return
	}
	t.DebugLogger.Printf(format+"\n", args...)
}

func (t *Tacker) Tack() error {
	t.Log("Tacking up %s (%d pages)", t.BaseDir, len(t.Pages))

	if _, err := os.Stat(filepath.Join(t.BaseDir, TargetDir)); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	} else if err == nil {
		if err := os.RemoveAll(filepath.Join(t.BaseDir, TargetDir)); err != nil {
			return err
		}
	}

	for _, page := range t.Pages {
		t.Debug("%s => %s (template: %s)", page.Permalink(), page.Slug, page.Template)
		if err := page.Generate(); err != nil {
			return err
		}
	}

	assetDir := filepath.Join(t.BaseDir, AssetDir)
	targetDir := filepath.Join(t.BaseDir, TargetDir)
	assets, err := FindFiles(assetDir)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	for _, i := range assets {
		if !strings.HasPrefix(i, assetDir) {
			continue
		}

		dest := targetDir + strings.TrimPrefix(i, assetDir)
		if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
			return err
		}

		t.Debug("Copying %s", strings.TrimPrefix(i, assetDir))
		if err := CopyFile(i, dest); err != nil {
			return err
		}
	}

	return nil
}

func (t *Tacker) FindTemplate(name string) (*Template, error) {
	if name == "" {
		name = "default"
	}

	fn := FirstFileWithExtension(filepath.Join(t.BaseDir, TemplateDir), name, TemplateExtensions...)
	if fn == "" {
		return nil, fmt.Errorf("Template '%s' not found", name)
	}

	provider := &mustache.FileProvider{
		Paths:      []string{filepath.Join(t.BaseDir, TemplateDir)},
		Extensions: []string{},
	}
	for _, i := range TemplateExtensions {
		if i == "" {
			provider.Extensions = append(provider.Extensions, i)
		} else {
			provider.Extensions = append(provider.Extensions, "."+i)
		}
	}

	tpl, err := mustache.ParseFilePartials(fn, provider)
	if err != nil {
		return nil, err
	}

	return &Template{tpl}, nil
}

func (t *Tacker) addTag(name string, page *Page) {
	slug := TagSlug(name)

	if t.Tags == nil {
		t.Tags = map[string][]*Page{}
	}
	if t.Tags[slug] == nil {
		t.Tags[slug] = []*Page{}
	}
	if t.TagNames == nil {
		t.TagNames = map[string]map[string]int{}
	}
	if t.TagNames[slug] == nil {
		t.TagNames[slug] = map[string]int{}
	}

	for _, i := range t.Tags[slug] {
		if i == page {
			return
		}
	}

	t.Tags[slug] = append(t.Tags[slug], page)
	t.TagNames[slug][name] = t.TagNames[slug][name] + 1
}

func (t *Tacker) Tag(name string) Tag {
	slug := TagSlug(name)

	if t.Tags == nil || t.Tags[slug] == nil {
		return Tag{Slug: slug, Name: name, Count: 0, Permalink: ""}
	}
	link := ""
	if t.TagIndex != nil {
		link = path.Join(t.TagIndex.Permalink(), slug)
	}

	bestName := ""
	bestCount := 0
	for name, count := range t.TagNames[slug] {
		if count > bestCount {
			bestName = name
			bestCount = count
		}
	}

	return Tag{
		Name:      bestName,
		Slug:      slug,
		Count:     len(t.Tags[slug]),
		Permalink: link,
	}
}

func (t *Tacker) findAllPages() error {
	pagesPath := filepath.Join(t.BaseDir, ContentDir)

	m, err := FindDirsWithFiles(pagesPath, append(MarkupExtensions, MetadataExtensions...)...)
	if err != nil {
		return err
	}

	all := []*Page{}
	seen := map[string]struct{}{}
	rootPage := ""

	for _, pageDir := range m {
		// backfill all ancestors, even if they do not contain sufficient files itself ...
		for p := pageDir; strings.HasPrefix(p+string(os.PathSeparator), pagesPath); p = filepath.Dir(p) {
			if _, visited := seen[p]; visited {
				continue
			}
			page := NewPage(t, p)
			if page.Root() {
				if rootPage != "" {
					if p == pageDir {
						return fmt.Errorf("multiple root pages detected: %s <-> %s", rootPage, p)
					} else {
						continue
					}
				}
				rootPage = p
			}
			all = append(all, page)
			seen[p] = struct{}{}
		}
	}
	t.Pages = all
	return nil
}

func ProcessMetadata(file string) (map[string]interface{}, error) {
	r, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	res := map[string]interface{}{}
	if err := yaml.NewDecoder(r).Decode(&res); err != nil && !errors.Is(err, io.EOF) {
		return nil, err
	}
	d := []string{}
	for k, v := range res {
		if strings.HasPrefix(k, ":") {
			res[strings.TrimPrefix(k, ":")] = v
			d = append(d, k)
		}
	}
	for _, i := range d {
		delete(res, i)
	}

	return res, nil
}

func (t *Tacker) loadSiteMetadata() error {
	files, err := filepath.Glob(filepath.Join(t.BaseDir, "*.*"))
	if err != nil {
		return err
	}
	for _, i := range files {
		if ext := strings.ToLower(filepath.Ext(i)); ext != ".yaml" && ext != ".yml" {
			continue
		}
		md, err := ProcessMetadata(i)
		if err != nil {
			return err
		}
		if t.Metadata == nil {
			t.Metadata = md
		} else {
			for k, v := range md {
				t.Metadata[k] = v
			}
		}
	}

	return nil
}
