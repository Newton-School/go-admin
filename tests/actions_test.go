package admin_test

import (
	"context"
	. "github.com/ns/go-admin"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func TestListPageRunsBulkAction(t *testing.T) {
	var next int64
	repo := NewMemoryRepository(MemoryRepositoryConfig[handlerProduct, int64]{
		GetID: func(p handlerProduct) int64 { return p.ID },
		SetID: func(p *handlerProduct, id int64) { p.ID = id },
		NextID: func() int64 {
			next++
			return next
		},
		Search: func(p handlerProduct, term string) bool { return stringsContainsFold(p.Name, term) },
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
		List: ListConfig{Columns: []string{"id", "name", "active"}},
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

	list := performRequest(site.Handler(), "GET", "/admin/catalog/products/", nil, nil)
	if list.Code != http.StatusOK {
		t.Fatalf("list status = %d; body=%s", list.Code, list.Body.String())
	}
	assertBodyContains(t, list.Body.String(), "Deactivate", "name=\"ids\"")
	csrf := csrfFromResponse(t, list)

	form := url.Values{"ids": {csrfSafeID(chair.ID)}, "go_admin_csrf": {csrf.Value}}
	action := performRequest(site.Handler(), "POST", "/admin/catalog/products/actions/deactivate", strings.NewReader(form.Encode()), []*http.Cookie{csrf})
	if action.Code != http.StatusSeeOther {
		t.Fatalf("action status = %d; body=%s", action.Code, action.Body.String())
	}
	updated, err := repo.Get(t.Context(), chair.ID)
	if err != nil {
		t.Fatalf("get chair: %v", err)
	}
	if updated.Active {
		t.Fatal("expected bulk action to deactivate product")
	}
}

func csrfSafeID(id int64) string {
	return Int64ID().Format(id)
}
