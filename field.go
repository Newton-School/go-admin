package admin

// FieldKind names a built-in field/widget family.
type FieldKind string

const (
	FieldKindText     FieldKind = "text"
	FieldKindTextarea FieldKind = "textarea"
	FieldKindInt64    FieldKind = "int64"
	FieldKindBool     FieldKind = "bool"
	FieldKindTime     FieldKind = "time"
)

// Field describes a model field for list display and forms.
type Field struct {
	NameValue     string
	LabelValue    string
	Kind          FieldKind
	ReadonlyValue bool
	RequiredValue bool
}

// Text creates a text input field.
func Text(name, label string) Field {
	return Field{NameValue: name, LabelValue: label, Kind: FieldKindText}
}

// Int64 creates a signed integer field.
func Int64(name, label string) Field {
	return Field{NameValue: name, LabelValue: label, Kind: FieldKindInt64}
}

// Bool creates a boolean field.
func Bool(name, label string) Field {
	return Field{NameValue: name, LabelValue: label, Kind: FieldKindBool}
}

// Time creates a date/time field.
func Time(name, label string) Field {
	return Field{NameValue: name, LabelValue: label, Kind: FieldKindTime}
}

// Readonly marks a field as display-only in admin forms.
func (f Field) Readonly() Field {
	f.ReadonlyValue = true
	return f
}

// Required marks a field as required in admin forms.
func (f Field) Required() Field {
	f.RequiredValue = true
	return f
}

func (f Field) name() string {
	return f.NameValue
}

func (f Field) label() string {
	return displayLabel(f.NameValue, f.LabelValue)
}
