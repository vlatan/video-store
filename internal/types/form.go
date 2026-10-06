package types

type FieldType int

// Returns true if the field type is input
func (ft FieldType) IsInput() bool {
	return ft == FieldTypeInput
}

// Returns true if the field type is textarea
func (ft FieldType) IsTextarea() bool {
	return ft == FieldTypeTextarea
}

type FormGroup struct {
	Type        FieldType
	Label       string
	Placeholder string
	Value       string
	Pattern     string
	Title       string
}

type Form struct {
	Legend      string
	Title       *FormGroup
	Content     *FormGroup
	Category    *FormGroup
	Directors   []*FormGroup
	ReleaseYear *FormGroup
	Error       *FlashMessage
}
