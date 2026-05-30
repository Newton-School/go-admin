package main

import (
	"context"
	"log"
	"net/http"
	"strings"

	admin "github.com/Newton-School/go-admin"
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
			return strings.Contains(strings.ToLower(p.Name), strings.ToLower(term))
		},
	})
	_, _ = repo.Create(context.Background(), Product{Name: "Chair", Active: true})

	site := admin.New(admin.SiteConfig{Title: "Middleware Admin"})
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
		List: admin.ListConfig{Columns: []string{"id", "name", "active"}, Search: []string{"name"}},
	}); err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.Handle("/admin/", requireAdmin(site.Handler()))
	mux.Handle("/admin", requireAdmin(site.Handler()))

	log.Println("send X-Admin: true and open http://localhost:8080/admin/")
	log.Fatal(http.ListenAndServe(":8080", mux))
}

func requireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Admin") != "true" {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}
