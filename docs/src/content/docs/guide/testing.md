---
title: Testing
description: Test admin resources, handlers, repositories, and JSON APIs.
---

`go-admin` is easy to test because the site is a normal `http.Handler` and persistence is behind a repository interface.

## Test A Site Handler

```go
func TestAdminIndex(t *testing.T) {
    site := admin.New(admin.SiteConfig{
        Title:    "Acme Admin",
        BasePath: "/admin",
    })

    req := httptest.NewRequest(http.MethodGet, "/admin/", nil)
    resp := httptest.NewRecorder()

    site.Handler().ServeHTTP(resp, req)

    if resp.Code != http.StatusOK {
        t.Fatalf("status = %d; body=%s", resp.Code, resp.Body.String())
    }
}
```

## Use The Memory Repository In Tests

The memory repository is useful for handler tests because it implements the full repository contract without a database.

```go
func newProductRepo() *admin.MemoryRepository[Product, int64] {
    var next int64

    return admin.NewMemoryRepository(admin.MemoryRepositoryConfig[Product, int64]{
        GetID: func(p Product) int64 { return p.ID },
        SetID: func(p *Product, id int64) { p.ID = id },
        NextID: func() int64 {
            next++
            return next
        },
        Search: func(p Product, term string) bool {
            return strings.Contains(strings.ToLower(p.Name), strings.ToLower(term))
        },
        Less: func(a, b Product, field string) bool {
            return field == "name" && a.Name < b.Name
        },
    })
}
```

## Test A JSON Route

```go
func TestCreateProductAPI(t *testing.T) {
    site := newTestAdminSite(t)

    body := bytes.NewReader([]byte(`{"name":"Lamp","active":true}`))
    req := httptest.NewRequest(http.MethodPost, "/admin/api/v1/catalog/products", body)
    req.Header.Set("Content-Type", "application/json")

    resp := httptest.NewRecorder()
    site.Handler().ServeHTTP(resp, req)

    if resp.Code != http.StatusCreated {
        t.Fatalf("status = %d; body=%s", resp.Code, resp.Body.String())
    }
}
```

## Test HTML Form Posts

Mutating HTML routes require the built-in CSRF token by default. For full handler tests, fetch the form first, read the `go_admin_csrf` cookie, and submit it back as a form value.

```go
newReq := httptest.NewRequest(http.MethodGet, "/admin/catalog/products/new", nil)
newResp := httptest.NewRecorder()
site.Handler().ServeHTTP(newResp, newReq)

var csrf *http.Cookie
for _, cookie := range newResp.Result().Cookies() {
    if cookie.Name == "go_admin_csrf" {
        csrf = cookie
        break
    }
}
if csrf == nil {
    t.Fatal("csrf cookie missing")
}

form := url.Values{
    "name":          {"Lamp"},
    "active":        {"true"},
    "go_admin_csrf": {csrf.Value},
}

postReq := httptest.NewRequest(
    http.MethodPost,
    "/admin/catalog/products/new",
    strings.NewReader(form.Encode()),
)
postReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
postReq.AddCookie(csrf)

postResp := httptest.NewRecorder()
site.Handler().ServeHTTP(postResp, postReq)
```

For narrow tests that do not need CSRF behavior, create the site with `DisableCSRF: true`.

## Test Repositories Directly

Repository tests should verify query mapping:

- Search terms are applied to the intended columns.
- Filter names are whitelisted.
- Sort fields map to safe storage columns.
- Pagination returns the requested page and the full matching total.
- Missing rows return `admin.ErrNotFound`.

Keep repository tests independent from the HTTP handler when the behavior is storage-specific.

