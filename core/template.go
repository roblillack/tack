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

	return map[string]interface{}{
		"permalink": p.Permalink(),
		"slug":      p.Name,
		"name":      strings.Replace(strings.ToTitle(p.Name), "-", " ", -1),
		"current":   ctx != nil && ctx == p,
	}
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

	for k, v := range page.Variables {
		ctx[k] = v
	}

	for k, v := range PageValues(page, page) {
		ctx[k] = v
	}

	ctx["parent"] = PageValues(page.Parent, page)
	ctx["siblings"] = PageListValues(page.Siblings, page)
	ctx["navigation"] = PageListValues(page.Tacker.Navigation, page)
	ctx["ancestors"] = PageListValues(page.Ancestors(), page)

	str, err := t.Template.Render(ctx)
	if err != nil {
		return err
	}
	if _, err := io.WriteString(w, str); err != nil {
		return err
	}

	return nil
}
