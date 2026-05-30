package admin

import (
	"net/http"
	"strings"
	"testing"
)

type slugProduct struct {
	Slug string `json:"slug"`
	Name string `json:"name"`
}

func TestResourceCanUseCustomIDFieldForLinks(t *testing.T) {
	repo := NewMemoryRepository(MemoryRepositoryConfig[slugProduct, string]{
		GetID: func(p slugProduct) string { return p.Slug },
		SetID: func(p *slugProduct, id string) { p.Slug = id },
	})
	if _, err := repo.Create(t.Context(), slugProduct{Slug: "oak-chair", Name: "Oak Chair"}); err != nil {
		t.Fatalf("seed product: %v", err)
	}

	site := New(SiteConfig{})
	app := site.App("catalog", "Catalog")
	if err := Register(app, Resource[slugProduct, string]{
		Name:       "products",
		Label:      "Products",
		Repository: repo,
		ID:         StringID(),
		IDField:    "slug",
		Fields: []Field{
			Text("slug", "Slug").Readonly(),
			Text("name", "Name").Required(),
		},
		List: ListConfig{Columns: []string{"slug", "name"}},
	}); err != nil {
		t.Fatalf("register products: %v", err)
	}

	list := performRequest(site.Handler(), "GET", "/admin/catalog/products/", nil, nil)
	if list.Code != http.StatusOK {
		t.Fatalf("list status = %d; body=%s", list.Code, list.Body.String())
	}
	assertBodyContains(t, list.Body.String(), `/admin/catalog/products/oak-chair`)
}

func TestListConfigSortAppliesAsDefaultSort(t *testing.T) {
	var next int64
	repo := NewMemoryRepository(MemoryRepositoryConfig[handlerProduct, int64]{
		GetID: func(p handlerProduct) int64 { return p.ID },
		SetID: func(p *handlerProduct, id int64) { p.ID = id },
		NextID: func() int64 {
			next++
			return next
		},
		Less: func(a, b handlerProduct, field string) bool {
			return field == "name" && a.Name < b.Name
		},
	})
	_, _ = repo.Create(t.Context(), handlerProduct{Name: "B"})
	_, _ = repo.Create(t.Context(), handlerProduct{Name: "A"})

	site := New(SiteConfig{})
	app := site.App("catalog", "Catalog")
	if err := Register(app, Resource[handlerProduct, int64]{
		Name:       "products",
		Label:      "Products",
		Repository: repo,
		ID:         Int64ID(),
		Fields:     []Field{Int64("id", "ID").Readonly(), Text("name", "Name")},
		List:       ListConfig{Columns: []string{"id", "name"}, Sort: []SortField{{Field: "name"}}},
	}); err != nil {
		t.Fatalf("register products: %v", err)
	}

	list := performRequest(site.Handler(), "GET", "/admin/catalog/products/", nil, nil)
	if list.Code != http.StatusOK {
		t.Fatalf("list status = %d; body=%s", list.Code, list.Body.String())
	}
	if indexA, indexB := stringsIndex(list.Body.String(), ">A<"), stringsIndex(list.Body.String(), ">B<"); indexA == -1 || indexB == -1 || indexA > indexB {
		t.Fatalf("expected A before B with default sort; body=%s", list.Body.String())
	}
}

func stringsIndex(value, substr string) int {
	return strings.Index(value, substr)
}
