package form

type FieldType int

const ( // Form fieled types
	FieldTypeInput FieldType = iota
	FieldTypeTextarea
)

// Returns true if the field type is input
func (ft FieldType) IsInput() bool {
	return ft == FieldTypeInput
}

// Returns true if the field type is textarea
func (ft FieldType) IsTextarea() bool {
	return ft == FieldTypeTextarea
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
