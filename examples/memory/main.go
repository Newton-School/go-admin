package main

import (
	"context"
	"log"
	"net/http"
	"strings"

	admin "github.com/ns/go-admin"
)

type Product struct {
	ID     int64
	Name   string
	Active bool
}

func main() {
	var next int64
	repo := admin.NewMemoryRepository(admin.MemoryRepositoryConfig[Product, int64]{
		GetID: func(p Product) int64 { return p.ID },
		SetID: func(p *Product, id int64) { p.ID = id },
		NextID: func() int64 {
			next++
			return next
		},
		Search: func(p Product, term string) bool {
			return stringsContainsFold(p.Name, term)
		},
		Filter: func(p Product, name string, values []string) bool {
			return name == "active" && boolInValues(p.Active, values)
		},
		Less: func(a, b Product, field string) bool {
			return field == "name" && a.Name < b.Name
		},
	})
	_, _ = repo.Create(context.Background(), Product{Name: "Chair", Active: true})
	_, _ = repo.Create(context.Background(), Product{Name: "Desk", Active: false})

	site := admin.New(admin.SiteConfig{Title: "Acme Admin", BasePath: "/admin"})
	app := site.App("catalog", "Catalog")
	if err := admin.Register(app, admin.Resource[Product, int64]{
		Name:       "products",
		Label:      "Products",
		Repository: repo,
		ID:         admin.Int64ID(),
		Fields: []admin.Field{
			admin.Int64("id", "ID").Readonly(),
			admin.Text("name", "Name").Required(),
			admin.Bool("active", "Active"),
		},
		List: admin.ListConfig{
			Columns: []string{"id", "name", "active"},
			Search:  []string{"name"},
			Sort:    []admin.SortField{{Field: "name"}},
			Filters: []admin.Filter{{Name: "active", Label: "Active"}},
		},
		Actions: []admin.Action[Product, int64]{
			{
				Name:  "deactivate",
				Label: "Deactivate",
				Run: func(ctx context.Context, req admin.ActionRequest[Product, int64]) (admin.ActionResult, error) {
					for _, product := range req.Objects {
						product.Active = false
						if _, err := repo.Update(ctx, product.ID, product); err != nil {
							return admin.ActionResult{}, err
						}
					}
					return admin.ActionResult{Message: "deactivated"}, nil
				},
			},
		},
	}); err != nil {
		log.Fatal(err)
	}

	http.Handle("/admin/", site.Handler())
	http.Handle("/admin", site.Handler())
	log.Println("admin listening at http://localhost:8080/admin/")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func stringsContainsFold(value, term string) bool {
	return strings.Contains(strings.ToLower(value), strings.ToLower(term))
}

func boolInValues(value bool, values []string) bool {
	for _, candidate := range values {
		normalized := strings.ToLower(strings.TrimSpace(candidate))
		if value && (normalized == "true" || normalized == "1" || normalized == "on") {
			return true
		}
		if !value && (normalized == "false" || normalized == "0" || normalized == "off") {
			return true
		}
	}
	return false
}
