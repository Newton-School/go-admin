---
title: Actions
description: Add bulk operations to resource list pages and API clients.
---

Actions let operators run custom logic against selected rows.

## Define An Action

```go
Actions: []admin.Action[Product, int64]{
    {
        Name:        "deactivate",
        Label:       "Deactivate selected products",
        Description: "Marks selected products as inactive.",
        Run: func(ctx context.Context, req admin.ActionRequest[Product, int64]) (admin.ActionResult, error) {
            for _, product := range req.Objects {
                product.Active = false
                if _, err := products.Update(ctx, product.ID, product); err != nil {
                    return admin.ActionResult{}, err
                }
            }
            return admin.ActionResult{Message: "products deactivated"}, nil
        },
    },
}
```

`Name` is used in the URL. `Label` is shown in the list page action selector. If the label is empty, the name is shown.

## Action Request

The SDK parses selected IDs, loads the corresponding objects, and calls your action.

```go
type ActionRequest[T any, ID comparable] struct {
    Resource admin.Resource[T, ID]
    IDs      []ID
    Objects  []T
}
```

Use `IDs` when you only need identifiers. Use `Objects` when the operation depends on existing object values.

## Action Result

```go
type ActionResult struct {
    Message string `json:"message"`
}
```

The JSON API returns the action result body. The built-in HTML list page redirects back to the list after the action runs.

## HTML Route

With `BasePath: "/admin"`, a product action is posted to:

```text
POST /admin/catalog/products/actions/deactivate
```

The form body contains:

| Field | Meaning |
| --- | --- |
| `ids` | One or more selected object IDs. |
| `go_admin_csrf` | Built-in form CSRF token. |

The HTML route requires the built-in CSRF token unless `DisableCSRF` is true.

## JSON API Route

API clients run the same action with JSON:

```text
POST /admin/api/v1/catalog/products/actions/deactivate
Content-Type: application/json

{"ids":[1,2,3]}
```

The response is the `ActionResult`:

```json
{"message":"products deactivated"}
```

## Error Handling

If an ID cannot be parsed, the route returns an error. If a selected object does not exist, `admin.ErrNotFound` becomes `404`. Any other returned error becomes a server error.

Keep action work idempotent where possible, especially when clients may retry requests.

## Confirm Metadata

`Action.Confirm` is included in JSON action metadata for clients that build their own UI. The current built-in HTML list page exposes the action selector and submit button; confirmation UX can be implemented by a custom client on top of the JSON API metadata.

