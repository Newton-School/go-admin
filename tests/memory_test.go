package admin_test

import (
	"context"
	. "github.com/Newton-School/go-admin"
	"testing"
)

type memoryProduct struct {
	ID     int64
	Name   string
	Active bool
}

func TestMemoryRepositoryCreatesUpdatesDeletesAndLists(t *testing.T) {
	var next int64
	repo := NewMemoryRepository(MemoryRepositoryConfig[memoryProduct, int64]{
		GetID: func(p memoryProduct) int64 { return p.ID },
		SetID: func(p *memoryProduct, id int64) { p.ID = id },
		NextID: func() int64 {
			next++
			return next
		},
		Search: func(p memoryProduct, term string) bool {
			return stringsContainsFold(p.Name, term)
		},
		Filter: func(p memoryProduct, name string, values []string) bool {
			return name == "active" && boolInValues(p.Active, values)
		},
		Less: func(a, b memoryProduct, field string) bool {
			return field == "name" && a.Name < b.Name
		},
	})

	chair, err := repo.Create(context.Background(), memoryProduct{Name: "Chair", Active: true})
	if err != nil {
		t.Fatalf("create chair: %v", err)
	}
	if chair.ID != 1 {
		t.Fatalf("expected generated id 1, got %d", chair.ID)
	}
	if _, err := repo.Create(context.Background(), memoryProduct{Name: "Desk", Active: false}); err != nil {
		t.Fatalf("create desk: %v", err)
	}
	if _, err := repo.Create(context.Background(), memoryProduct{Name: "Table", Active: true}); err != nil {
		t.Fatalf("create table: %v", err)
	}

	page, err := repo.List(context.Background(), Query{
		Search:  "a",
		Filters: map[string][]string{"active": {"true"}},
		Sort:    []SortField{{Field: "name", Desc: true}},
		Page:    1,
		PerPage: 1,
	})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if page.Total != 2 || len(page.Items) != 1 || page.Items[0].Name != "Table" {
		t.Fatalf("unexpected page: %#v", page)
	}

	updated, err := repo.Update(context.Background(), chair.ID, memoryProduct{ID: chair.ID, Name: "Armchair", Active: true})
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if updated.Name != "Armchair" {
		t.Fatalf("unexpected updated object: %#v", updated)
	}

	if err := repo.Delete(context.Background(), chair.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, err := repo.Get(context.Background(), chair.ID); err == nil {
		t.Fatal("expected not found after delete")
	}
}
