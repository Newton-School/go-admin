package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDemoSiteShowsMultipleAppsAndModels(t *testing.T) {
	site := newDemoSite()

	resp := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/admin/", nil)
	site.Handler().ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d; body=%s", resp.Code, resp.Body.String())
	}
	body := resp.Body.String()
	for _, expected := range []string{
		"Catalog",
		"Products",
		"Categories",
		"Sales",
		"Customers",
		"Orders",
		"Content",
		"Articles",
		"Operations",
		"Warehouses",
	} {
		if !strings.Contains(body, expected) {
			t.Fatalf("expected admin index to contain %q; body=%s", expected, body)
		}
	}
}

func TestDemoProductListIncludesSeededData(t *testing.T) {
	site := newDemoSite()

	resp := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/admin/catalog/products/", nil)
	site.Handler().ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d; body=%s", resp.Code, resp.Body.String())
	}
	body := resp.Body.String()
	for _, expected := range []string{"Ergonomic Chair", "Standing Desk", "Active"} {
		if !strings.Contains(body, expected) {
			t.Fatalf("expected product list to contain %q; body=%s", expected, body)
		}
	}
}
