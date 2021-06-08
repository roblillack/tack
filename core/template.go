package core

import (
	"io"
	"strings"

	"github.com/cbroglie/mustache"
)

type Template struct {
	*mustache.Template
}

func PageValues(p *Page, ctx *Page) map[string]interface{} {
	if p == nil {
		return nil
	}

	data := map[string]interface{}{}

	data["name"] = strings.Replace(strings.Title(p.Name), "-", " ", -1)

	for k, v := range p.Variables {
		data[k] = v
	}

	data["permalink"] = p.Permalink()
	data["slug"] = p.Name
	data["current"] = ctx != nil && ctx == p

	return data
}

func PageListValues(pages []*Page, ctx *Page) []map[string]interface{} {
	r := []map[string]interface{}{}
	for _, i := range pages {
		r = append(r, PageValues(i, ctx))
	}
	return r
}

func (t *Template) Render(page *Page, w io.Writer) error {
	ctx := map[string]interface{}{}

	for k, v := range page.Tacker.Metadata {
		ctx[k] = v
	}

	for k, v := range PageValues(page, page) {
		ctx[k] = v
	}

	ctx["parent"] = PageValues(page.Parent, page)
	ctx["siblings"] = PageListValues(page.Siblings, page)
	ctx["navigation"] = PageListValues(page.Tacker.Navigation, page)
	ctx["ancestors"] = PageListValues(page.Ancestors(), page)

	return t.Template.FRender(w, ctx)
}
