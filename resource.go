package admin

import (
	"context"
	"fmt"
	"regexp"
)

var slugPattern = regexp.MustCompile(`^[a-z][a-z0-9_/-]*$`)

// Repository is the ORM-neutral persistence contract for one admin resource.
type Repository[T any, ID comparable] interface {
	List(context.Context, Query) (Page[T], error)
	Get(context.Context, ID) (T, error)
	Create(context.Context, T) (T, error)
	Update(context.Context, ID, T) (T, error)
	Delete(context.Context, ID) error
}

// Resource describes one model-like admin resource.
type Resource[T any, ID comparable] struct {
	Name       string
	Label      string
	Repository Repository[T, ID]
	ID         IDCodec[ID]
	Fields     []Field
	List       ListConfig
	Fieldsets  []Fieldset
	Actions    []Action[T, ID]
}

// ListConfig controls the changelist page and matching API list endpoint.
type ListConfig struct {
	Columns []string
	Search  []string
	Sort    []SortField
	Filters []Filter
}

// Fieldset controls grouping on create and change pages.
type Fieldset struct {
	Title       string
	Description string
	Fields      []string
	Rows        [][]string
	Collapsed   bool
}

// Filter describes one list filter.
type Filter struct {
	Name    string
	Label   string
	Choices []Choice
}

// Choice is used by enum, select, and filter controls.
type Choice struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

// Action describes a bulk or object-level admin operation.
type Action[T any, ID comparable] struct {
	Name        string
	Label       string
	Description string
	Confirm     bool
	Run         func(context.Context, ActionRequest[T, ID]) (ActionResult, error)
}

// ActionRequest is passed to custom actions.
type ActionRequest[T any, ID comparable] struct {
	Resource Resource[T, ID]
	IDs      []ID
	Objects  []T
}

// ActionResult is returned by custom actions.
type ActionResult struct {
	Message string `json:"message"`
}

type resourceMeta struct {
	Name  string
	Label string
}

type resourceRuntime interface {
	meta() resourceMeta
	fields() []Field
	listConfig() ListConfig
	fieldsets() []Fieldset
	list(context.Context, Query) (untypedPage, error)
	get(context.Context, string) (any, error)
	create(context.Context, urlValues) (any, ValidationErrors, error)
	update(context.Context, string, urlValues) (any, ValidationErrors, error)
	delete(context.Context, string) error
	idString(any) string
	fieldValue(any, string) any
}

func validateResource[T any, ID comparable](resource Resource[T, ID]) error {
	if !validSlug(resource.Name) {
		return fmt.Errorf("invalid resource name %q", resource.Name)
	}
	if resource.Repository == nil {
		return fmt.Errorf("resource %q is missing a repository", resource.Name)
	}
	if resource.ID == nil {
		return fmt.Errorf("resource %q is missing an id codec", resource.Name)
	}
	return nil
}

func validSlug(value string) bool {
	return slugPattern.MatchString(value)
}

func displayLabel(name, label string) string {
	if label != "" {
		return label
	}
	return name
}
