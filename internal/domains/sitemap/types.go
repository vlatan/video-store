package sitemap

import "encoding/json"

type Item struct {
	Type         string
	BucketId     int
	Location     string
	LastModified string
}

// ============================================================================= //

type Part struct {
	Entries      []*Item
	Location     string
	LastModified string
}

// MarshalBinary implements the encoding.BinaryMarshaler interface
func (p Part) MarshalBinary() (data []byte, err error) {
	return json.Marshal(p)
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface
func (p *Part) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, p)
}

// ============================================================================= //

type Index map[string]*Part
