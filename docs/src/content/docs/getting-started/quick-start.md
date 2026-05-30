---
title: Quick Start
description: Register one resource and serve a working admin panel.
---

This page shows the smallest practical setup: a site, one app, one resource, one repository, and the mounted handler.

## Define A Model

Use exported struct fields. Field names are matched by Go field name, lowercase field name, snake_case field name, `json`, `db`, or `form` tag.

```go
type Product struct {
    ID     int64
    Name   string
    SKU    string `json:"sku"`
    Price  float64
    Active bool
}
```

## Create A Repository

For demos and tests, use the built-in memory repository:

```go
var nextProductID int64

products := admin.NewMemoryRepository(admin.MemoryRepositoryConfig[Product, int64]{
    GetID: func(p Product) int64 {
        return p.ID
    },
    SetID: func(p *Product, id int64) {
        p.ID = id
    },
    NextID: func() int64 {
        nextProductID++
        return nextProductID
    },
    Search: func(p Product, term string) bool {
        haystack := strings.ToLower(p.Name + " " + p.SKU)
        return strings.Contains(haystack, strings.ToLower(term))
    },
    Filter: func(p Product, name string, values []string) bool {
        if name != "active" {
            return true
        }
        for _, value := range values {
            if value == "true" && p.Active {
                return true
            }
            if value == "false" && !p.Active {
                return true
            }
        }
        return false
    },
    Less: func(a, b Product, field string) bool {
        switch field {
        case "name":
            return a.Name < b.Name
        case "price":
            return a.Price < b.Price
        default:
            return a.ID < b.ID
        }
    },
})
```

Production applications usually implement `admin.Repository` on top of a database, ORM, or service client.

## Register The Resource

Apps group related resources on the dashboard. Resources describe the model, fields, list page, and persistence layer.

```go
site := admin.New(admin.SiteConfig{
    Title:    "Acme Admin",
    BasePath: "/admin",
})

catalog := site.App("catalog", "Catalog")

err := admin.Register(catalog, admin.Resource[Product, int64]{
    Name:       "products",
    Label:      "Products",
    Repository: products,
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
```

## Serve It

```go
http.Handle("/admin/", requireAdmin(site.Handler()))
http.Handle("/admin", requireAdmin(site.Handler()))

log.Fatal(http.ListenAndServe(":8080", nil))
```

Open `/admin/` after your auth middleware allows the request.

