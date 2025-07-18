package models

import "time"

type SitemapItem struct {
	Type         string
	Location     string
	LastModified *time.Time
}

type SitemapPart struct {
	Entries      []SitemapItem
	LastModified *time.Time
}

type Sitemap map[string]SitemapPart
