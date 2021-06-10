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
	if !p.Date.IsZero() {
		data["date"] = p.Date.Format("2006-01-02")
	}

	return data
}

func PageListValues(pages []*Page, ctx *Page) []map[string]interface{} {
	r := []map[string]interface{}{}
	for idx, i := range pages {
		data := PageValues(i, ctx)
		data["first"] = idx == 0
		data["last"] = idx == len(pages)-1
		r = append(r, data)
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
	ctx["siblings"] = PageListValues(page.Siblings(), page)
	ctx["children"] = PageListValues(page.Children, page)
	ctx["navigation"] = PageListValues(page.Tacker.Navigation, page)
	ctx["menu"] = PageListValues(page.SiblingsAndMe, page)
	ctx["ancestors"] = PageListValues(page.Ancestors(), page)
	ctx["posts"] = PageListValues(page.Posts, page)

	return t.Template.FRender(w, ctx)
}
