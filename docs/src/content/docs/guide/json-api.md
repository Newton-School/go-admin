---
title: JSON API
description: Use the built-in API for admin clients and integrations.
---

Every admin site exposes a JSON API under the same `BasePath` as the HTML admin.

With `BasePath: "/admin"`, API routes start at:

```text
/admin/api/v1
```

Protect this API with the same middleware that protects the HTML admin.

## Apps

List registered apps and resources:

```text
GET /admin/api/v1/apps
```

Response:

```json
{
  "apps": [
    {
      "name": "catalog",
      "label": "Catalog",
      "resources": [
        {"name": "products", "label": "Products"}
      ]
    }
  ]
}
```

## List Objects

```text
GET /admin/api/v1/{app}/{resource}
```

Example:

```text
GET /admin/api/v1/catalog/products?q=chair&page=1&per_page=25&sort=-price&active=true
```

Response shape:

```json
{
  "items": [
    {
      "ID": 1,
      "Name": "Chair",
      "sku": "CHR-001",
      "Price": 249.99,
      "Active": true
    }
  ],
  "total": 1,
  "page": 1,
  "per_page": 25
}
```

Objects are encoded by Go's standard JSON encoder. Add `json` tags to your model when you want stable API field names.

## Create Object

```text
POST /admin/api/v1/{app}/{resource}
Content-Type: application/json
```

Body:

```json
{
  "name": "Lamp",
  "sku": "LMP-001",
  "price": 89.5,
  "active": true
}
```

Create requests use full JSON binding. Required fields must be present and non-empty.

Successful response:

```text
201 Created
```

The response body is the created object.

## Get Object

```text
GET /admin/api/v1/{app}/{resource}/{id}
```

Example:

```text
GET /admin/api/v1/catalog/products/1
```

## Update Object

```text
PATCH /admin/api/v1/{app}/{resource}/{id}
PUT /admin/api/v1/{app}/{resource}/{id}
Content-Type: application/json
```

Body:

```json
{
  "name": "Tall Lamp"
}
```

`PATCH` and `PUT` both use partial binding. Missing fields keep their existing values. Readonly fields are ignored.

Successful response:

```text
200 OK
```

The response body is the updated object.

## Delete Object

```text
DELETE /admin/api/v1/{app}/{resource}/{id}
```

Successful response:

```text
204 No Content
```

## Run Action

```text
POST /admin/api/v1/{app}/{resource}/actions/{action}
Content-Type: application/json
```

Body:

```json
{"ids":[1,2,3]}
```

Response:

```json
{"message":"done"}
```

IDs can be JSON strings or numbers. The resource ID codec parses the final string value.

## Lookup Choices

Lookup routes return `{value,label}` pairs for a field.

```text
GET /admin/api/v1/{app}/{resource}/lookup/{field}?q=cha
```

Response:

```json
{
  "results": [
    {"value": "1", "label": "Chair"}
  ]
}
```

The lookup route calls the resource repository `List` method with a small query and formats each object's ID as `value`. The `label` is the formatted value of the requested field.

## Status Codes

| Status | Meaning |
| --- | --- |
| `200` | Successful read, update, action, or lookup. |
| `201` | Object created. |
| `204` | Object deleted. |
| `400` | Invalid JSON. |
| `404` | API version, route, resource, action, or object not found. |
| `405` | Method not allowed. |
| `422` | Field validation failed. |
| `500` | Repository or action returned an unexpected error. |

Validation errors use this shape:

```json
{
  "errors": {
    "name": "Name is required"
  }
}
```

Other errors use this shape:

```json
{"error":"not found"}
```

## CSRF And Auth

The JSON API does not use the HTML form CSRF token. Protect the mounted handler with your own auth middleware and any API-specific CSRF policy your application requires.

