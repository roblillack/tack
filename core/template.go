package core

import (
	"io"
	"sort"
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

	// can be overridden from page variable
	data["name"] = p.Name

	for k, v := range p.Variables {
		data[k] = v
	}

	data["permalink"] = p.Permalink()
	data["slug"] = p.Slug
	data["current"] = ctx != nil && ctx == p
	data["root"] = p.Root()
	if p.Post() {
		data["date"] = p.Date.Format("2006-01-02")
		data["year"] = p.Date.Format("2006")
		data["month"] = p.Date.Format("January")
	}
	data["tags"] = TagList(p)

	return data
}

func PageListValues(pages []*Page, ctx *Page) []map[string]interface{} {
	r := []map[string]interface{}{}
	var year int
	var month string
	for idx, i := range pages {
		data := PageValues(i, ctx)
		data["first"] = idx == 0
		data["last"] = idx == len(pages)-1
		if i.Post() {
			var next *Page
			if idx+1 < len(pages) {
				next = pages[idx+1]
			}
			m := i.Date.Format("2006-January")
			data["first_in_year"] = year != i.Date.Year()
			data["first_in_month"] = month != m
			year = i.Date.Year()
			month = m

			data["last_in_year"] = next == nil || year != next.Date.Year()
			data["last_in_month"] = next == nil || month != next.Date.Format("2006-January")
		}
		r = append(r, data)
	}
	return r
}

func TagList(page *Page) []map[string]interface{} {
	list := []Tag{}

	if page.addTagPages {
		for n := range page.Tacker.Tags {
			list = append(list, page.Tacker.Tag(n))
		}
	} else if tags, ok := page.Variables["tags"].([]interface{}); ok {
		for _, i := range tags {
			s, ok := i.(string)
			if s == "" || !ok {
				continue
			}
			list = append(list, page.Tacker.Tag(s))
		}
	}

	sort.Slice(list, func(i, j int) bool {
		if list[i].Count == list[j].Count {
			return strings.Compare(list[i].Name, list[j].Name) < 0
		}

		return list[i].Count > list[j].Count
	})

	r := []map[string]interface{}{}
	for _, i := range list {
		r = append(r, map[string]interface{}{
			"name":      i.Name,
			"slug":      i.Slug,
			"count":     i.Count,
			"permalink": i.Permalink,
		})
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

	if !page.Post() && !page.addTagPages {
		posts := page.Posts
		if len(posts) == 0 && page.Parent != nil && len(page.Parent.Posts) > 0 {
			posts = page.Parent.Posts
		} else if len(posts) == 0 && page.Parent == nil {
			for _, i := range page.Tacker.Posts {
				if i.Parent == nil {
					posts = append(posts, i)
				}
			}
		}
		ctx["posts"] = PageListValues(limitPageList(posts, page, "posts_limit"), page)
	}

	return t.Template.FRender(w, ctx)
}

func limitPageList(list []*Page, page *Page, name string) []*Page {
	v, ok := page.Variables[name].(int)
	if !ok || v < 1 || v > len(list) {
		return list
	}

	return list[0:v]
}
