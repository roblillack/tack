package core

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
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
	BaseDir    string
	Metadata   map[string]interface{}
	Pages      []*Page
	Navigation []*Page
	Logger     *log.Logger
}

func NewTacker(dir string) (*Tacker, error) {
	mustache.AllowMissingVariables = true

	if !DirExists(dir) {
		return nil, fmt.Errorf("directory does not exist: %s", dir)
	}

	if !DirExists(filepath.Join(dir, ContentDir)) || !DirExists(filepath.Join(dir, TemplateDir)) {
		return nil, fmt.Errorf("does not look like a Tack-able site directory: %s", dir)
	}

	t := &Tacker{
		BaseDir: dir,
		Logger:  log.New(os.Stdout, "", 0),
	}

	if err := t.Reload(); err != nil {
		return nil, err
	}

	return t, nil
}

func (t *Tacker) Reload() error {
	if err := t.loadSiteMetadata(); err != nil {
		return err
	}
	if err := t.findAllPages(); err != nil {
		return err
	}

	navi := []*Page{}
	for _, i := range t.Pages {
		if err := i.Init(); err != nil {
			return err
		}
		if i.Parent == nil && !i.Floating {
			navi = append(navi, i)
		}
	}
	sort.Slice(navi, func(i, j int) bool {
		return strings.Compare(filepath.Base(navi[i].DiskPath), filepath.Base(navi[j].DiskPath)) == -1
	})
	t.Navigation = navi

	return nil
}

func (t *Tacker) Log(format string, args ...interface{}) {
	if t.Logger == nil {
		return
	}
	t.Logger.Printf(format+"\n", args...)
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
		t.Log("%s => %s (template: %s)", page.Permalink(), page.Name, page.Template)
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

		t.Log("Copying %s", strings.TrimPrefix(i, assetDir))
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
