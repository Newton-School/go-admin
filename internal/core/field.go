package core

import (
	"encoding/json"
	"fmt"
	"html"
	"html/template"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unicode"
)

// FieldKind names a built-in field/widget family.
type FieldKind string

const (
	FieldKindText        FieldKind = "text"
	FieldKindTextarea    FieldKind = "textarea"
	FieldKindInt         FieldKind = "int"
	FieldKindInt64       FieldKind = "int64"
	FieldKindFloat       FieldKind = "float"
	FieldKindBool        FieldKind = "bool"
	FieldKindDate        FieldKind = "date"
	FieldKindDateTime    FieldKind = "datetime"
	FieldKindJSON        FieldKind = "json"
	FieldKindEnum        FieldKind = "enum"
	FieldKindSelect      FieldKind = "select"
	FieldKindMultiSelect FieldKind = "multiselect"
	FieldKindRelation    FieldKind = "relation"
)

// Field describes a model field for list display and forms.
type Field struct {
	NameValue        string
	LabelValue       string
	Kind             FieldKind
	ReadonlyValue    bool
	RequiredValue    bool
	HelpValue        string
	PlaceholderValue string
	ChoicesValue     []Choice
}

// Text creates a text input field.
func Text(name, label string) Field {
	return Field{NameValue: name, LabelValue: label, Kind: FieldKindText}
}

// Textarea creates a textarea field.
func Textarea(name, label string) Field {
	return Field{NameValue: name, LabelValue: label, Kind: FieldKindTextarea}
}

// Int creates a signed integer field.
func Int(name, label string) Field {
	return Field{NameValue: name, LabelValue: label, Kind: FieldKindInt}
}

// Int64 creates a signed integer field.
func Int64(name, label string) Field {
	return Field{NameValue: name, LabelValue: label, Kind: FieldKindInt64}
}

// Float creates a floating-point field.
func Float(name, label string) Field {
	return Field{NameValue: name, LabelValue: label, Kind: FieldKindFloat}
}

// Bool creates a boolean field.
func Bool(name, label string) Field {
	return Field{NameValue: name, LabelValue: label, Kind: FieldKindBool}
}

// Date creates a date field.
func Date(name, label string) Field {
	return Field{NameValue: name, LabelValue: label, Kind: FieldKindDate}
}

// DateTime creates a date/time field.
func DateTime(name, label string) Field {
	return Field{NameValue: name, LabelValue: label, Kind: FieldKindDateTime}
}

// Time is kept as a convenience alias for DateTime.
func Time(name, label string) Field {
	return DateTime(name, label)
}

// JSON creates a JSON textarea field.
func JSON(name, label string) Field {
	return Field{NameValue: name, LabelValue: label, Kind: FieldKindJSON}
}

// Enum creates a single-choice field.
func Enum(name, label string) Field {
	return Field{NameValue: name, LabelValue: label, Kind: FieldKindEnum}
}

// Select creates a single-choice select field.
func Select(name, label string) Field {
	return Field{NameValue: name, LabelValue: label, Kind: FieldKindSelect}
}

// MultiSelect creates a multi-choice select field.
func MultiSelect(name, label string) Field {
	return Field{NameValue: name, LabelValue: label, Kind: FieldKindMultiSelect}
}

