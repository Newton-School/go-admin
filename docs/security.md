# Security Notes

`go-admin` deliberately excludes auth and permissions in phase 1.

Required host responsibilities:

- Mount the admin behind authentication middleware.
- Restrict network access when appropriate.
- Add authorization checks in middleware, repositories, or action handlers.
- Avoid exposing admin APIs publicly without a separate auth layer.

Built-in protections:

- HTML forms use a double-submit CSRF token by default.
- Template rendering escapes values by default.
- JSON APIs are not CSRF-protected by default because they are intended to be protected by host API auth middleware.

Disable form CSRF only for controlled deployments:

```go
site := admin.New(admin.SiteConfig{DisableCSRF: true})
```

