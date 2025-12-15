package models

import "encoding/json"

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

// MarshalBinary implements the encoding.BinaryMarshaler interface
func (s SitemapPart) MarshalBinary() (data []byte, err error) {
	return json.Marshal(s)
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface
func (s *SitemapPart) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, s)
}

type SitemapIndex map[string]*SitemapPart
