# Repositories

`go-admin` does not require a specific ORM. Each resource uses this contract:

```go
type Repository[T any, ID comparable] interface {
    List(context.Context, admin.Query) (admin.Page[T], error)
    Get(context.Context, ID) (T, error)
    Create(context.Context, T) (T, error)
    Update(context.Context, ID, T) (T, error)
    Delete(context.Context, ID) error
}
```

`Query` contains normalized search, filters, sort fields, page, and page size. Repositories decide how those map to SQL, an ORM, an API, or another data source.

Use `admin.NewMemoryRepository` for demos, tests, and contract checks.

