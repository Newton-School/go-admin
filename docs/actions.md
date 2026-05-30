# Actions

Actions run on selected rows from either the HTML list page or the JSON API.

```go
admin.Action[Product, int64]{
    Name:  "deactivate",
    Label: "Deactivate",
    Run: func(ctx context.Context, req admin.ActionRequest[Product, int64]) (admin.ActionResult, error) {
        for _, product := range req.Objects {
            product.Active = false
            if _, err := repo.Update(ctx, product.ID, product); err != nil {
                return admin.ActionResult{}, err
            }
        }
        return admin.ActionResult{Message: "deactivated"}, nil
    },
}
```

HTML route:

```text
POST /admin/{app}/{resource}/actions/{action}
```

JSON route:

```text
POST /admin/api/v1/{app}/{resource}/actions/{action}
```