// Relation creates an async relation lookup field.
func Relation(name, label string) Field {
	return Field{NameValue: name, LabelValue: label, Kind: FieldKindRelation}
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

// Help attaches helper text to the field.
func (f Field) Help(help string) Field {
	f.HelpValue = help
	return f
}

// Placeholder attaches placeholder text to input-like widgets.
func (f Field) Placeholder(placeholder string) Field {
	f.PlaceholderValue = placeholder
	return f
}

// Options attaches choices to enum/select widgets.
func (f Field) Options(choices []Choice) Field {
	f.ChoicesValue = append([]Choice(nil), choices...)
	return f
}

func (f Field) name() string {
	return f.NameValue
}

func (f Field) label() string {
	return displayLabel(f.NameValue, f.LabelValue)
}

// ValidationErrors stores user-facing field validation failures.
type ValidationErrors map[string]string

// Empty reports whether there are any validation failures.
func (v ValidationErrors) Empty() bool {
	return len(v) == 0
}

// Get returns the message for a field.
func (v ValidationErrors) Get(field string) string {
	return v[field]
}

func (v ValidationErrors) add(field, message string) {
	if _, exists := v[field]; !exists {
		v[field] = message
	}
}

// BindForm applies form values to a pointer to a struct using admin fields.
func BindForm(fields []Field, values url.Values, dst any) ValidationErrors {
	errs := ValidationErrors{}
	target, ok := targetStruct(dst)
	if !ok {
		errs.add("_", "target must be a pointer to a struct")
		return errs
	}

	for _, field := range fields {
		if field.ReadonlyValue {
			continue
		}
		structField, ok := findStructField(target, field.NameValue)
		if !ok {
			continue
		}
		rawValues := values[field.NameValue]
		if field.Kind == FieldKindBool && len(rawValues) == 0 {
			rawValues = []string{"false"}
		}
		if field.RequiredValue && isEmptyInput(rawValues) {
			errs.add(field.NameValue, fmt.Sprintf("%s is required", field.label()))
			continue
		}
		if isEmptyInput(rawValues) {
			continue
		}
		if err := setFieldValue(structField, field, rawValues); err != nil {
			errs.add(field.NameValue, err.Error())
		}
	}

	return errs
}

// BindJSON applies JSON object values to a pointer to a struct using admin fields.
func BindJSON(fields []Field, values map[string]any, dst any, partial bool) ValidationErrors {
	errs := ValidationErrors{}
	target, ok := targetStruct(dst)
	if !ok {
		errs.add("_", "target must be a pointer to a struct")
		return errs
	}

	for _, field := range fields {
		if field.ReadonlyValue {
			continue
		}
		value, exists := values[field.NameValue]
		if !exists {
			if field.RequiredValue && !partial {
				errs.add(field.NameValue, fmt.Sprintf("%s is required", field.label()))
			}
			continue
		}
		if field.RequiredValue && isEmptyJSONValue(value) {
			errs.add(field.NameValue, fmt.Sprintf("%s is required", field.label()))
			continue
		}
		structField, ok := findStructField(target, field.NameValue)
		if !ok {
			continue
		}
		rawValues, err := jsonValueStrings(value, field)
		if err != nil {
			errs.add(field.NameValue, err.Error())
			continue
		}
		if isEmptyInput(rawValues) {
			continue
		}
		if err := setFieldValue(structField, field, rawValues); err != nil {
			errs.add(field.NameValue, err.Error())
		}
	}

	return errs
}

func isEmptyJSONValue(value any) bool {
	if value == nil {
		return true
	}
	switch typed := value.(type) {
	case string:
		return strings.TrimSpace(typed) == ""
	case []any:
		return len(typed) == 0
	default:
		return false
	}
}

func jsonValueStrings(value any, field Field) ([]string, error) {
	if field.Kind == FieldKindJSON {
		encoded, err := json.Marshal(value)
		if err != nil {
			return nil, fmt.Errorf("%s must be valid JSON", field.label())
		}
		return []string{string(encoded)}, nil
	}
	if field.Kind == FieldKindMultiSelect {
		switch typed := value.(type) {
		case []any:
			values := make([]string, 0, len(typed))
			for _, item := range typed {
				values = append(values, jsonScalarString(item))
			}
			return values, nil
		case []string:
			return append([]string(nil), typed...), nil
		default:
			return []string{jsonScalarString(value)}, nil
		}
	}
	return []string{jsonScalarString(value)}, nil
}

func jsonScalarString(value any) string {
	switch typed := value.(type) {
	case string:
		return typed
	case float64:
		return strconv.FormatFloat(typed, 'f', -1, 64)
	case bool:
		if typed {
			return "true"
		}
		return "false"
	default:
		return fmt.Sprint(value)
	}
}

func targetStruct(dst any) (reflect.Value, bool) {
	value := reflect.ValueOf(dst)
	if !value.IsValid() || value.Kind() != reflect.Pointer || value.IsNil() {
		return reflect.Value{}, false
	}
	value = value.Elem()
	if value.Kind() != reflect.Struct {
		return reflect.Value{}, false
	}
	return value, true
}

func findStructField(target reflect.Value, name string) (reflect.Value, bool) {
	targetType := target.Type()
	for i := 0; i < target.NumField(); i++ {
		structField := targetType.Field(i)
		value := target.Field(i)
		if structField.PkgPath != "" || !value.CanSet() {
			continue
		}
		candidates := []string{
			structField.Name,
			strings.ToLower(structField.Name),
			toSnakeCase(structField.Name),
			tagName(structField.Tag.Get("json")),
			tagName(structField.Tag.Get("db")),
			tagName(structField.Tag.Get("form")),
		}
		for _, candidate := range candidates {
			if candidate == name {
				return value, true
			}
		}
	}
	return reflect.Value{}, false
}

func tagName(tag string) string {
	name, _, _ := strings.Cut(tag, ",")
	if name == "-" {
		return ""
	}
	return name
}

func toSnakeCase(value string) string {
	var b strings.Builder
	for i, r := range value {
		if unicode.IsUpper(r) {
			if i > 0 {
				b.WriteByte('_')
			}
			b.WriteRune(unicode.ToLower(r))
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

func isEmptyInput(values []string) bool {
	if len(values) == 0 {
		return true
	}
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return false
		}
	}
	return true
}

func setFieldValue(dst reflect.Value, field Field, values []string) error {
	switch field.Kind {
	case FieldKindBool:
		return setBool(dst, firstValue(values))
	case FieldKindInt, FieldKindInt64:
		return setInt(dst, firstValue(values), field.label())
	case FieldKindFloat:
		return setFloat(dst, firstValue(values), field.label())
	case FieldKindDate:
		parsed, err := parseDate(firstValue(values))
		if err != nil {
			return fmt.Errorf("%s must be a valid date", field.label())
		}
		return setTime(dst, parsed)
	case FieldKindDateTime:
		parsed, err := parseDateTime(firstValue(values))
		if err != nil {
			return fmt.Errorf("%s must be a valid date/time", field.label())
		}
		return setTime(dst, parsed)
	case FieldKindJSON:
		return setJSON(dst, firstValue(values), field.label())
	case FieldKindMultiSelect:
		return setSlice(dst, values, field)
	default:
		if err := validateChoice(field, firstValue(values)); err != nil {
			return err
		}
		return setScalarFromString(dst, firstValue(values), field.label())
	}
}

func firstValue(values []string) string {
	if len(values) == 0 {
		return ""
	}
	return values[0]
}

func setBool(dst reflect.Value, raw string) error {
	value := false
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "1", "true", "on", "yes":
		value = true
	case "", "0", "false", "off", "no":
		value = false
	default:
		return fmt.Errorf("must be a boolean")
	}
	if dst.Kind() != reflect.Bool {
		return fmt.Errorf("target field must be bool")
	}
	dst.SetBool(value)
	return nil
}

func setInt(dst reflect.Value, raw, label string) error {
	parsed, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
	if err != nil {
		return fmt.Errorf("%s must be a whole number", label)
	}
	return setIntValue(dst, parsed, label)
}

func setIntValue(dst reflect.Value, value int64, label string) error {
	switch dst.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if dst.OverflowInt(value) {
			return fmt.Errorf("%s is out of range", label)
		}
		dst.SetInt(value)
		return nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if value < 0 || dst.OverflowUint(uint64(value)) {
			return fmt.Errorf("%s is out of range", label)
		}
		dst.SetUint(uint64(value))
		return nil
	default:
		return fmt.Errorf("target field must be integer")
	}
}

