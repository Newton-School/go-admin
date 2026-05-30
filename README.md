# go-admin

`go-admin` is a Go SDK for building Django-style admin panels. It mounts as a standard `net/http` handler and leaves authentication, sessions, roles, and permissions to the host application.

## Install

```bash
go get github.com/Newton-School/go-admin
```

Requires Go `1.26` or newer.

## Use

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
    site := admin.New(admin.SiteConfig{
        Title:    "Acme Admin",
        BasePath: "/admin",
    })

    catalog := site.App("catalog", "Catalog")

    err := admin.Register(catalog, admin.Resource[Product, int64]{
        Name:       "products",
        Label:      "Products",
        Repository: NewProductRepository(),
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

Protect `site.Handler()` with your own middleware. The same handler serves the HTML admin and JSON API.

## Example

```bash
go run ./examples/memory
```

Open:

```text
http://localhost:8080/admin/
```

## Documentation

Full documentation is available at:

- <https://newton-school.github.io/go-admin>
- [Documentation source](docs/src/content/docs/index.md)

Use the docs for installation details, repositories, fields, list pages, actions, JSON API routes, testing, and deployment.
