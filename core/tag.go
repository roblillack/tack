package core

import "strings"

type Tag struct {
	Name      string
	Slug      string
	Count     int
	Permalink string
}

func TagSlug(name string) string {
	return strings.Replace(strings.ToLower(name), " ", "-", -1)
}
