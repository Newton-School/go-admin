# go-admin

`go-admin` is a Go SDK for building Django-style admin panels. It mounts as a standard `net/http` handler, so it can be used with the standard library or any router that can serve an `http.Handler`.

Auth and permissions are intentionally not built in. Protect the mounted handler with your own middleware.

## Install

```bash
go get github.com/Newton-School/go-admin
```

Requires Go `1.26` or newer.

## Run The Example

```bash
go run ./examples/memory
```

Open:

```text
http://localhost:8080/admin/
```

The example registers multiple apps and models:

- `Catalog`: Categories, Products
- `Sales`: Customers, Orders
- `Content`: Articles
- `Operations`: Warehouses

## Minimal Setup

```go
package main

import (
    "log"
    "net/http"

    admin "github.com/Newton-School/go-admin"
)

type Product struct {
    ID     int64
    Name   string
    SKU    string
    Price  float64
    Active bool
}

func main() {
    productRepo := NewProductRepository()

    site := admin.New(admin.SiteConfig{
        Title:    "Acme Admin",
        BasePath: "/admin",
    })

    catalog := site.App("catalog", "Catalog")

    err := admin.Register(catalog, admin.Resource[Product, int64]{
        Name:       "products",
        Label:      "Products",
        Repository: productRepo,
        ID:         admin.Int64ID(),
        Fields: []admin.Field{
            admin.Int64("id", "ID").Readonly(),
            admin.Text("name", "Name").Required(),
            admin.Text("sku", "SKU").Required(),
            admin.Float("price", "Price"),
            admin.Bool("active", "Active"),
        },
        List: admin.ListConfig{
            Columns: []string{"id", "name", "sku", "price", "active"},
            Search:  []string{"name", "sku"},
            Sort:    []admin.SortField{{Field: "name"}},
            Filters: []admin.Filter{
                {Name: "active", Label: "Active"},
            },
        },
    })
    if err != nil {
        log.Fatal(err)
    }

    http.Handle("/admin/", requireAdmin(site.Handler()))
    http.Handle("/admin", requireAdmin(site.Handler()))

    log.Fatal(http.ListenAndServe(":8080", nil))
}
```

## Secure The Admin

`go-admin` does not provide login, sessions, users, roles, or permissions. Wrap the handler with your own auth middleware.

```go
func requireAdmin(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        user := currentUser(r)
        if user == nil || !user.IsStaff {
            http.Error(w, "forbidden", http.StatusForbidden)
            return
        }

        next.ServeHTTP(w, r)
    })
}
```

The same wrapper protects HTML pages and JSON APIs because both are served by `site.Handler()`.

## Repository Interface

Each admin resource needs a repository. The core SDK is ORM-agnostic.

```go
type Repository[T any, ID comparable] interface {
    List(context.Context, admin.Query) (admin.Page[T], error)
    Get(context.Context, ID) (T, error)
    Create(context.Context, T) (T, error)
    Update(context.Context, ID, T) (T, error)
    Delete(context.Context, ID) error
}
```

`admin.Query` contains normalized list-page input:

```go
type Query struct {
    Search  string
    Filters map[string][]string
    Sort    []admin.SortField
    Page    int
    PerPage int
}
```

Map this query to your database, ORM, or service layer. Return `admin.ErrNotFound` when an object does not exist.

For local demos, use the in-memory repository:

```go
repo := admin.NewMemoryRepository(admin.MemoryRepositoryConfig[Product, int64]{
    GetID: func(p Product) int64 {
        return p.ID
    },
    SetID: func(p *Product, id int64) {
        p.ID = id
    },
    NextID: func() int64 {
        nextID++
        return nextID
    },
    Search: func(p Product, term string) bool {
        return strings.Contains(strings.ToLower(p.Name), strings.ToLower(term))
    },
    Filter: func(p Product, name string, values []string) bool {
        return true
    },
    Less: func(a, b Product, field string) bool {
        return field == "name" && a.Name < b.Name
    },
})
```

## Apps And Resources

Apps group resources on the admin dashboard.

```go
catalog := site.App("catalog", "Catalog")
sales := site.App("sales", "Sales")
```

Register each model-like resource under an app:

```go
err := admin.Register(catalog, admin.Resource[Product, int64]{
    Name:       "products",
    Label:      "Products",
    Repository: productRepo,
    ID:         admin.Int64ID(),
    Fields:     productFields,
})
```

