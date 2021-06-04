package core

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/renderer/html"
)

var enumerationRegex = regexp.MustCompile(`^[0-9]+\.\s*`)

type Page struct {
	// available directly after construction.
	Name     string
	DiskPath string
	Tacker   *Tacker
	Floating bool

	inited bool
	// first available after call to Init()
	Parent    *Page
	Siblings  []*Page
	Assets    map[string]struct{}
	Variables map[string]interface{}
	Template  string
}

func NewPage(tacker *Tacker, realPath string) *Page {
	fn := filepath.Base(realPath)

	return &Page{
		Tacker:   tacker,
		DiskPath: realPath,
		Name:     enumerationRegex.ReplaceAllLiteralString(fn, ""),
		Floating: !enumerationRegex.MatchString(fn),
	}
}

func (p *Page) Permalink() string {
	if p.Parent == nil {
		if p.Name == "index" {
			return "/"
		}
		return "/" + p.Name
	}

	return p.Parent.Permalink() + "/" + p.Name
}

func (p *Page) TargetDir() []string {
	if p.Parent == nil {
		if p.Name == "index" {
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

func (p *Page) Init() error {
	parent := filepath.Dir(p.DiskPath)
	siblings := []*Page{}

	for _, i := range p.Tacker.Pages {
		if i.DiskPath == parent {
			p.Parent = i
		}
		if filepath.Dir(i.DiskPath) == parent && i != p && !i.Floating {
			siblings = append(siblings, i)
		}
	}

	sort.Slice(siblings, func(i, j int) bool {
		return strings.Compare(siblings[i].Name, siblings[j].Name) == -1
	})
	p.Siblings = siblings
	p.Assets = map[string]struct{}{}

	metadata := map[string]interface{}{}
	allFiles, err := filepath.Glob(filepath.Join(p.DiskPath, "*.*"))
	if err != nil {
		return err
	}
	for _, filename := range allFiles {
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
	// 			var pagePaths = new HashSet<string> ();
	// 			foreach (var page in Tacker.Pages) {
	// 				pagePaths.Add (page.DiskPath);
	// 			}

	// 			foreach (var subdir in Files.EnumerateAllSubdirs (DiskPath)) {
	// 				if (!pagePaths.Contains (subdir)) {
	// 					assets.AddAll (Files.GetAllFiles (subdir));
	// 				}
	// 			}

	// 			Assets = new HashSet<string> (assets.Select (x => x.Replace (DiskPath, "")));
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
	for _, i := range p.Siblings {
		s = append(s, i.Name)
	}

	destDir := filepath.Join(append([]string{p.Tacker.BaseDir, TargetDir}, p.TargetDir()...)...)

	fmt.Printf("Generating %s\n", p.Name)
	fmt.Printf(" - permalink: %s\n", p.Permalink())
	fmt.Printf(" - destdir: %s\n", destDir)
	fmt.Printf(" - ancestors: %s\n", strings.Join(a, " << "))
	fmt.Printf(" - siblings: %s\n", strings.Join(s, ", "))

	if err := os.MkdirAll(destDir, 0755); err != nil {
		return err
	}

	tpl, err := p.Tacker.FindTemplate(p.Template)
	if err != nil {
		return fmt.Errorf("unable to load template '%s' when rendering '%s': %s", p.Template, p.Permalink(), err)
	}

	// using (var writer = File.CreateText(Path.Combine (Tacker.TargetDir + Permalink, "index.html"))) {
	//     Tacker.FindTemplate (Template).Render (new DictWrapper (this, new RenderContext (this)), writer, Tacker.FindTemplate);
	// }

	f, err := os.OpenFile(filepath.Join(destDir, "index.html"), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	if err := tpl.Render(p, f); err != nil {
		return fmt.Errorf("unable to render template '%s' when rendering '%s': %s", p.Template, p.Permalink(), err)
	}

	for i := range p.Assets {
		fmt.Printf("Copying ...%s\n", i)
		if err := os.MkdirAll(filepath.Dir(filepath.Join(destDir, i)), 0755); err != nil {
			return err
		}
		if err := CopyFile(filepath.Join(p.DiskPath, i), filepath.Join(destDir, i)); err != nil {
			return err
		}
	}

	return nil
}
