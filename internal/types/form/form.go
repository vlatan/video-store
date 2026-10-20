package form

type FieldType int

const ( // Form fieled types
	Input FieldType = iota
	Textarea
)

// Returns true if the field type is input
func (ft FieldType) IsInput() bool {
	return ft == Input
}

// Returns true if the field type is textarea
func (ft FieldType) IsTextarea() bool {
	return ft == Textarea
}

// ============================================================================= //

type Group struct {
	Type        FieldType
	Label       string
	Placeholder string
	Value       string
	Pattern     string
	Title       string
}
