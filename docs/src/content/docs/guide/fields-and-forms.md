---
title: Fields And Forms
description: Configure form inputs, validation, readonly values, choices, and fieldsets.
---

Fields define how a resource is rendered, edited, validated, and bound to Go structs.

## Field Constructors

| Constructor | Widget | Typical Go target |
| --- | --- | --- |
| `admin.Text` | Text input | `string` |
| `admin.Textarea` | Textarea | `string` |
| `admin.Int` | Number input | `int`, other signed or unsigned ints |
| `admin.Int64` | Number input | `int64` |
| `admin.Float` | Number input | `float32`, `float64` |
| `admin.Bool` | Checkbox | `bool` |
| `admin.Date` | Date input | `time.Time` |
| `admin.DateTime` | Datetime input | `time.Time` |
| `admin.Time` | Alias for `DateTime` | `time.Time` |
| `admin.JSON` | JSON textarea | `string`, struct, map, slice |
| `admin.Enum` | Select | `string` or integer |
| `admin.Select` | Select | `string` or integer |
| `admin.MultiSelect` | Multi-select | slice |
| `admin.Relation` | Select-style relation field | `string` or integer ID |

Example:

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

## Field Names

The field name is matched against exported struct fields using:

- Exact Go field name, such as `Name`.
- Lowercase Go field name, such as `name`.
- Snake case Go field name, such as `created_at`.
- `json` tag.
- `db` tag.
- `form` tag.

```go
type Product struct {
    ID        int64
    CreatedAt time.Time `json:"created_at"`
}

admin.DateTime("created_at", "Created At")
```

Unexported struct fields are ignored.

## Modifiers

Fields are immutable value objects. Chain modifiers when building the field.

```go
admin.Text("name", "Name").
    Required().
    Placeholder("Desk chair").
    Help("Shown in search results and row labels.")
```

Available modifiers:

| Modifier | Effect |
| --- | --- |
| `Readonly()` | Shows the value but does not bind incoming form or JSON values. |
| `Required()` | Rejects empty form values and missing full JSON create values. |
| `Help(text)` | Shows helper text near the widget. |
| `Placeholder(text)` | Adds placeholder text to input-like widgets. |
| `Options(choices)` | Adds allowed choices for enum, select, multiselect, and relation-style fields. |

## Choices

Use choices for controlled values.

```go
admin.Enum("status", "Status").Options([]admin.Choice{
    {Value: "draft", Label: "Draft"},
    {Value: "published", Label: "Published"},
})
```

When choices are present, incoming values must match one of the configured `Value` entries.

## Form Binding

HTML form posts use `admin.BindForm` internally. JSON create and update requests use `admin.BindJSON`.

You can call them directly when you want the same binding behavior outside the built-in handler:

```go
var product Product
errs := admin.BindForm(productFields, r.PostForm, &product)
if !errs.Empty() {
    // render or return errs
}
```

Readonly fields are skipped. Missing optional values are skipped. Bool fields become false when no checkbox value is submitted.

## JSON Binding

Create requests require required fields. Patch requests are partial.

```go
errs := admin.BindJSON(productFields, values, &product, true)
```

Set `partial` to `true` for patch-like updates. Set it to `false` for create-like requests.

## Date And DateTime Values

Date fields parse:

```text
2006-01-02
```

Datetime fields parse:

```text
2006-01-02T15:04
2006-01-02 15:04:05
RFC3339 values
```

Both bind to `time.Time`.

## JSON Fields

`admin.JSON` accepts JSON input. If the target field is a string, the raw JSON string is stored. For other target types, the value is unmarshaled into that type.

```go
type Product struct {
    Metadata map[string]any `json:"metadata"`
}

admin.JSON("metadata", "Metadata")
```

## Fieldsets

Fieldsets group fields on create and edit pages.

```go
Fieldsets: []admin.Fieldset{
    {
        Title:       "Basic details",
        Description: "Core product information.",
        Fields:      []string{"name", "sku", "active"},
    },
    {
        Title:     "Inventory",
        Fields:    []string{"price", "stock"},
        Collapsed: false,
    },
}
```

If no fieldsets are configured, all fields render in the order listed on the resource.

`Rows` can also be used to describe grouped field order:

```go
Fieldsets: []admin.Fieldset{
    {
        Title: "Pricing",
        Rows: [][]string{
            {"price", "stock"},
            {"active"},
        },
    },
}
```

The current server-rendered form uses the field order from `Fields` or flattened `Rows`.

