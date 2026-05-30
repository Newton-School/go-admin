# Fields And Widgets

Fields describe both display and form behavior:

```go
[]admin.Field{
    admin.Int64("id", "ID").Readonly(),
    admin.Text("name", "Name").Required().Help("Shown in admin lists"),
    admin.Bool("active", "Active"),
    admin.JSON("metadata", "Metadata"),
}
```

Struct fields are matched by Go name, lower-case name, snake_case name, `json`, `db`, or `form` tag.

Built-ins:

- `Text`, `Textarea`
- `Int`, `Int64`, `Float`
- `Bool`
- `Date`, `DateTime`
- `JSON`
- `Enum`, `Select`, `MultiSelect`
- `Relation`

