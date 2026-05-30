# go-admin

`go-admin` is a Django-inspired admin SDK for Go. It mounts as a standard `net/http` handler and builds apps, resources, list pages, create/edit forms, delete flows, bulk actions, lookup endpoints, and JSON APIs from one resource registry.

Phase 1 intentionally does not include login, sessions, users, groups, or permissions. Mount the handler behind your own middleware.

## Quick Start

```go
site := admin.New(admin.SiteConfig{Title: "Acme Admin", BasePath: "/admin"})
catalog := site.App("catalog", "Catalog")

err := admin.Register(catalog, admin.Resource[Product, int64]{
    Name:       "products",
    Label:      "Products",
    Repository: productRepo,
    ID:         admin.Int64ID(),
    Fields: []admin.Field{
        admin.Int64("id", "ID").Readonly(),
        admin.Text("name", "Name").Required(),
        admin.Bool("active", "Active"),
    },
    List: admin.ListConfig{
        Columns: []string{"id", "name", "active"},
        Search:  []string{"name"},
        Filters: []admin.Filter{{Name: "active", Label: "Active"}},
    },
})
if err != nil {
    log.Fatal(err)
}

http.Handle("/admin/", requireAdmin(site.Handler()))
```

See `examples/memory` for a runnable in-memory admin and `examples/nethttp` for a custom middleware mount.

## Project Layout

- `admin.go`, `doc.go`: public SDK facade for `github.com/ns/go-admin`.
- `internal/core`: implementation-only admin runtime.
- `internal/core/assets`: embedded templates and static files.
- `tests`: external integration tests against the public SDK API.
- `examples`: runnable usage examples.
- `docs`: user-facing guides.

## What Is Included

- App and resource registry.
- ORM-agnostic `Repository[T, ID]` interface.
- Server-rendered admin UI with embedded templates and assets.
- List, detail, create, edit, delete, filters, search, sorting, pagination, and bulk actions.
- Built-in fields/widgets for text, textarea, int, int64, float, bool, date, datetime, JSON, enum/select, multi-select, and relation-style lookup.
- JSON API under `/admin/api/v1`.
- CSRF protection for HTML forms by default.

## What Is Not Included

- Authentication, authorization, sessions, users, groups, or permissions.
- Object-level permission checks.
- Audit history.
- Complex nested inlines.
- File storage/upload toolkit.
