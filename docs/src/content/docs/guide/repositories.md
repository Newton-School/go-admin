---
title: Repositories
description: Implement the persistence contract for admin resources.
---

Repositories are the only persistence API that `go-admin` needs. They let the SDK stay independent from SQL builders, ORMs, document stores, and service APIs.

## Contract

```go
type Repository[T any, ID comparable] interface {
    List(context.Context, admin.Query) (admin.Page[T], error)
    Get(context.Context, ID) (T, error)
    Create(context.Context, T) (T, error)
    Update(context.Context, ID, T) (T, error)
    Delete(context.Context, ID) error
}
```

Return `admin.ErrNotFound` from `Get`, `Update`, or `Delete` when the object does not exist. The HTML UI and JSON API translate that error to `404`.

## Query Input

`List` receives a normalized query:

```go
type Query struct {
    Search  string
    Filters map[string][]string
    Sort    []admin.SortField
    Page    int
    PerPage int
}
```

`QueryFromRequest` only accepts filter names and sort fields that were configured on the resource list page. Unknown filters and unknown sorts are ignored before the repository sees the query.

## Page Output

```go
type Page[T any] struct {
    Items   []T `json:"items"`
    Total   int `json:"total"`
    Page    int `json:"page"`
    PerPage int `json:"per_page"`
}
```

`Total` is the number of matching objects before pagination. `Items` should contain only the requested page.

## Pagination

Use `Page` and `PerPage` to calculate the result window.

```go
offset := (query.Page - 1) * query.PerPage
limit := query.PerPage
```

The built-in parser defaults to page `1`, defaults to `25` rows per page, and caps `per_page` at `100` for admin list routes.

## Search

List search is configured on the resource:

```go
List: admin.ListConfig{
    Search: []string{"name", "sku"},
}
```

The repository receives only the search term in `query.Search`. Use the configured searchable columns as your own whitelist when building SQL or service queries.

```go
if query.Search != "" {
    term := "%" + strings.ToLower(query.Search) + "%"
    stmt = stmt.Where("lower(name) like ? or lower(sku) like ?", term, term)
}
```

## Filters

Configured filters appear in `query.Filters`.

```go
List: admin.ListConfig{
    Filters: []admin.Filter{
        {Name: "active", Label: "Active"},
        {Name: "status", Label: "Status"},
    },
}
```

Repository mapping:

```go
if values := query.Filters["active"]; len(values) > 0 {
    stmt = stmt.Where("active in ?", values)
}
```

Filters can have multiple selected values because query strings may include repeated keys.

## Sorting

Configured sort fields become `query.Sort`.

```go
List: admin.ListConfig{
    Sort: []admin.SortField{{Field: "created_at", Desc: true}},
}
```

Sort requests use the `sort` query parameter. Prefix the field with `-` for descending sort:

```text
/admin/catalog/products/?sort=-created_at
```

Map field names to known columns. Do not concatenate untrusted field names into raw SQL.

```go
sortColumns := map[string]string{
    "name":       "name",
    "created_at": "created_at",
}

for _, sort := range query.Sort {
    column, ok := sortColumns[sort.Field]
    if !ok {
        continue
    }
    direction := "asc"
    if sort.Desc {
        direction = "desc"
    }
    stmt = stmt.Order(column + " " + direction)
}
```

## Memory Repository

Use `admin.NewMemoryRepository` for examples, tests, and local demos.

```go
repo := admin.NewMemoryRepository(admin.MemoryRepositoryConfig[Product, int64]{
    GetID:  func(p Product) int64 { return p.ID },
    SetID:  func(p *Product, id int64) { p.ID = id },
    NextID: func() int64 { nextID++; return nextID },
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

The memory repository is concurrency-safe, but it is not a database and should not be used as production storage.

