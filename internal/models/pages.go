package models

import "html/template"

type Page struct {
	Slug        string
	Title       string
	Content     string
	HTMLContent template.HTML
}
