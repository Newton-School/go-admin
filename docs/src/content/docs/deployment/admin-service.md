---
title: Admin Service
description: Deploy a Go application that mounts go-admin.
---

`go-admin` is deployed with your Go service. It does not run a separate admin server unless you choose to create one.

## Mount Behind Middleware

Always protect the mounted handler:

```go
adminHandler := requireAdmin(site.Handler())

http.Handle("/admin/", adminHandler)
http.Handle("/admin", adminHandler)
```

The same handler serves HTML pages, static assets, and JSON API routes.

## Reverse Proxies

Keep `BasePath` aligned with the path that reaches your Go process.

```go
site := admin.New(admin.SiteConfig{
    Title:    "Acme Admin",
    BasePath: "/admin",
})
```

If a proxy strips `/admin` before forwarding, either stop stripping the prefix or mount the admin at the forwarded path. The handler expects incoming request paths to include `BasePath`.

## TLS And Cookies

Terminate TLS before admin traffic reaches users. The built-in CSRF cookie is `HttpOnly` and `SameSite=Lax`. Authentication cookies and session settings belong to your application middleware.

## Persistence

Use production repositories backed by your database or service layer. The built-in memory repository is for examples, tests, and local demos.

Production repositories should handle:

- Tenant scoping from request context.
- Audit metadata from request context.
- Database transactions for multi-row mutations.
- Safe mapping from list sort/filter names to storage columns.
- `admin.ErrNotFound` for missing objects.

## JSON API Clients

The JSON API is served from the same handler:

```text
/admin/api/v1
```

If API clients use bearer tokens, mTLS, or internal service auth, enforce that in middleware before the admin handler. The SDK does not authenticate API clients.

## Static Assets

Admin CSS, JavaScript, and templates are embedded in the Go module. You do not need a separate static file server for the admin UI.

## Health Checks

Use your application health endpoint for process health. Avoid using admin routes as unauthenticated health checks because they should remain protected by admin middleware.

