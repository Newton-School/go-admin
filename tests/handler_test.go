package admin_test

import (
	. "github.com/Newton-School/go-admin"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strings"
	"testing"
)

type handlerProduct struct {
	ID     int64
	Name   string
	Active bool
}

func newHandlerTestSite(t *testing.T) (*Site, *MemoryRepository[handlerProduct, int64]) {
	t.Helper()

	var next int64
	repo := NewMemoryRepository(MemoryRepositoryConfig[handlerProduct, int64]{
		GetID: func(p handlerProduct) int64 { return p.ID },
		SetID: func(p *handlerProduct, id int64) { p.ID = id },
		NextID: func() int64 {
			next++
			return next
		},
		Search: func(p handlerProduct, term string) bool {
			return stringsContainsFold(p.Name, term)
		},
		Filter: func(p handlerProduct, name string, values []string) bool {
			return name == "active" && boolInValues(p.Active, values)
		},
		Less: func(a, b handlerProduct, field string) bool {
			return field == "name" && a.Name < b.Name
		},
	})
	if _, err := repo.Create(t.Context(), handlerProduct{Name: "Chair", Active: true}); err != nil {
		t.Fatalf("seed chair: %v", err)
	}
	if _, err := repo.Create(t.Context(), handlerProduct{Name: "Desk", Active: false}); err != nil {
		t.Fatalf("seed desk: %v", err)
	}

	site := New(SiteConfig{Title: "Acme Admin", BasePath: "/admin"})
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
		List: ListConfig{
			Columns: []string{"id", "name", "active"},
			Search:  []string{"name"},
			Sort:    []SortField{{Field: "name"}},
			Filters: []Filter{{Name: "active", Label: "Active"}},
		},
	}); err != nil {
		t.Fatalf("register products: %v", err)
	}

	return site, repo
}

func TestAdminIndexAndListPages(t *testing.T) {
	site, _ := newHandlerTestSite(t)

	index := performRequest(site.Handler(), "GET", "/admin/", nil, nil)
	if index.Code != http.StatusOK {
		t.Fatalf("index status = %d; body=%s", index.Code, index.Body.String())
	}
	assertBodyContains(t, index.Body.String(), "Acme Admin", "Catalog", "Products")

	list := performRequest(site.Handler(), "GET", "/admin/catalog/products/?q=chair&active=true", nil, nil)
	if list.Code != http.StatusOK {
		t.Fatalf("list status = %d; body=%s", list.Code, list.Body.String())
	}
	assertBodyContains(t, list.Body.String(), "Chair", "Add Products")
	if strings.Contains(list.Body.String(), "Desk") {
		t.Fatalf("expected search/filter to hide Desk; body=%s", list.Body.String())
	}
}

func TestCreateEditAndDeleteFlow(t *testing.T) {
	site, repo := newHandlerTestSite(t)

	newResp := performRequest(site.Handler(), "GET", "/admin/catalog/products/new", nil, nil)
	if newResp.Code != http.StatusOK {
		t.Fatalf("new status = %d; body=%s", newResp.Code, newResp.Body.String())
	}
	assertBodyContains(t, newResp.Body.String(), "New Products", "go_admin_csrf")
	csrf := csrfFromResponse(t, newResp)

	createForm := url.Values{"name": {"Lamp"}, "active": {"true"}, "go_admin_csrf": {csrf.Value}}
	createResp := performRequest(site.Handler(), "POST", "/admin/catalog/products/new", strings.NewReader(createForm.Encode()), []*http.Cookie{csrf})
	if createResp.Code != http.StatusSeeOther {
		t.Fatalf("create status = %d; body=%s", createResp.Code, createResp.Body.String())
	}
	if location := createResp.Header().Get("Location"); location != "/admin/catalog/products/3" {
		t.Fatalf("expected redirect to detail, got %q", location)
	}

	detail := performRequest(site.Handler(), "GET", "/admin/catalog/products/3", nil, nil)
	if detail.Code != http.StatusOK {
		t.Fatalf("detail status = %d; body=%s", detail.Code, detail.Body.String())
	}
	assertBodyContains(t, detail.Body.String(), "Edit Products", "Lamp")
	editCSRF := csrfFromResponse(t, detail)

	updateForm := url.Values{"name": {"Floor Lamp"}, "go_admin_csrf": {editCSRF.Value}}
	updateResp := performRequest(site.Handler(), "POST", "/admin/catalog/products/3", strings.NewReader(updateForm.Encode()), []*http.Cookie{editCSRF})
	if updateResp.Code != http.StatusSeeOther {
		t.Fatalf("update status = %d; body=%s", updateResp.Code, updateResp.Body.String())
	}
	updated, err := repo.Get(t.Context(), 3)
	if err != nil {
		t.Fatalf("get updated: %v", err)
	}
	if updated.Name != "Floor Lamp" || updated.Active {
		t.Fatalf("unexpected updated product: %#v", updated)
	}

	deletePage := performRequest(site.Handler(), "GET", "/admin/catalog/products/3/delete", nil, nil)
	if deletePage.Code != http.StatusOK {
		t.Fatalf("delete page status = %d; body=%s", deletePage.Code, deletePage.Body.String())
	}
	deleteCSRF := csrfFromResponse(t, deletePage)
	deleteForm := url.Values{"go_admin_csrf": {deleteCSRF.Value}}
	deleteResp := performRequest(site.Handler(), "POST", "/admin/catalog/products/3/delete", strings.NewReader(deleteForm.Encode()), []*http.Cookie{deleteCSRF})
	if deleteResp.Code != http.StatusSeeOther {
		t.Fatalf("delete status = %d; body=%s", deleteResp.Code, deleteResp.Body.String())
	}
	if _, err := repo.Get(t.Context(), 3); err == nil {
		t.Fatal("expected deleted product to be gone")
	}
}

func TestMutatingFormsRequireCSRF(t *testing.T) {
	site, _ := newHandlerTestSite(t)
	form := url.Values{"name": {"Lamp"}}
	resp := performRequest(site.Handler(), "POST", "/admin/catalog/products/new", strings.NewReader(form.Encode()), nil)
	if resp.Code != http.StatusForbidden {
		t.Fatalf("expected forbidden without csrf, got %d; body=%s", resp.Code, resp.Body.String())
	}
}

func performRequest(handler http.Handler, method, target string, body io.Reader, cookies []*http.Cookie) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, target, body)
	if body != nil {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}
	resp := httptest.NewRecorder()
	handler.ServeHTTP(resp, req)
	return resp
}

func csrfFromResponse(t *testing.T, resp *httptest.ResponseRecorder) *http.Cookie {
	t.Helper()

	body := resp.Body.String()
	if !regexp.MustCompile(`name="go_admin_csrf"`).MatchString(body) {
		t.Fatalf("csrf hidden input missing from body: %s", body)
	}
	for _, cookie := range resp.Result().Cookies() {
		if cookie.Name == "go_admin_csrf" {
			return cookie
		}
	}
	t.Fatalf("csrf cookie missing from response: %#v", resp.Result().Cookies())
	return nil
}

func assertBodyContains(t *testing.T, body string, values ...string) {
	t.Helper()
	for _, value := range values {
		if !strings.Contains(body, value) {
			t.Fatalf("expected body to contain %q; body=%s", value, body)
		}
	}
}
