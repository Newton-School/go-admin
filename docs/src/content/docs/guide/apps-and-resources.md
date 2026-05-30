---
title: Apps And Resources
description: Group models into apps and register them as admin resources.
---

Apps and resources define the admin navigation and route shape.

## Route Shape

With `BasePath: "/admin"`, a resource registered as `catalog/products` gets these HTML routes:

| Route | Purpose |
| --- | --- |
| `/admin/` | Admin dashboard. |
| `/admin/catalog/` | App dashboard. |
| `/admin/catalog/products/` | Product list page. |
| `/admin/catalog/products/new` | Product create page. |
| `/admin/catalog/products/{id}` | Product edit/detail page. |
| `/admin/catalog/products/{id}/delete` | Product delete confirmation page. |
| `/admin/catalog/products/actions/{name}` | Bulk action form endpoint. |

The matching JSON API routes are documented in the API guide.

## Create Apps

Call `site.App` once per logical area:

```go
catalog := site.App("catalog", "Catalog")
sales := site.App("sales", "Sales")
content := site.App("content", "Content")
```

If you call `site.App` again with the same name, the existing app is returned. A non-empty label updates the app label.

## Register Resources

```go
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
})
if err != nil {
    log.Fatal(err)
}
```

`Register` validates the app name, resource name, repository, ID codec, and duplicate resource names.

## Name Rules

App and resource names are URL segments. Use lowercase names that match this pattern:

```text
^[a-z][a-z0-9_/-]*$
```

Recommended examples:

- `catalog`
- `products`
- `support/tickets`
- `inventory_items`

Avoid names with spaces, uppercase letters, or punctuation.

## Labels

Labels are display text. If a label is empty, the name is used.

```go
site.App("catalog", "Catalog")

admin.Resource[Product, int64]{
    Name:  "products",
    Label: "Products",
}
```

## ID Field

By default, object links are built from `id` or `ID`.

Set `IDField` when the primary key is named differently:

```go
type Article struct {
    Slug  string
    Title string
}

admin.Resource[Article, string]{
    Name:    "articles",
    Label:   "Articles",
    ID:      admin.StringID(),
    IDField: "slug",
}
```

`IDField` is only used to read the ID from returned objects. The repository still owns actual persistence.

## Registration Order

Apps and resources render in the order you register them.

```go
catalog := site.App("catalog", "Catalog")
must(admin.Register(catalog, categoriesResource))
must(admin.Register(catalog, productsResource))
```

Use registration order to keep the dashboard predictable.