func setFloat(dst reflect.Value, raw, label string) error {
	parsed, err := strconv.ParseFloat(strings.TrimSpace(raw), 64)
	if err != nil {
		return fmt.Errorf("%s must be a number", label)
	}
	switch dst.Kind() {
	case reflect.Float32, reflect.Float64:
		if dst.OverflowFloat(parsed) {
			return fmt.Errorf("%s is out of range", label)
		}
		dst.SetFloat(parsed)
		return nil
	default:
		return fmt.Errorf("target field must be float")
	}
}

func parseDate(raw string) (time.Time, error) {
	return time.Parse("2006-01-02", strings.TrimSpace(raw))
}

func parseDateTime(raw string) (time.Time, error) {
	raw = strings.TrimSpace(raw)
	for _, layout := range []string{time.RFC3339, "2006-01-02T15:04", "2006-01-02 15:04:05"} {
		parsed, err := time.Parse(layout, raw)
		if err == nil {
			return parsed, nil
		}
	}
	return time.Time{}, fmt.Errorf("invalid date/time")
}

func setTime(dst reflect.Value, value time.Time) error {
	if dst.Type() != reflect.TypeOf(time.Time{}) {
		return fmt.Errorf("target field must be time.Time")
	}
	dst.Set(reflect.ValueOf(value))
	return nil
}

