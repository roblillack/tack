package core

import (
	"errors"
	"fmt"
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

var TemplateExtensions = []string{"mustache"}
var MetadataExtensions = []string{"yml"}
var MarkupExtensions = []string{"mkd"}

type Tacker struct {
	BaseDir    string
	Metadata   map[string]interface{}
	Pages      []*Page
	Navigation []*Page
}

// public delegate void LogFn (string format, params object[] args);
// public LogFn Logger { get; set; }

// Markdown markdown;
// IDictionary<string, AssetFilter> assetFilters;

func NewTacker(dir string) (*Tacker, error) {
	// markdown = new Markdown ();
	// markdown.AutoHyperlink = true;

	// assetFilters = new Dictionary<string, AssetFilter> ();
	// assetFilters.Add ("less", new LessFilter ());

	mustache.AllowMissingVariables = true

	if !DirExists(dir) {
		return nil, fmt.Errorf("directory does not exist: %s", dir)
	}

	if !DirExists(filepath.Join(dir, ContentDir)) || !DirExists(filepath.Join(dir, TemplateDir)) {
		return nil, fmt.Errorf("does not look like a Tack-able site directory: %s", dir)
	}

	t := &Tacker{
		BaseDir: dir,
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
	fmt.Printf(format+"\n", args...)
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
	if err != nil {
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
		// 	if (assetFilters.ContainsKey (Path.GetExtension (i).Replace (".", ""))) {
		// 		var filter = assetFilters [Path.GetExtension (i).Replace (".", "")];
		// 		Log ("Applying {0} to {1} ...", filter.GetType ().Name, i);
		// 		filter.Filter (this, i);
		// 	} else {
		fmt.Printf("Copying %s\n", strings.TrimPrefix(i, assetDir))
		if err := CopyFile(i, dest); err != nil {
			return err
		}
		// 	}
	}

	return nil
}

func (t *Tacker) FindTemplate(name string) (*Template, error) {
	if name == "" {
		name = "default"
	}

	tpl, err := mustache.ParseFilePartials(filepath.Join(t.BaseDir, TemplateDir, name+".mustache"), &mustache.FileProvider{
		Paths:      []string{filepath.Join(t.BaseDir, TemplateDir)},
		Extensions: []string{".mustache"},
	})
	if err != nil {
		return nil, err
	}

	return &Template{tpl}, nil
}

func (t *Tacker) findAllPages() error {
	all := []*Page{}
	m, err := FindDirsWithFiles(filepath.Join(t.BaseDir, ContentDir), append(MarkupExtensions, MetadataExtensions...)...)
	if err != nil {
		return err
	}
	for _, pageDir := range m {
		all = append(all, NewPage(t, pageDir))
	}
	t.Pages = all
	return nil
}

// 		IEnumerable<string> FindAllAssets ()
// 		{
// 			return Files.EnumerateAllFiles (AssetDir);
// 		}

// 		public IDictionary<string, object> ProcessMetadata (string file)
// 		{
// 			foreach (var ext in METADATA_LANGS) {
// 				if (file.EndsWith ("." + ext)) {
// 					var map = new Dictionary<string, object> ();
// 					var stream = new YamlStream ();
// 					stream.Load (new StreamReader (file));

// 					foreach (var doc in stream.Documents) {
// 						if (doc.RootNode is YamlMappingNode) {
// 							var seq = doc.RootNode as YamlMappingNode;
// 							foreach (var node in seq.Children) {
// 								var key = node.Key as YamlScalarNode;
// 								object val = node.Value;
// 								if (val is YamlScalarNode && (val as YamlScalarNode).Style == YamlDotNet.Core.ScalarStyle.Literal) {
// 									val = markdown.Transform (val.ToString ());
// 								}
// 								map.Add (key.Style == YamlDotNet.Core.ScalarStyle.Plain ?
// 								         key.Value.Substring (1) : key.Value,
// 								         val);
// 							}
// 						}
// 					}
// 					return map;
// 				}
// 			}

// 			// Not a known meta-data format
// 			return null;
// 		}

// 		public string ProcessMarkup (string file)
// 		{
// 			foreach (var ext in MARKUP_LANGS) {
// 				if (Path.GetExtension (file).Equals ("." + ext)) {
// 					return markdown.Transform (File.ReadAllText (file));
// 				}
// 			}

// 			return null;
// 		}

func ProcessMetadata(file string) (map[string]interface{}, error) {
	r, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	res := map[string]interface{}{}
	if err := yaml.NewDecoder(r).Decode(&res); err != nil {
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
	files, err := filepath.Glob(filepath.Join(t.BaseDir, "*.yml"))
	if err != nil {
		return err
	}
	for _, i := range files {
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
