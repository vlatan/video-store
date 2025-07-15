package models

import "html/template"

type Page struct {
	Slug            string
	Title           string
	MarkdownContent string
	HTMLContent     template.HTML
}
