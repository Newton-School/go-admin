package core

import (
	"context"
	"fmt"
	"net/url"
	"reflect"
	"strings"
)

type urlValues = url.Values

type untypedPage struct {
	Items   []any `json:"items"`
	Total   int   `json:"total"`
	Page    int   `json:"page"`
	PerPage int   `json:"per_page"`
}

func (r *typedResource[T, ID]) fields() []Field {
	return append([]Field(nil), r.resource.Fields...)
}

func (r *typedResource[T, ID]) listConfig() ListConfig {
	return r.resource.List
}

func (r *typedResource[T, ID]) fieldsets() []Fieldset {
	return append([]Fieldset(nil), r.resource.Fieldsets...)
}

func (r *typedResource[T, ID]) list(ctx context.Context, query Query) (untypedPage, error) {
	page, err := r.resource.Repository.List(ctx, query)
	if err != nil {
		return untypedPage{}, err
	}
	items := make([]any, 0, len(page.Items))
	for _, item := range page.Items {
		items = append(items, item)
	}
	return untypedPage{
		Items:   items,
		Total:   page.Total,
		Page:    page.Page,
		PerPage: page.PerPage,
	}, nil
}

func (r *typedResource[T, ID]) get(ctx context.Context, rawID string) (any, error) {
	id, err := r.resource.ID.Parse(rawID)
	if err != nil {
		return nil, err
	}
	return r.resource.Repository.Get(ctx, id)
}

func (r *typedResource[T, ID]) create(ctx context.Context, values urlValues) (any, ValidationErrors, error) {
	var obj T
	errs := BindForm(r.resource.Fields, values, &obj)
	if !errs.Empty() {
		return obj, errs, nil
	}
	created, err := r.resource.Repository.Create(ctx, obj)
	return created, nil, err
}

func (r *typedResource[T, ID]) update(ctx context.Context, rawID string, values urlValues) (any, ValidationErrors, error) {
	id, err := r.resource.ID.Parse(rawID)
	if err != nil {
		return nil, nil, err
	}
	obj, err := r.resource.Repository.Get(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	errs := BindForm(r.resource.Fields, values, &obj)
	if !errs.Empty() {
		return obj, errs, nil
	}
	updated, err := r.resource.Repository.Update(ctx, id, obj)
	return updated, nil, err
}

func (r *typedResource[T, ID]) createJSON(ctx context.Context, values map[string]any) (any, ValidationErrors, error) {
	var obj T
	errs := BindJSON(r.resource.Fields, values, &obj, false)
	if !errs.Empty() {
		return obj, errs, nil
	}
	created, err := r.resource.Repository.Create(ctx, obj)
	return created, nil, err
}

func (r *typedResource[T, ID]) updateJSON(ctx context.Context, rawID string, values map[string]any) (any, ValidationErrors, error) {
	id, err := r.resource.ID.Parse(rawID)
	if err != nil {
		return nil, nil, err
	}
	obj, err := r.resource.Repository.Get(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	errs := BindJSON(r.resource.Fields, values, &obj, true)
	if !errs.Empty() {
		return obj, errs, nil
	}
	updated, err := r.resource.Repository.Update(ctx, id, obj)
	return updated, nil, err
}

func (r *typedResource[T, ID]) delete(ctx context.Context, rawID string) error {
	id, err := r.resource.ID.Parse(rawID)
	if err != nil {
		return err
	}
	return r.resource.Repository.Delete(ctx, id)
}

func (r *typedResource[T, ID]) actions() []ActionMeta {
	actions := make([]ActionMeta, 0, len(r.resource.Actions))
	for _, action := range r.resource.Actions {
		actions = append(actions, ActionMeta{
			Name:        action.Name,
			Label:       displayLabel(action.Name, action.Label),
			Description: action.Description,
			Confirm:     action.Confirm,
		})
	}
	return actions
}

func (r *typedResource[T, ID]) runAction(ctx context.Context, name string, rawIDs []string) (ActionResult, error) {
	for _, action := range r.resource.Actions {
		if action.Name != name {
			continue
		}
		ids := make([]ID, 0, len(rawIDs))
		objects := make([]T, 0, len(rawIDs))
		for _, rawID := range rawIDs {
			id, err := r.resource.ID.Parse(rawID)
			if err != nil {
				return ActionResult{}, err
			}
			obj, err := r.resource.Repository.Get(ctx, id)
			if err != nil {
				return ActionResult{}, err
			}
			ids = append(ids, id)
			objects = append(objects, obj)
		}
		if action.Run == nil {
			return ActionResult{}, fmt.Errorf("action %q has no runner", name)
		}
		return action.Run(ctx, ActionRequest[T, ID]{
			Resource: r.resource,
			IDs:      ids,
			Objects:  objects,
		})
	}
	return ActionResult{}, ErrNotFound
}

func (r *typedResource[T, ID]) lookup(ctx context.Context, fieldName string, query Query) ([]Choice, error) {
	page, err := r.resource.Repository.List(ctx, query)
	if err != nil {
		return nil, err
	}
	choices := make([]Choice, 0, len(page.Items))
	for _, item := range page.Items {
		choices = append(choices, Choice{
			Value: r.idString(item),
			Label: formatValue(r.fieldValue(item, fieldName)),
		})
	}
	return choices, nil
}

func (r *typedResource[T, ID]) idString(obj any) string {
	var value any
	var ok bool
	if r.resource.IDField != "" {
		value, ok = readFieldValue(obj, r.resource.IDField)
	}
	if !ok {
		value, ok = readFieldValue(obj, "id")
	}
	if !ok {
		value, ok = readFieldValue(obj, "ID")
	}
	if !ok {
		return ""
	}
	typed, ok := value.(ID)
	if ok {
		return r.resource.ID.Format(typed)
	}
	converted, err := convertID[ID](value)
	if err != nil {
		return fmt.Sprint(value)
	}
	return r.resource.ID.Format(converted)
}

func (r *typedResource[T, ID]) fieldValue(obj any, fieldName string) any {
	value, _ := readFieldValue(obj, fieldName)
	return value
}

func convertID[ID comparable](value any) (ID, error) {
	var zero ID
	targetType := reflect.TypeOf(zero)
	if targetType == nil {
		return zero, fmt.Errorf("nil id type")
	}
	rv := reflect.ValueOf(value)
	if !rv.IsValid() {
		return zero, fmt.Errorf("invalid id")
	}
	if rv.Type().AssignableTo(targetType) {
		return rv.Interface().(ID), nil
	}
	if rv.Type().ConvertibleTo(targetType) {
		return rv.Convert(targetType).Interface().(ID), nil
	}
	return zero, fmt.Errorf("cannot convert id")
}

func readFieldValue(obj any, name string) (any, bool) {
	value := reflect.ValueOf(obj)
	if !value.IsValid() {
		return nil, false
	}
	if value.Kind() == reflect.Pointer {
		if value.IsNil() {
			return nil, false
		}
		value = value.Elem()
	}
	if value.Kind() != reflect.Struct {
		return nil, false
	}
	targetType := value.Type()
	for i := 0; i < value.NumField(); i++ {
		structField := targetType.Field(i)
		if structField.PkgPath != "" {
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
				return value.Field(i).Interface(), true
			}
		}
	}
	return nil, false
}
