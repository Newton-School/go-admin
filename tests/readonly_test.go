package admin_test

import (
	. "github.com/Newton-School/go-admin"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"testing"
)

type readonlyProduct struct {
	ID   int64
	Slug string
	Name string
}

func newReadonlyFieldTestSite(t *testing.T) (*Site, *MemoryRepository[readonlyProduct, int64]) {
	t.Helper()

	var next int64
	repo := NewMemoryRepository(MemoryRepositoryConfig[readonlyProduct, int64]{
		GetID: func(p readonlyProduct) int64 { return p.ID },
		SetID: func(p *readonlyProduct, id int64) { p.ID = id },
		NextID: func() int64 {
			next++
			return next
		},
	})

	site := New(SiteConfig{Title: "Acme Admin", BasePath: "/admin"})
	app := site.App("catalog", "Catalog")
	if err := Register(app, Resource[readonlyProduct, int64]{
		Name:       "products",
		Label:      "Products",
		Repository: repo,
		ID:         Int64ID(),
		Fields: []Field{
			Int64("id", "ID").Readonly(),
			Text("slug", "Slug").Readonly(),
			Text("name", "Name").Required(),
		},
		List: ListConfig{Columns: []string{"id", "slug", "name"}},
	}); err != nil {
		t.Fatalf("register readonly products: %v", err)
	}

	return site, repo
}

func TestReadonlyFieldsCannotBeMutatedThroughJSONAPI(t *testing.T) {
	site, repo := newReadonlyFieldTestSite(t)
	seed, err := repo.Create(t.Context(), readonlyProduct{Slug: "server-slug", Name: "Chair"})
	if err != nil {
		t.Fatalf("seed product: %v", err)
	}

	update := performJSONRequest(site.Handler(), "PATCH", "/admin/api/v1/catalog/products/"+strconv.FormatInt(seed.ID, 10), map[string]any{
		"slug": "client-slug",
		"name": "Updated Chair",
	})
	if update.Code != http.StatusOK {
		t.Fatalf("update status = %d; body=%s", update.Code, update.Body.String())
	}
	updated, err := repo.Get(t.Context(), seed.ID)
	if err != nil {
		t.Fatalf("get updated product: %v", err)
	}
	if updated.Slug != "server-slug" || updated.Name != "Updated Chair" {
		t.Fatalf("readonly JSON update changed protected data: %#v", updated)
	}
	if strings.Contains(update.Body.String(), "client-slug") {
		t.Fatalf("readonly JSON response exposed hijacked value: %s", update.Body.String())
	}

	create := performJSONRequest(site.Handler(), "POST", "/admin/api/v1/catalog/products", map[string]any{
		"slug": "client-created",
		"name": "Desk",
	})
	if create.Code != http.StatusCreated {
		t.Fatalf("create status = %d; body=%s", create.Code, create.Body.String())
	}
	created, err := repo.Get(t.Context(), seed.ID+1)
	if err != nil {
		t.Fatalf("get created product: %v", err)
	}
	if created.Slug != "" || created.Name != "Desk" {
		t.Fatalf("readonly JSON create accepted protected data: %#v", created)
	}
	if strings.Contains(create.Body.String(), "client-created") {
		t.Fatalf("readonly JSON response exposed hijacked create value: %s", create.Body.String())
	}
}

func TestReadonlyFieldsCannotBeMutatedThroughHTMLForms(t *testing.T) {
	site, repo := newReadonlyFieldTestSite(t)
	seed, err := repo.Create(t.Context(), readonlyProduct{Slug: "server-slug", Name: "Chair"})
	if err != nil {
		t.Fatalf("seed product: %v", err)
	}

	detail := performRequest(site.Handler(), "GET", "/admin/catalog/products/"+strconv.FormatInt(seed.ID, 10), nil, nil)
	if detail.Code != http.StatusOK {
		t.Fatalf("detail status = %d; body=%s", detail.Code, detail.Body.String())
	}
	csrf := csrfFromResponse(t, detail)
	updateForm := url.Values{
		"slug":          {"client-slug"},
		"name":          {"Updated Chair"},
		"go_admin_csrf": {csrf.Value},
	}
	update := performRequest(site.Handler(), "POST", "/admin/catalog/products/"+strconv.FormatInt(seed.ID, 10), strings.NewReader(updateForm.Encode()), []*http.Cookie{csrf})
	if update.Code != http.StatusSeeOther {
		t.Fatalf("update status = %d; body=%s", update.Code, update.Body.String())
	}
	updated, err := repo.Get(t.Context(), seed.ID)
	if err != nil {
		t.Fatalf("get updated product: %v", err)
	}
	if updated.Slug != "server-slug" || updated.Name != "Updated Chair" {
		t.Fatalf("readonly form update changed protected data: %#v", updated)
	}

	newPage := performRequest(site.Handler(), "GET", "/admin/catalog/products/new", nil, nil)
	if newPage.Code != http.StatusOK {
		t.Fatalf("new status = %d; body=%s", newPage.Code, newPage.Body.String())
	}
	createCSRF := csrfFromResponse(t, newPage)
	createForm := url.Values{
		"slug":          {"client-created"},
		"name":          {"Desk"},
		"go_admin_csrf": {createCSRF.Value},
	}
	create := performRequest(site.Handler(), "POST", "/admin/catalog/products/new", strings.NewReader(createForm.Encode()), []*http.Cookie{createCSRF})
	if create.Code != http.StatusSeeOther {
		t.Fatalf("create status = %d; body=%s", create.Code, create.Body.String())
	}
	created, err := repo.Get(t.Context(), seed.ID+1)
	if err != nil {
		t.Fatalf("get created product: %v", err)
	}
	if created.Slug != "" || created.Name != "Desk" {
		t.Fatalf("readonly form create accepted protected data: %#v", created)
	}
}
