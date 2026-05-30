# Quickstart

Create a site, group resources into apps, register typed resources, then mount the handler.

```go
site := admin.New(admin.SiteConfig{Title: "Acme Admin", BasePath: "/admin"})
app := site.App("catalog", "Catalog")

admin.Register(app, admin.Resource[Product, int64]{
    Name:       "products",
    Label:      "Products",
    Repository: productRepo,
    ID:         admin.Int64ID(),
    Fields: []admin.Field{
        admin.Int64("id", "ID").Readonly(),
        admin.Text("name", "Name").Required(),
        admin.Bool("active", "Active"),
    },
})

http.Handle("/admin/", site.Handler())
```

Use host middleware to protect the handler:

```go
http.Handle("/admin/", requireInternalUser(site.Handler()))
```

The same resource registry powers HTML pages and JSON APIs.

