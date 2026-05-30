---
title: List Pages
description: Configure columns, search, filters, sorting, and pagination.
---

The list page is the admin changelist for a resource. It is controlled by `admin.ListConfig`.

```go
List: admin.ListConfig{
    Columns: []string{"id", "name", "sku", "price", "active"},
    Search:  []string{"name", "sku"},
    Sort:    []admin.SortField{{Field: "name"}},
    Filters: []admin.Filter{
        {Name: "active", Label: "Active"},
    },
}
```

## Columns

`Columns` controls which fields appear in the table.

```go
Columns: []string{"id", "name", "sku", "price", "active"}
```

If `Columns` is empty, all resource fields are shown.

Rows link to the detail page using the resource ID codec.

## Search

`Search` declares which fields are searchable.

```go
Search: []string{"name", "sku"}
```

The SDK parses `q` from the request and passes the term to the repository as `query.Search`.

```text
/admin/catalog/products/?q=chair
```

The repository decides how to apply the search term.

## Filters

Filters appear in the sidebar and are passed to the repository as `query.Filters`.

```go
Filters: []admin.Filter{
    {Name: "active", Label: "Active"},
}
```

If no choices are provided, a boolean filter renders as `Yes` and `No`.

```go
Filters: []admin.Filter{
    {
        Name:  "status",
        Label: "Status",
        Choices: []admin.Choice{
            {Value: "draft", Label: "Draft"},
            {Value: "published", Label: "Published"},
        },
    },
}
```

Filter values are accepted from repeated query parameters:

```text
/admin/content/articles/?status=draft&status=published
```

## Sorting

`Sort` defines the default sort when the request does not include a sort parameter.

```go
Sort: []admin.SortField{{Field: "placed_at", Desc: true}}
```

Users and API clients can request sorting with `sort`:

```text
/admin/sales/orders/?sort=total
/admin/sales/orders/?sort=-placed_at
```

Only list columns and configured sort fields are accepted. Unknown sort values are ignored.

## Pagination

List routes parse:

| Query parameter | Meaning |
| --- | --- |
| `page` | 1-based page number. |
| `per_page` | Rows per page. |

Defaults:

| Setting | Value |
| --- | --- |
| Default page | `1` |
| Default rows per page | `25` |
| Maximum rows per page | `100` |

The repository receives the normalized values in `query.Page` and `query.PerPage`.

## Query Example

This request:

```text
/admin/catalog/products/?q=desk&active=true&page=2&per_page=10&sort=-price
```

becomes a query like:

```go
admin.Query{
    Search: "desk",
    Filters: map[string][]string{
        "active": {"true"},
    },
    Sort: []admin.SortField{
        {Field: "price", Desc: true},
    },
    Page:    2,
    PerPage: 10,
}
```

