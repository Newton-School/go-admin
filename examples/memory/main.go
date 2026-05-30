package main

import (
	"context"
	"log"
	"net/http"
	"strings"
	"time"

	admin "github.com/Newton-School/go-admin"
)

type Product struct {
	ID         int64
	Name       string
	CategoryID int64  `json:"category_id"`
	SKU        string `json:"sku"`
	Price      float64
	Stock      int
	Active     bool
	CreatedAt  time.Time `json:"created_at"`
}

type Category struct {
	ID          int64
	Name        string
	Description string
	Active      bool
}

type Customer struct {
	ID        int64
	Name      string
	Email     string
	Segment   string
	Active    bool
	CreatedAt time.Time `json:"created_at"`
}

type Order struct {
	ID         int64
	Number     string
	CustomerID int64 `json:"customer_id"`
	Status     string
	Total      float64
	Paid       bool
	PlacedAt   time.Time `json:"placed_at"`
}

type Article struct {
	ID          int64
	Title       string
	Slug        string
	Status      string
	Author      string
	PublishedAt time.Time `json:"published_at"`
}

type Warehouse struct {
	ID       int64
	Name     string
	Location string
	Capacity int
	Active   bool
}

func main() {
	site := newDemoSite()

	http.Handle("/admin/", site.Handler())
	http.Handle("/admin", site.Handler())
	log.Println("admin listening at http://localhost:8080/admin/")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func newDemoSite() *admin.Site {
	site := admin.New(admin.SiteConfig{Title: "Acme Admin", BasePath: "/admin"})

	catalog := site.App("catalog", "Catalog")
	categories := newRepo(
		func(c Category) int64 { return c.ID },
		func(c *Category, id int64) { c.ID = id },
		func(c Category) string { return c.Name + " " + c.Description },
		activeFilter[Category](func(c Category) bool { return c.Active }),
		func(a, b Category, field string) bool { return compareStrings(a.Name, b.Name, field, "name") },
	)
	seed(categories,
		Category{Name: "Furniture", Description: "Office furniture and fixtures", Active: true},
		Category{Name: "Accessories", Description: "Desk accessories and small items", Active: true},
	)
	mustRegister(admin.Register(catalog, admin.Resource[Category, int64]{
		Name:       "categories",
		Label:      "Categories",
		Repository: categories,
		ID:         admin.Int64ID(),
		Fields: []admin.Field{
			admin.Int64("id", "ID").Readonly(),
			admin.Text("name", "Name").Required(),
			admin.Textarea("description", "Description"),
			admin.Bool("active", "Active"),
		},
		List: admin.ListConfig{
			Columns: []string{"id", "name", "active"},
			Search:  []string{"name", "description"},
			Sort:    []admin.SortField{{Field: "name"}},
			Filters: []admin.Filter{{Name: "active", Label: "Active"}},
		},
	}))

	products := newRepo(
		func(p Product) int64 { return p.ID },
		func(p *Product, id int64) { p.ID = id },
		func(p Product) string { return p.Name + " " + p.SKU },
		activeFilter[Product](func(p Product) bool { return p.Active }),
		func(a, b Product, field string) bool {
			switch field {
			case "name":
				return a.Name < b.Name
			case "price":
				return a.Price < b.Price
			default:
				return a.ID < b.ID
			}
		},
	)
	seed(products,
		Product{Name: "Ergonomic Chair", CategoryID: 1, SKU: "CHR-ERG-001", Price: 249.99, Stock: 42, Active: true, CreatedAt: daysAgo(18)},
		Product{Name: "Standing Desk", CategoryID: 1, SKU: "DSK-STD-002", Price: 599.00, Stock: 15, Active: true, CreatedAt: daysAgo(11)},
		Product{Name: "Cable Tray", CategoryID: 2, SKU: "ACC-CBL-003", Price: 39.50, Stock: 80, Active: false, CreatedAt: daysAgo(6)},
	)
	mustRegister(admin.Register(catalog, admin.Resource[Product, int64]{
		Name:       "products",
		Label:      "Products",
		Repository: products,
		ID:         admin.Int64ID(),
		Fields: []admin.Field{
			admin.Int64("id", "ID").Readonly(),
			admin.Text("name", "Name").Required(),
			admin.Int64("category_id", "Category ID"),
			admin.Text("sku", "SKU").Required(),
			admin.Float("price", "Price"),
			admin.Int("stock", "Stock"),
			admin.Bool("active", "Active"),
			admin.DateTime("created_at", "Created At").Readonly(),
		},
		List: admin.ListConfig{
			Columns: []string{"id", "name", "sku", "price", "stock", "active"},
			Search:  []string{"name", "sku"},
			Sort:    []admin.SortField{{Field: "name"}},
			Filters: []admin.Filter{{Name: "active", Label: "Active"}},
		},
		Actions: []admin.Action[Product, int64]{
			{
				Name:  "deactivate",
				Label: "Deactivate selected products",
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
		},
	}))

	sales := site.App("sales", "Sales")
	customers := newRepo(
		func(c Customer) int64 { return c.ID },
		func(c *Customer, id int64) { c.ID = id },
		func(c Customer) string { return c.Name + " " + c.Email + " " + c.Segment },
		activeFilter[Customer](func(c Customer) bool { return c.Active }),
		func(a, b Customer, field string) bool { return compareStrings(a.Name, b.Name, field, "name") },
	)
	seed(customers,
		Customer{Name: "Asha Mehta", Email: "asha@example.com", Segment: "enterprise", Active: true, CreatedAt: daysAgo(90)},
		Customer{Name: "Noah Smith", Email: "noah@example.com", Segment: "startup", Active: true, CreatedAt: daysAgo(45)},
		Customer{Name: "Mira Kapoor", Email: "mira@example.com", Segment: "education", Active: false, CreatedAt: daysAgo(21)},
	)
	mustRegister(admin.Register(sales, admin.Resource[Customer, int64]{
		Name:       "customers",
		Label:      "Customers",
		Repository: customers,
		ID:         admin.Int64ID(),
		Fields: []admin.Field{
			admin.Int64("id", "ID").Readonly(),
			admin.Text("name", "Name").Required(),
			admin.Text("email", "Email").Required(),
			admin.Enum("segment", "Segment").Options([]admin.Choice{
				{Value: "enterprise", Label: "Enterprise"},
				{Value: "startup", Label: "Startup"},
				{Value: "education", Label: "Education"},
			}),
			admin.Bool("active", "Active"),
			admin.DateTime("created_at", "Created At").Readonly(),
		},
		List: admin.ListConfig{
			Columns: []string{"id", "name", "email", "segment", "active"},
			Search:  []string{"name", "email", "segment"},
			Sort:    []admin.SortField{{Field: "name"}},
			Filters: []admin.Filter{{Name: "active", Label: "Active"}},
		},
	}))

	orders := newRepo(
		func(o Order) int64 { return o.ID },
		func(o *Order, id int64) { o.ID = id },
		func(o Order) string { return o.Number + " " + o.Status },
		func(o Order, name string, values []string) bool {
			switch name {
			case "paid":
				return boolInValues(o.Paid, values)
			default:
				return true
			}
		},
		func(a, b Order, field string) bool {
			switch field {
			case "placed_at":
				return a.PlacedAt.Before(b.PlacedAt)
			case "total":
				return a.Total < b.Total
			default:
				return a.ID < b.ID
			}
		},
	)
	seed(orders,
		Order{Number: "SO-1001", CustomerID: 1, Status: "paid", Total: 849.99, Paid: true, PlacedAt: daysAgo(10)},
		Order{Number: "SO-1002", CustomerID: 2, Status: "pending", Total: 599.00, Paid: false, PlacedAt: daysAgo(4)},
		Order{Number: "SO-1003", CustomerID: 1, Status: "shipped", Total: 89.50, Paid: true, PlacedAt: daysAgo(2)},
	)
	mustRegister(admin.Register(sales, admin.Resource[Order, int64]{
		Name:       "orders",
		Label:      "Orders",
		Repository: orders,
		ID:         admin.Int64ID(),
		Fields: []admin.Field{
			admin.Int64("id", "ID").Readonly(),
			admin.Text("number", "Number").Required(),
			admin.Int64("customer_id", "Customer ID"),
			admin.Enum("status", "Status").Options([]admin.Choice{
				{Value: "pending", Label: "Pending"},
				{Value: "paid", Label: "Paid"},
				{Value: "shipped", Label: "Shipped"},
			}),
			admin.Float("total", "Total"),
			admin.Bool("paid", "Paid"),
			admin.DateTime("placed_at", "Placed At"),
		},
		List: admin.ListConfig{
			Columns: []string{"id", "number", "status", "total", "paid", "placed_at"},
			Search:  []string{"number", "status"},
			Sort:    []admin.SortField{{Field: "placed_at", Desc: true}},
			Filters: []admin.Filter{{Name: "paid", Label: "Paid"}},
		},
	}))

	content := site.App("content", "Content")
	articles := newRepo(
		func(a Article) int64 { return a.ID },
		func(a *Article, id int64) { a.ID = id },
		func(a Article) string { return a.Title + " " + a.Slug + " " + a.Author + " " + a.Status },
		func(a Article, name string, values []string) bool {
			switch name {
			case "status":
				return stringInValues(a.Status, values)
			default:
				return true
			}
		},
		func(a, b Article, field string) bool { return compareStrings(a.Title, b.Title, field, "title") },
	)
	seed(articles,
		Article{Title: "Workspace Setup Guide", Slug: "workspace-setup-guide", Status: "published", Author: "Editorial", PublishedAt: daysAgo(30)},
		Article{Title: "Ergonomics Checklist", Slug: "ergonomics-checklist", Status: "draft", Author: "Asha Mehta"},
	)
	mustRegister(admin.Register(content, admin.Resource[Article, int64]{
		Name:       "articles",
		Label:      "Articles",
		Repository: articles,
		ID:         admin.Int64ID(),
		Fields: []admin.Field{
			admin.Int64("id", "ID").Readonly(),
			admin.Text("title", "Title").Required(),
			admin.Text("slug", "Slug").Required(),
			admin.Enum("status", "Status").Options([]admin.Choice{
				{Value: "draft", Label: "Draft"},
				{Value: "published", Label: "Published"},
			}),
			admin.Text("author", "Author"),
			admin.DateTime("published_at", "Published At"),
		},
		List: admin.ListConfig{
			Columns: []string{"id", "title", "status", "author", "published_at"},
			Search:  []string{"title", "slug", "author"},
			Sort:    []admin.SortField{{Field: "title"}},
			Filters: []admin.Filter{{Name: "status", Label: "Status", Choices: []admin.Choice{
				{Value: "draft", Label: "Draft"},
				{Value: "published", Label: "Published"},
			}}},
		},
	}))

	operations := site.App("operations", "Operations")
	warehouses := newRepo(
		func(w Warehouse) int64 { return w.ID },
		func(w *Warehouse, id int64) { w.ID = id },
		func(w Warehouse) string { return w.Name + " " + w.Location },
		activeFilter[Warehouse](func(w Warehouse) bool { return w.Active }),
		func(a, b Warehouse, field string) bool { return compareStrings(a.Name, b.Name, field, "name") },
	)
	seed(warehouses,
		Warehouse{Name: "North Fulfillment", Location: "Delhi", Capacity: 12000, Active: true},
		Warehouse{Name: "West Storage", Location: "Mumbai", Capacity: 8000, Active: true},
	)
	mustRegister(admin.Register(operations, admin.Resource[Warehouse, int64]{
		Name:       "warehouses",
		Label:      "Warehouses",
		Repository: warehouses,
		ID:         admin.Int64ID(),
		Fields: []admin.Field{
			admin.Int64("id", "ID").Readonly(),
			admin.Text("name", "Name").Required(),
			admin.Text("location", "Location"),
			admin.Int("capacity", "Capacity"),
			admin.Bool("active", "Active"),
		},
		List: admin.ListConfig{
			Columns: []string{"id", "name", "location", "capacity", "active"},
			Search:  []string{"name", "location"},
			Sort:    []admin.SortField{{Field: "name"}},
			Filters: []admin.Filter{{Name: "active", Label: "Active"}},
		},
	}))

	return site
}

func newRepo[T any](
	getID func(T) int64,
	setID func(*T, int64),
	searchText func(T) string,
	filter func(T, string, []string) bool,
	less func(T, T, string) bool,
) *admin.MemoryRepository[T, int64] {
	var next int64
	return admin.NewMemoryRepository(admin.MemoryRepositoryConfig[T, int64]{
		GetID: getID,
		SetID: setID,
		NextID: func() int64 {
			next++
			return next
		},
		Search: func(item T, term string) bool {
			return stringsContainsFold(searchText(item), term)
		},
		Filter: filter,
		Less:   less,
	})
}

func activeFilter[T any](active func(T) bool) func(T, string, []string) bool {
	return func(item T, name string, values []string) bool {
		switch name {
		case "active":
			return boolInValues(active(item), values)
		default:
			return true
		}
	}
}

func seed[T any](repo *admin.MemoryRepository[T, int64], items ...T) {
	for _, item := range items {
		if _, err := repo.Create(context.Background(), item); err != nil {
			panic(err)
		}
	}
}

func mustRegister(err error) {
	if err != nil {
		panic(err)
	}
}

func daysAgo(days int) time.Time {
	return time.Now().AddDate(0, 0, -days).Truncate(time.Minute)
}

func compareStrings(a, b, field, expected string) bool {
	if field != expected {
		return false
	}
	return a < b
}

func stringsContainsFold(value, term string) bool {
	return strings.Contains(strings.ToLower(value), strings.ToLower(term))
}

func boolInValues(value bool, values []string) bool {
	for _, candidate := range values {
		normalized := strings.ToLower(strings.TrimSpace(candidate))
		if value && (normalized == "true" || normalized == "1" || normalized == "on") {
			return true
		}
		if !value && (normalized == "false" || normalized == "0" || normalized == "off") {
			return true
		}
	}
	return false
}

func stringInValues(value string, values []string) bool {
	if len(values) == 0 {
		return true
	}
	for _, candidate := range values {
		if value == candidate {
			return true
		}
	}
	return false
}
