package models

import (
	"html/template"
	"time"
)

type Page struct {
	Slug        string        `json:"slug,omitempty"`
	Title       string        `json:"title,omitempty"`
	Content     string        `json:"content,omitempty"`
	HTMLContent template.HTML `json:"html_content,omitempty"`
	CreatedAt   *time.Time    `json:"created_at,omitempty"`
	UpdatedAt   *time.Time    `json:"updated_at,omitempty"`
}
