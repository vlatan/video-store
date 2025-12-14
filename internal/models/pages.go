package models

import (
	"encoding/json"
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

// MarshalBinary implements the encoding.BinaryMarshaler interface
func (p Page) MarshalBinary() (data []byte, err error) {
	return json.Marshal(p)
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface
func (p *Page) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, p)
}
