package admin_test

import (
	. "github.com/Newton-School/go-admin"
	"net/http/httptest"
	"testing"
)

func TestQueryFromRequestAppliesSafeDefaults(t *testing.T) {
	req := httptest.NewRequest("GET", "/admin/catalog/products/?q=chair&page=3&per_page=200&sort=-created_at&active=true&active=false", nil)
	query := QueryFromRequest(req, QueryConfig{
		DefaultPerPage: 25,
		MaxPerPage:     100,
		AllowedSorts:   []string{"created_at", "name"},
		FilterNames:    []string{"active"},
	})

	if query.Search != "chair" {
		t.Fatalf("expected search chair, got %q", query.Search)
	}
	if query.Page != 3 {
		t.Fatalf("expected page 3, got %d", query.Page)
	}
	if query.PerPage != 100 {
		t.Fatalf("expected capped per_page 100, got %d", query.PerPage)
	}
	if len(query.Sort) != 1 || query.Sort[0].Field != "created_at" || !query.Sort[0].Desc {
		t.Fatalf("unexpected sort: %#v", query.Sort)
	}
	if got := query.Filters["active"]; len(got) != 2 || got[0] != "true" || got[1] != "false" {
		t.Fatalf("unexpected active filter: %#v", got)
	}
}

func TestQueryFromRequestIgnoresUnknownSortsAndFilters(t *testing.T) {
	req := httptest.NewRequest("GET", "/admin/catalog/products/?page=-1&per_page=0&sort=-unknown&secret=1", nil)
	query := QueryFromRequest(req, QueryConfig{
		DefaultPerPage: 20,
		MaxPerPage:     50,
		AllowedSorts:   []string{"name"},
		FilterNames:    []string{"active"},
	})

	if query.Page != 1 {
		t.Fatalf("expected page fallback to 1, got %d", query.Page)
	}
	if query.PerPage != 20 {
		t.Fatalf("expected default per_page 20, got %d", query.PerPage)
	}
	if len(query.Sort) != 0 {
		t.Fatalf("expected unknown sort ignored, got %#v", query.Sort)
	}
	if len(query.Filters) != 0 {
		t.Fatalf("expected unknown filters ignored, got %#v", query.Filters)
	}
}
