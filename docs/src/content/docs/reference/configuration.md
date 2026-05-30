---
title: Configuration
description: Reference for site, resource, list, fieldset, filter, and action configuration.
---

This page lists the configuration structs used by the SDK.

## SiteConfig

```go
type SiteConfig struct {
    Title       string
    BasePath    string
    DisableCSRF bool
}
```

| Field | Description |
| --- | --- |
| `Title` | Admin title shown in the header and page titles. Defaults to `Admin`. |
| `BasePath` | Mount path for all admin routes. Defaults to `/admin`. |
| `DisableCSRF` | Disables built-in CSRF checks for HTML form posts. Keep false in normal deployments. |

## Resource

```go
type Resource[T any, ID comparable] struct {
    Name       string
    Label      string
    Repository admin.Repository[T, ID]
    ID         admin.IDCodec[ID]
    IDField    string
    Fields     []admin.Field
    List       admin.ListConfig
    Fieldsets  []admin.Fieldset
    Actions    []admin.Action[T, ID]
}
```

| Field | Required | Description |
| --- | --- | --- |
| `Name` | Yes | URL-safe resource name. |
| `Label` | No | Display label. Falls back to `Name`. |
| `Repository` | Yes | Persistence implementation. |
| `ID` | Yes | Route ID parser and formatter. |
| `IDField` | No | Struct field used for links when not `id` or `ID`. |
| `Fields` | No | Form, detail, list, and JSON binding fields. |
| `List` | No | List page behavior. |
| `Fieldsets` | No | Form grouping. |
| `Actions` | No | Bulk operations. |

## ListConfig

```go
type ListConfig struct {
    Columns []string
    Search  []string
    Sort    []admin.SortField
    Filters []admin.Filter
}
```

| Field | Description |
| --- | --- |
| `Columns` | Field names shown in the table. Empty means all fields. |
| `Search` | Searchable field names for repository query mapping. |
| `Sort` | Default sort and additional allowed sort fields. |
| `Filters` | Filter controls and allowed filter names. |

## SortField

```go
type SortField struct {
    Field string
    Desc  bool
}
```

`Desc` requests descending order.

## Filter

```go
type Filter struct {
    Name    string
    Label   string
    Choices []admin.Choice
}
```

If `Choices` is empty, the built-in list page renders boolean choices:

```go
[]admin.Choice{
    {Value: "true", Label: "Yes"},
    {Value: "false", Label: "No"},
}
```

## Choice

```go
type Choice struct {
    Value string `json:"value"`
    Label string `json:"label"`
}
```

Choices are used by enum/select fields, filters, and lookup responses.

## Fieldset

```go
type Fieldset struct {
    Title       string
    Description string
    Fields      []string
    Rows        [][]string
    Collapsed   bool
}
```

`Fields` controls field order directly. If `Fields` is empty, `Rows` is flattened and used as field order.

## Action

```go
type Action[T any, ID comparable] struct {
    Name        string
    Label       string
    Description string
    Confirm     bool
    Run         func(context.Context, admin.ActionRequest[T, ID]) (admin.ActionResult, error)
}
```

`Run` must be set for executable actions.

## QueryConfig

```go
type QueryConfig struct {
    DefaultPerPage int
    MaxPerPage     int
    AllowedSorts   []string
    FilterNames    []string
}
```

Used by `admin.QueryFromRequest`.

## MemoryRepositoryConfig

```go
type MemoryRepositoryConfig[T any, ID comparable] struct {
    GetID  func(T) ID
    SetID  func(*T, ID)
    NextID func() ID
    Search func(T, string) bool
    Filter func(T, string, []string) bool
    Less   func(T, T, string) bool
}
```

`GetID` is required for normal memory repository use. `SetID` and `NextID` let the repository assign IDs for new objects.

