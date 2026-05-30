---
title: Core Concepts
description: The main building blocks in a go-admin application.
---

`go-admin` follows the same mental model as Django admin, but keeps persistence and auth outside the SDK.

## Site

A `Site` owns the admin registry and exposes one HTTP handler.

```go
site := admin.New(admin.SiteConfig{
    Title:    "Acme Admin",
    BasePath: "/admin",
})
```

The site serves:

- Dashboard pages for registered apps.
- List, create, edit, and delete pages for resources.
- Bulk action form posts.
- JSON API routes under `/admin/api/v1`.
- Embedded static assets for the server-rendered UI.

## App

An app is a group of related resources.

```go
catalog := site.App("catalog", "Catalog")
sales := site.App("sales", "Sales")
```

Apps appear on the dashboard in registration order. App names are used in URLs, so use lowercase URL-safe names.

## Resource

A resource is the admin representation of one model-like type.

```go
admin.Resource[Product, int64]{
    Name:       "products",
    Label:      "Products",
    Repository: products,
    ID:         admin.Int64ID(),
    Fields:     productFields,
}
```

The first generic type is the object type. The second generic type is the ID type used by the repository.

Resources define:

| Property | Purpose |
| --- | --- |
| `Name` | URL-safe resource name, for example `products`. |
| `Label` | Human-readable label shown in the UI. |
| `Repository` | Persistence implementation for list/get/create/update/delete. |
| `ID` | Codec for parsing route IDs and formatting object links. |
| `IDField` | Optional struct field name for non-standard primary key fields. |
| `Fields` | Form, detail, list, and API binding definitions. |
| `List` | List page columns, search, filters, and default sort. |
| `Fieldsets` | Optional form grouping. |
| `Actions` | Optional bulk actions. |

## Field

Fields connect an admin input to an exported struct field.

```go
admin.Text("name", "Name").Required()
admin.Bool("active", "Active")
admin.DateTime("created_at", "Created At").Readonly()
```

The field name can match the Go field name, lowercase field name, snake_case name, `json`, `db`, or `form` tag.

## Repository

Repositories are the persistence boundary. `go-admin` never imports your ORM.

```go
type Repository[T any, ID comparable] interface {
    List(context.Context, admin.Query) (admin.Page[T], error)
    Get(context.Context, ID) (T, error)
    Create(context.Context, T) (T, error)
    Update(context.Context, ID, T) (T, error)
    Delete(context.Context, ID) error
}
```

Return `admin.ErrNotFound` when an object is missing.

## Query And Page

List pages and list API requests become an `admin.Query`.

```go
type Query struct {
    Search  string
    Filters map[string][]string
    Sort    []admin.SortField
    Page    int
    PerPage int
}
```

Repositories return an `admin.Page[T]` with the current items and total count.

## ID Codec

ID codecs parse route IDs and format object links.

```go
ID: admin.Int64ID()
```

```go
ID: admin.StringID()
```

You can implement `admin.IDCodec[ID]` for custom ID formats.

## Action

Actions run against selected objects from the list page or JSON API.

```go
admin.Action[Product, int64]{
    Name:  "deactivate",
    Label: "Deactivate selected products",
    Run: func(ctx context.Context, req admin.ActionRequest[Product, int64]) (admin.ActionResult, error) {
        return admin.ActionResult{Message: "done"}, nil
    },
}
```

The SDK loads selected objects before calling your action.

