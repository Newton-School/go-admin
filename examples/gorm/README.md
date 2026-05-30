# GORM Adapter Sketch

The core package is ORM-agnostic. A GORM adapter can satisfy `admin.Repository[T, ID]` by translating `admin.Query` into GORM clauses:

- `Query.Search` maps to `LIKE`/full-text expressions chosen by the adapter user.
- `Query.Filters` maps to whitelisted column filters.
- `Query.Sort` maps to whitelisted `ORDER BY` clauses.
- `Page` and `PerPage` map to `LIMIT` and `OFFSET`.

Keep adapter packages outside the core dependency graph so projects that do not use GORM do not inherit it.

