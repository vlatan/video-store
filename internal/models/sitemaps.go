package models

type SitemapItem struct {
	Type         string
	Location     string
	LastModified string
}

type SitemapPart struct {
	Entries      []*SitemapItem
	Location     string
	LastModified string
}

type Sitemap map[string]SitemapPart
