package admin

import (
	"context"
	"errors"
	"testing"
)

type coreProduct struct {
	ID   int64
	Name string
}

type coreRepo[T any, ID comparable] struct{}

func (coreRepo[T, ID]) List(context.Context, Query) (Page[T], error) {
	return Page[T]{}, nil
}

func (coreRepo[T, ID]) Get(context.Context, ID) (T, error) {
	var zero T
	return zero, errors.New("not implemented")
}

func (coreRepo[T, ID]) Create(context.Context, T) (T, error) {
	var zero T
	return zero, nil
}

func (coreRepo[T, ID]) Update(context.Context, ID, T) (T, error) {
	var zero T
	return zero, nil
}

func (coreRepo[T, ID]) Delete(context.Context, ID) error {
	return nil
}

func TestSiteRegistersAppsAndResources(t *testing.T) {
	site := New(SiteConfig{Title: "Acme Admin", BasePath: "admin"})
	app := site.App("catalog", "Catalog")

	err := Register(app, Resource[coreProduct, int64]{
		Name:       "products",
		Label:      "Products",
		Repository: coreRepo[coreProduct, int64]{},
		ID:         Int64ID(),
		Fields: []Field{
			Int64("id", "ID").Readonly(),
			Text("name", "Name").Required(),
		},
	})
	if err != nil {
		t.Fatalf("register resource: %v", err)
	}

	if site.BasePath() != "/admin" {
		t.Fatalf("expected normalized base path /admin, got %q", site.BasePath())
	}

	apps := site.Apps()
	if len(apps) != 1 {
		t.Fatalf("expected 1 app, got %d", len(apps))
	}
	if apps[0].Name != "catalog" || apps[0].Label != "Catalog" {
		t.Fatalf("unexpected app meta: %#v", apps[0])
	}
	if len(apps[0].Resources) != 1 {
		t.Fatalf("expected 1 resource, got %d", len(apps[0].Resources))
	}
	if apps[0].Resources[0].Name != "products" || apps[0].Resources[0].Label != "Products" {
		t.Fatalf("unexpected resource meta: %#v", apps[0].Resources[0])
	}
}

func TestRegisterRejectsInvalidAndDuplicateResources(t *testing.T) {
	site := New(SiteConfig{})
	app := site.App("catalog", "Catalog")

	valid := Resource[coreProduct, int64]{
		Name:       "products",
		Label:      "Products",
		Repository: coreRepo[coreProduct, int64]{},
		ID:         Int64ID(),
		Fields:     []Field{Text("name", "Name")},
	}

	if err := Register(app, valid); err != nil {
		t.Fatalf("first register: %v", err)
	}
	if err := Register(app, valid); err == nil {
		t.Fatal("expected duplicate resource error")
	}

	invalid := valid
	invalid.Name = "Bad Products"
	if err := Register(app, invalid); err == nil {
		t.Fatal("expected invalid slug error")
	}

	missingRepo := valid
	missingRepo.Name = "orders"
	missingRepo.Repository = nil
	if err := Register(app, missingRepo); err == nil {
		t.Fatal("expected missing repository error")
	}
}
