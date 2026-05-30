---
title: Secure The Admin
description: Protect go-admin with application-owned authentication and authorization.
---

`go-admin` does not provide login, sessions, users, roles, or permissions. That is intentional. Mount the admin behind middleware owned by your application or by a separate auth toolkit.

## Wrap The Handler

```go
http.Handle("/admin/", requireAdmin(site.Handler()))
http.Handle("/admin", requireAdmin(site.Handler()))
```

The wrapper protects all admin routes, including HTML pages, static assets, and JSON API routes.

```go
func requireAdmin(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        user := currentUser(r)
        if user == nil || !user.IsStaff {
            http.Error(w, "forbidden", http.StatusForbidden)
            return
        }

        next.ServeHTTP(w, r)
    })
}
```

## Pass Context To Repositories

Middleware can attach request-scoped values before the admin handler runs. Repository methods receive the same `context.Context`.

```go
func requireAdmin(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        user := currentUser(r)
        if user == nil || !user.IsStaff {
            http.Error(w, "forbidden", http.StatusForbidden)
            return
        }

        ctx := context.WithValue(r.Context(), currentUserKey{}, user)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

Use that context for auditing, tenant scoping, request tracing, and application-specific authorization checks inside your repository or service layer.

## CSRF For HTML Forms

Mutating HTML forms use a built-in CSRF cookie and hidden form field named `go_admin_csrf`.

The JSON API does not use that form token. If API clients are browsers using cookies, protect the mounted handler with your own API CSRF strategy at the middleware layer.

:::caution
Only set `DisableCSRF: true` in controlled tests or when an outer middleware layer already enforces equivalent CSRF protection for HTML form posts.
:::

```go
site := admin.New(admin.SiteConfig{
    Title:       "Acme Admin",
    BasePath:    "/admin",
    DisableCSRF: true,
})
```