Resource names and app names are used in URLs, so keep them lowercase and URL-safe.

## ID Codecs

Use an ID codec to parse route IDs and format object links.

```go
ID: admin.Int64ID()
```

```go
ID: admin.StringID()
```

If your model uses a non-standard ID field name, set `IDField`:

```go
admin.Resource[Article, string]{
    Name:    "articles",
    ID:      admin.StringID(),
    IDField: "slug",
}
```

## Fields

Fields define form inputs, detail pages, and list values.

```go
Fields: []admin.Field{
    admin.Int64("id", "ID").Readonly(),
    admin.Text("name", "Name").Required(),
    admin.Textarea("description", "Description"),
    admin.Float("price", "Price"),
    admin.Int("stock", "Stock"),
    admin.Bool("active", "Active"),
    admin.Date("available_on", "Available On"),
    admin.DateTime("created_at", "Created At").Readonly(),
    admin.JSON("metadata", "Metadata"),
}
```

Choice fields:

```go
admin.Enum("status", "Status").Options([]admin.Choice{
    {Value: "draft", Label: "Draft"},
    {Value: "published", Label: "Published"},
})
```

Supported field constructors:

- `Text`
- `Textarea`
- `Int`
- `Int64`
- `Float`
- `Bool`
- `Date`
- `DateTime`
- `Time`
- `JSON`
- `Enum`
- `Select`
- `MultiSelect`
- `Relation`

Field names are matched against struct fields by Go name, lowercase name, snake_case name, and `json`, `db`, or `form` tags.

## List Pages

Configure columns, search, sorting, and filters with `ListConfig`.

```go
List: admin.ListConfig{
    Columns: []string{"id", "name", "sku", "price", "active"},
    Search:  []string{"name", "sku"},
    Sort:    []admin.SortField{{Field: "name"}},
    Filters: []admin.Filter{
        {Name: "active", Label: "Active"},
        {
            Name:  "status",
            Label: "Status",
            Choices: []admin.Choice{
                {Value: "draft", Label: "Draft"},
                {Value: "published", Label: "Published"},
            },
        },
    },
}
```

Your repository decides how `Search`, `Filters`, and `Sort` are applied.

## Fieldsets

Use fieldsets to group fields on create and edit pages.

```go
Fieldsets: []admin.Fieldset{
    {
        Title:  "Main",
        Fields: []string{"name", "sku", "price"},
    },
    {
        Title:       "Publishing",
        Description: "Controls visibility in the storefront.",
        Fields:      []string{"active", "created_at"},
    },
}
```

## Actions

Actions run against selected rows from the list page or from the JSON API.

```go
Actions: []admin.Action[Product, int64]{
    {
        Name:  "deactivate",
        Label: "Deactivate selected products",
        Run: func(ctx context.Context, req admin.ActionRequest[Product, int64]) (admin.ActionResult, error) {
            for _, product := range req.Objects {
                product.Active = false
                if _, err := productRepo.Update(ctx, product.ID, product); err != nil {
                    return admin.ActionResult{}, err
                }
            }

            return admin.ActionResult{Message: "products deactivated"}, nil
        },
    },
}
```

## JSON API

The JSON API is mounted under the same admin base path.

```text
GET    /admin/api/v1/apps
GET    /admin/api/v1/{app}/{resource}
POST   /admin/api/v1/{app}/{resource}
GET    /admin/api/v1/{app}/{resource}/{id}
PATCH  /admin/api/v1/{app}/{resource}/{id}
PUT    /admin/api/v1/{app}/{resource}/{id}
DELETE /admin/api/v1/{app}/{resource}/{id}
POST   /admin/api/v1/{app}/{resource}/actions/{action}
GET    /admin/api/v1/{app}/{resource}/lookup/{field}?q=term
```

Example:

```bash
curl http://localhost:8080/admin/api/v1/catalog/products
```

If you mount `site.Handler()` behind auth middleware, these API routes are protected by the same middleware.

## CSRF

HTML forms use CSRF protection by default.

Disable it only if your deployment handles CSRF elsewhere:

```go
site := admin.New(admin.SiteConfig{
    Title:       "Acme Admin",
    BasePath:    "/admin",
    DisableCSRF: true,
})
```

JSON API requests are intended to be protected by your API/auth middleware.
