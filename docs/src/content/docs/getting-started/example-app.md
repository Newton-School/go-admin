---
title: Example App
description: Run the bundled multi-app admin example.
---

The repository includes a memory-backed example that registers several apps and resources.

## Run It

From the repository root:

```bash
go run ./examples/memory
```

Open:

```text
http://localhost:8080/admin/
```

## Included Apps

The example is intentionally broader than a toy product table:

| App | Resources |
| --- | --- |
| Catalog | Categories, Products |
| Sales | Customers, Orders |
| Content | Articles |
| Operations | Warehouses |

Use it as a reference for multiple apps, search, filters, default sorting, enum fields, date/time fields, and a bulk action.

## net/http Auth Example

The `examples/nethttp` package shows how to mount the admin behind a simple middleware:

```bash
go run ./examples/nethttp
```

That example expects requests to include:

```text
X-Admin: true
```

It is a development example only. Replace it with real authentication before exposing an admin panel.