func setJSON(dst reflect.Value, raw, label string) error {
	if dst.Kind() == reflect.String {
		dst.SetString(raw)
		return nil
	}
	target := reflect.New(dst.Type())
	if err := json.Unmarshal([]byte(raw), target.Interface()); err != nil {
		return fmt.Errorf("%s must be valid JSON", label)
	}
	dst.Set(target.Elem())
	return nil
}

func setSlice(dst reflect.Value, values []string, field Field) error {
	if dst.Kind() != reflect.Slice {
		return fmt.Errorf("target field must be a slice")
	}
	slice := reflect.MakeSlice(dst.Type(), 0, len(values))
	for _, value := range values {
		if err := validateChoice(field, value); err != nil {
			return err
		}
		item := reflect.New(dst.Type().Elem()).Elem()
		if err := setScalarFromString(item, value, field.label()); err != nil {
			return err
		}
		slice = reflect.Append(slice, item)
	}
	dst.Set(slice)
	return nil
}

func validateChoice(field Field, value string) error {
	if len(field.ChoicesValue) == 0 {
		return nil
	}
	for _, choice := range field.ChoicesValue {
		if choice.Value == value {
			return nil
		}
	}
	return fmt.Errorf("%s has an invalid choice", field.label())
}

func setScalarFromString(dst reflect.Value, value, label string) error {
	switch dst.Kind() {
	case reflect.String:
		dst.SetString(value)
		return nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		parsed, err := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
		if err != nil {
			return fmt.Errorf("%s must be a whole number", label)
		}
		return setIntValue(dst, parsed, label)
	default:
		return fmt.Errorf("target field must be string or integer")
	}
}

// WidgetContext is passed to RenderWidget.
type WidgetContext struct {
	Field  Field
	Value  any
	Errors ValidationErrors
}

