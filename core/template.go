package core

import (
	"io"
	"strings"

	"github.com/cbroglie/mustache"
)

type Template struct {
	*mustache.Template
}

func NewTemplate(raw []byte) (*Template, error) {
	tpl, err := mustache.ParseString(string(raw))
	if err != nil {
		return nil, err
	}

	return &Template{tpl}, nil
}

func (t *Template) Render(page *Page, w io.Writer) error {
	data := map[string]interface{}{}

	for k, v := range page.Tacker.Metadata {
		data[k] = v
	}

	for k, v := range page.Variables {
		data[k] = v
	}

	data["permalink"] = page.Permalink()
	data["slug"] = page.Name
	data["name"] = strings.Replace(strings.ToTitle(page.Name), "-", " ", -1)
	data["parent"] = page.Parent
	data["siblings"] = page.Siblings
	data["navigation"] = page.Tacker.Navigation
	data["ancestors"] = page.Ancestors()
	//data["current"] = ctx != null && ctx.Page == this;

	// for k, v := range data {
	// 	fmt.Printf("%20s = %+v\n", k, v)
	// }

	str, err := t.Template.Render(data)
	if err != nil {
		return err
	}
	if _, err := io.WriteString(w, str); err != nil {
		return err
	}

	return nil
}
