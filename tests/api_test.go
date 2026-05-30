package admin_test

import (
	"bytes"
	"context"
	"encoding/json"
	. "github.com/ns/go-admin"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAdminJSONAPIListsAndMutatesResources(t *testing.T) {
	site, repo := newHandlerTestSite(t)

	apps := performJSONRequest(site.Handler(), "GET", "/admin/api/v1/apps", nil)
	if apps.Code != http.StatusOK {
		t.Fatalf("apps status = %d; body=%s", apps.Code, apps.Body.String())
	}
	assertBodyContains(t, apps.Body.String(), `"name":"catalog"`, `"resources"`)

	list := performJSONRequest(site.Handler(), "GET", "/admin/api/v1/catalog/products?q=chair", nil)
	if list.Code != http.StatusOK {
		t.Fatalf("list status = %d; body=%s", list.Code, list.Body.String())
	}
	assertBodyContains(t, list.Body.String(), `"total":1`, `"Name":"Chair"`)

	create := performJSONRequest(site.Handler(), "POST", "/admin/api/v1/catalog/products", map[string]any{
		"name":   "Lamp",
		"active": true,
	})
	if create.Code != http.StatusCreated {
		t.Fatalf("create status = %d; body=%s", create.Code, create.Body.String())
	}
	assertBodyContains(t, create.Body.String(), `"Name":"Lamp"`)

	update := performJSONRequest(site.Handler(), "PATCH", "/admin/api/v1/catalog/products/3", map[string]any{
		"name": "Tall Lamp",
	})
	if update.Code != http.StatusOK {
		t.Fatalf("update status = %d; body=%s", update.Code, update.Body.String())
	}
	updated, err := repo.Get(t.Context(), 3)
	if err != nil {
		t.Fatalf("get updated: %v", err)
	}
	if updated.Name != "Tall Lamp" || !updated.Active {
		t.Fatalf("partial json update should preserve active flag: %#v", updated)
	}

	del := performJSONRequest(site.Handler(), "DELETE", "/admin/api/v1/catalog/products/3", nil)
	if del.Code != http.StatusNoContent {
		t.Fatalf("delete status = %d; body=%s", del.Code, del.Body.String())
	}
	if _, err := repo.Get(t.Context(), 3); err == nil {
		t.Fatal("expected deleted api product to be gone")
	}
}

func TestAdminJSONAPIRunsActionsAndLookup(t *testing.T) {
	var next int64
	repo := NewMemoryRepository(MemoryRepositoryConfig[handlerProduct, int64]{
		GetID: func(p handlerProduct) int64 { return p.ID },
		SetID: func(p *handlerProduct, id int64) { p.ID = id },
		NextID: func() int64 {
			next++
			return next
		},
		Search: func(p handlerProduct, term string) bool { return stringsContainsFold(p.Name, term) },
		Less:   func(a, b handlerProduct, field string) bool { return field == "name" && a.Name < b.Name },
	})
	chair, err := repo.Create(t.Context(), handlerProduct{Name: "Chair", Active: true})
	if err != nil {
		t.Fatalf("seed chair: %v", err)
	}

	site := New(SiteConfig{Title: "Acme Admin"})
	app := site.App("catalog", "Catalog")
	if err := Register(app, Resource[handlerProduct, int64]{
		Name:       "products",
		Label:      "Products",
		Repository: repo,
		ID:         Int64ID(),
		Fields: []Field{
			Int64("id", "ID").Readonly(),
			Text("name", "Name").Required(),
			Bool("active", "Active"),
		},
		List: ListConfig{Columns: []string{"id", "name", "active"}, Search: []string{"name"}},
		Actions: []Action[handlerProduct, int64]{
			{
				Name:  "deactivate",
				Label: "Deactivate",
				Run: func(ctx context.Context, req ActionRequest[handlerProduct, int64]) (ActionResult, error) {
					for _, item := range req.Objects {
						item.Active = false
						if _, err := repo.Update(ctx, item.ID, item); err != nil {
							return ActionResult{}, err
						}
					}
					return ActionResult{Message: "deactivated"}, nil
				},
			},
		},
	}); err != nil {
		t.Fatalf("register products: %v", err)
	}

	action := performJSONRequest(site.Handler(), "POST", "/admin/api/v1/catalog/products/actions/deactivate", map[string]any{
		"ids": []any{float64(chair.ID)},
	})
	if action.Code != http.StatusOK {
		t.Fatalf("action status = %d; body=%s", action.Code, action.Body.String())
	}
	assertBodyContains(t, action.Body.String(), `"message":"deactivated"`)
	updated, err := repo.Get(t.Context(), chair.ID)
	if err != nil {
		t.Fatalf("get chair: %v", err)
	}
	if updated.Active {
		t.Fatal("expected action to deactivate product")
	}

	lookup := performJSONRequest(site.Handler(), "GET", "/admin/api/v1/catalog/products/lookup/name?q=cha", nil)
	if lookup.Code != http.StatusOK {
		t.Fatalf("lookup status = %d; body=%s", lookup.Code, lookup.Body.String())
	}
	assertBodyContains(t, lookup.Body.String(), `"value":"1"`, `"label":"Chair"`)
}

func performJSONRequest(handler http.Handler, method, target string, body any) *httptest.ResponseRecorder {
	var reader *bytes.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			panic(err)
		}
		reader = bytes.NewReader(encoded)
	} else {
		reader = bytes.NewReader(nil)
	}
	req := httptest.NewRequest(method, target, reader)
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp := httptest.NewRecorder()
	handler.ServeHTTP(resp, req)
	return resp
}