// RenderWidget renders a safe built-in HTML form control.
func RenderWidget(ctx WidgetContext) template.HTML {
	field := ctx.Field
	name := html.EscapeString(field.NameValue)
	label := html.EscapeString(field.label())
	value := html.EscapeString(formatWidgetValue(field, ctx.Value))
	placeholder := html.EscapeString(field.PlaceholderValue)

	var b strings.Builder
	b.WriteString(`<div class="form-row field-`)
	b.WriteString(name)
	b.WriteString(`"><div>`)
	b.WriteString(`<label for="`)
	b.WriteString(`id_`)
	b.WriteString(name)
	b.WriteString(`">`)
	b.WriteString(label)
	if field.RequiredValue {
		b.WriteString(` <span aria-hidden="true">*</span>`)
	}
	b.WriteString(`</label>`)

	switch field.Kind {
	case FieldKindTextarea, FieldKindJSON:
		b.WriteString(`<textarea id="id_`)
		b.WriteString(name)
		b.WriteString(`" name="`)
		b.WriteString(name)
		b.WriteString(`" placeholder="`)
		b.WriteString(placeholder)
		b.WriteString(`">`)
		b.WriteString(value)
		b.WriteString(`</textarea>`)
	case FieldKindBool:
		b.WriteString(`<input type="checkbox" id="id_`)
		b.WriteString(name)
		b.WriteString(`" name="`)
		b.WriteString(name)
		b.WriteString(`" value="true"`)
		if ctx.Value == true || strings.EqualFold(fmt.Sprint(ctx.Value), "true") {
			b.WriteString(` checked`)
		}
		b.WriteString(`>`)
	case FieldKindEnum, FieldKindSelect, FieldKindRelation:
		renderSelect(&b, name, value, field.ChoicesValue, false)
	case FieldKindMultiSelect:
		renderSelect(&b, name, value, field.ChoicesValue, true)
	case FieldKindDate:
		renderInput(&b, "date", name, value, placeholder)
	case FieldKindDateTime:
		renderInput(&b, "datetime-local", name, value, placeholder)
	case FieldKindInt, FieldKindInt64:
		renderInput(&b, "number", name, value, placeholder)
	case FieldKindFloat:
		renderInput(&b, "number", name, value, placeholder)
	default:
		renderInput(&b, "text", name, value, placeholder)
	}

	if field.HelpValue != "" {
		b.WriteString(`<div class="help">`)
		b.WriteString(html.EscapeString(field.HelpValue))
		b.WriteString(`</div>`)
	}
	if ctx.Errors != nil {
		if message := ctx.Errors.Get(field.NameValue); message != "" {
			b.WriteString(`<div class="errornote">`)
			b.WriteString(html.EscapeString(message))
			b.WriteString(`</div>`)
		}
	}
	b.WriteString(`</div></div>`)

	return template.HTML(b.String())
}

func renderInput(b *strings.Builder, inputType, name, value, placeholder string) {
	b.WriteString(`<input type="`)
	b.WriteString(inputType)
	b.WriteString(`" id="id_`)
	b.WriteString(name)
	b.WriteString(`" name="`)
	b.WriteString(name)
	b.WriteString(`" value="`)
	b.WriteString(value)
	b.WriteString(`" placeholder="`)
	b.WriteString(placeholder)
	b.WriteString(`">`)
}

func renderSelect(b *strings.Builder, name, value string, choices []Choice, multiple bool) {
	b.WriteString(`<select id="id_`)
	b.WriteString(name)
	b.WriteString(`" name="`)
	b.WriteString(name)
	b.WriteString(`"`)
	if multiple {
		b.WriteString(` multiple`)
	}
	b.WriteString(`>`)
	for _, choice := range choices {
		choiceValue := html.EscapeString(choice.Value)
		b.WriteString(`<option value="`)
		b.WriteString(choiceValue)
		b.WriteString(`"`)
		if choiceValue == value || strings.Contains(","+value+",", ","+choiceValue+",") {
			b.WriteString(` selected`)
		}
		b.WriteString(`>`)
		b.WriteString(html.EscapeString(choice.Label))
		b.WriteString(`</option>`)
	}
	b.WriteString(`</select>`)
}

func formatWidgetValue(field Field, value any) string {
	if value == nil {
		return ""
	}
	if parsed, ok := value.(time.Time); ok {
		if parsed.IsZero() {
			return ""
		}
		if field.Kind == FieldKindDate {
			return parsed.Format("2006-01-02")
		}
		return parsed.Format("2006-01-02T15:04")
	}
	if field.Kind == FieldKindJSON {
		encoded, err := json.MarshalIndent(value, "", "  ")
		if err == nil {
			return string(encoded)
		}
	}
	rv := reflect.ValueOf(value)
	if rv.IsValid() && rv.Kind() == reflect.Slice {
		parts := make([]string, 0, rv.Len())
		for i := 0; i < rv.Len(); i++ {
			parts = append(parts, fmt.Sprint(rv.Index(i).Interface()))
		}
		return strings.Join(parts, ",")
	}
	return fmt.Sprint(value)
}
