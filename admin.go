package admin

import (
	"html/template"
	"net/http"
	"net/url"

	"github.com/ns/go-admin/internal/core"
)

type SiteConfig = core.SiteConfig
type Site = core.Site
type App = core.App
type AppMeta = core.AppMeta
type ResourceMeta = core.ResourceMeta

type IDCodec[ID comparable] = core.IDCodec[ID]
type Repository[T any, ID comparable] = core.Repository[T, ID]
type Resource[T any, ID comparable] = core.Resource[T, ID]
type ListConfig = core.ListConfig
type Fieldset = core.Fieldset
type Filter = core.Filter
type Choice = core.Choice
type Action[T any, ID comparable] = core.Action[T, ID]
type ActionRequest[T any, ID comparable] = core.ActionRequest[T, ID]
type ActionResult = core.ActionResult
type ActionMeta = core.ActionMeta

type FieldKind = core.FieldKind
type Field = core.Field
type ValidationErrors = core.ValidationErrors
type WidgetContext = core.WidgetContext

type SortField = core.SortField
type Query = core.Query
type Page[T any] = core.Page[T]
type QueryConfig = core.QueryConfig

type MemoryRepositoryConfig[T any, ID comparable] = core.MemoryRepositoryConfig[T, ID]
type MemoryRepository[T any, ID comparable] = core.MemoryRepository[T, ID]

const (
	FieldKindText        = core.FieldKindText
	FieldKindTextarea    = core.FieldKindTextarea
	FieldKindInt         = core.FieldKindInt
	FieldKindInt64       = core.FieldKindInt64
	FieldKindFloat       = core.FieldKindFloat
	FieldKindBool        = core.FieldKindBool
	FieldKindDate        = core.FieldKindDate
	FieldKindDateTime    = core.FieldKindDateTime
	FieldKindJSON        = core.FieldKindJSON
	FieldKindEnum        = core.FieldKindEnum
	FieldKindSelect      = core.FieldKindSelect
	FieldKindMultiSelect = core.FieldKindMultiSelect
	FieldKindRelation    = core.FieldKindRelation
)

var ErrNotFound = core.ErrNotFound

func New(config SiteConfig) *Site {
	return core.New(config)
}

func Register[T any, ID comparable](app *App, resource Resource[T, ID]) error {
	return core.Register(app, resource)
}

func Int64ID() IDCodec[int64] {
	return core.Int64ID()
}

func StringID() IDCodec[string] {
	return core.StringID()
}

func Text(name, label string) Field {
	return core.Text(name, label)
}

func Textarea(name, label string) Field {
	return core.Textarea(name, label)
}

func Int(name, label string) Field {
	return core.Int(name, label)
}

func Int64(name, label string) Field {
	return core.Int64(name, label)
}

func Float(name, label string) Field {
	return core.Float(name, label)
}

func Bool(name, label string) Field {
	return core.Bool(name, label)
}

func Date(name, label string) Field {
	return core.Date(name, label)
}

func DateTime(name, label string) Field {
	return core.DateTime(name, label)
}

func Time(name, label string) Field {
	return core.Time(name, label)
}

func JSON(name, label string) Field {
	return core.JSON(name, label)
}

func Enum(name, label string) Field {
	return core.Enum(name, label)
}

func Select(name, label string) Field {
	return core.Select(name, label)
}

func MultiSelect(name, label string) Field {
	return core.MultiSelect(name, label)
}

func Relation(name, label string) Field {
	return core.Relation(name, label)
}

func BindForm(fields []Field, values url.Values, dst any) ValidationErrors {
	return core.BindForm(fields, values, dst)
}

func BindJSON(fields []Field, values map[string]any, dst any, partial bool) ValidationErrors {
	return core.BindJSON(fields, values, dst, partial)
}

func RenderWidget(ctx WidgetContext) template.HTML {
	return core.RenderWidget(ctx)
}

func QueryFromRequest(r *http.Request, config QueryConfig) Query {
	return core.QueryFromRequest(r, config)
}

func NewMemoryRepository[T any, ID comparable](config MemoryRepositoryConfig[T, ID]) *MemoryRepository[T, ID] {
	return core.NewMemoryRepository(config)
}
