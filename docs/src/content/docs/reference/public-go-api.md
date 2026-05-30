---
title: Public Go API
description: Exported functions and types from the go-admin package.
---

Import the package with an alias:

```go
import admin "github.com/Newton-School/go-admin"
```

## Site

```go
func New(config admin.SiteConfig) *admin.Site
```

Creates an admin site.

```go
func (s *admin.Site) BasePath() string
func (s *admin.Site) App(name, label string) *admin.App
func (s *admin.Site) Apps() []admin.AppMeta
func (s *admin.Site) Handler() http.Handler
```

`Handler` returns the standard `net/http` handler to mount behind middleware.

## Registration

```go
func Register[T any, ID comparable](app *admin.App, resource admin.Resource[T, ID]) error
```

Registers a typed resource under an app.

## IDs

```go
type IDCodec[ID comparable] interface {
    Parse(raw string) (ID, error)
    Format(id ID) string
}
```

Built-in codecs:

```go
func Int64ID() admin.IDCodec[int64]
func StringID() admin.IDCodec[string]
```

## Repositories

```go
type Repository[T any, ID comparable] interface {
    List(context.Context, admin.Query) (admin.Page[T], error)
    Get(context.Context, ID) (T, error)
    Create(context.Context, T) (T, error)
    Update(context.Context, ID, T) (T, error)
    Delete(context.Context, ID) error
}
```

Missing objects should return:

```go
var ErrNotFound error
```

## Fields

Field constructors:

```go
func Text(name, label string) admin.Field
func Textarea(name, label string) admin.Field
func Int(name, label string) admin.Field
func Int64(name, label string) admin.Field
func Float(name, label string) admin.Field
func Bool(name, label string) admin.Field
func Date(name, label string) admin.Field
func DateTime(name, label string) admin.Field
func Time(name, label string) admin.Field
func JSON(name, label string) admin.Field
func Enum(name, label string) admin.Field
func Select(name, label string) admin.Field
func MultiSelect(name, label string) admin.Field
func Relation(name, label string) admin.Field
```

Field modifiers:

```go
func (f admin.Field) Readonly() admin.Field
func (f admin.Field) Required() admin.Field
func (f admin.Field) Help(help string) admin.Field
func (f admin.Field) Placeholder(placeholder string) admin.Field
func (f admin.Field) Options(choices []admin.Choice) admin.Field
```

## Binding

```go
func BindForm(fields []admin.Field, values url.Values, dst any) admin.ValidationErrors
func BindJSON(fields []admin.Field, values map[string]any, dst any, partial bool) admin.ValidationErrors
```

`dst` must be a pointer to a struct.

```go
type ValidationErrors map[string]string

func (v admin.ValidationErrors) Empty() bool
func (v admin.ValidationErrors) Get(field string) string
```

## Widgets

```go
func RenderWidget(ctx admin.WidgetContext) template.HTML
```

Renders the built-in safe HTML widget for a field.

```go
type WidgetContext struct {
    Field  admin.Field
    Value  any
    Errors admin.ValidationErrors
}
```

## Query Parsing

```go
func QueryFromRequest(r *http.Request, config admin.QueryConfig) admin.Query
```

Use this when building compatible custom list or API handlers.

## Memory Repository

```go
func NewMemoryRepository[T any, ID comparable](
    config admin.MemoryRepositoryConfig[T, ID],
) *admin.MemoryRepository[T, ID]
```

The memory repository implements `admin.Repository[T, ID]` and is intended for examples, tests, and local demos.

