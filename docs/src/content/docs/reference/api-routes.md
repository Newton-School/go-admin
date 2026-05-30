---
title: API Routes
description: Route reference for the HTML admin and JSON API.
---

The route tables below assume:

```go
admin.New(admin.SiteConfig{BasePath: "/admin"})
```

## HTML Routes

| Method | Route | Purpose |
| --- | --- | --- |
| `GET` | `/admin/` | Admin dashboard. |
| `GET` | `/admin/{app}/` | App dashboard. |
| `GET` | `/admin/{app}/{resource}/` | Resource list page. |
| `GET` | `/admin/{app}/{resource}/new` | Create form. |
| `POST` | `/admin/{app}/{resource}/new` | Create object. |
| `GET` | `/admin/{app}/{resource}/{id}` | Edit/detail form. |
| `POST` | `/admin/{app}/{resource}/{id}` | Update object. |
| `GET` | `/admin/{app}/{resource}/{id}/delete` | Delete confirmation page. |
| `POST` | `/admin/{app}/{resource}/{id}/delete` | Delete object. |
| `POST` | `/admin/{app}/{resource}/actions/{action}` | Run bulk action. |
| `GET` | `/admin/static/*` | Embedded admin assets. |

HTML mutating routes require `go_admin_csrf` unless `DisableCSRF` is true.

## JSON Routes

| Method | Route | Purpose |
| --- | --- | --- |
| `GET` | `/admin/api/v1/apps` | List registered apps and resources. |
| `GET` | `/admin/api/v1/{app}/{resource}` | List objects. |
| `POST` | `/admin/api/v1/{app}/{resource}` | Create object. |
| `GET` | `/admin/api/v1/{app}/{resource}/{id}` | Get object. |
| `PATCH` | `/admin/api/v1/{app}/{resource}/{id}` | Partially update object. |
| `PUT` | `/admin/api/v1/{app}/{resource}/{id}` | Partially update object. |
| `DELETE` | `/admin/api/v1/{app}/{resource}/{id}` | Delete object. |
| `POST` | `/admin/api/v1/{app}/{resource}/actions/{action}` | Run action. |
| `GET` | `/admin/api/v1/{app}/{resource}/lookup/{field}` | Return lookup choices. |

## List Query Parameters

| Parameter | Applies To | Description |
| --- | --- | --- |
| `q` | HTML and JSON list routes | Search term. |
| `page` | HTML and JSON list routes | 1-based page number. |
| `per_page` | HTML and JSON list routes | Rows per page, capped by query config. |
| `sort` | HTML and JSON list routes | Sort field. Prefix with `-` for descending. |
| Filter names | HTML and JSON list routes | Repeated filter values, such as `active=true&active=false`. |

## Lookup Query Parameters

| Parameter | Description |
| --- | --- |
| `q` | Search term passed to the repository. |
| `page` | 1-based page number. |
| `per_page` | Rows per page, capped at `50` for lookup routes. |

## Method Behavior

Unsupported methods return `405`. Unknown app names, resource names, object IDs, actions, and routes return `404` when they map to `admin.ErrNotFound` or cannot be resolved.

