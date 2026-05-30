package core

import (
	"context"
	"errors"
	"sort"
	"strings"
	"sync"
)

// ErrNotFound is returned when a repository cannot find an object.
var ErrNotFound = errors.New("admin: not found")

// MemoryRepositoryConfig configures the in-memory repository used by examples and tests.
type MemoryRepositoryConfig[T any, ID comparable] struct {
	GetID  func(T) ID
	SetID  func(*T, ID)
	NextID func() ID
	Search func(T, string) bool
	Filter func(T, string, []string) bool
	Less   func(T, T, string) bool
}

// MemoryRepository is a concurrency-safe Repository implementation.
type MemoryRepository[T any, ID comparable] struct {
	mu     sync.RWMutex
	config MemoryRepositoryConfig[T, ID]
	items  map[ID]T
}

// NewMemoryRepository creates an empty in-memory repository.
func NewMemoryRepository[T any, ID comparable](config MemoryRepositoryConfig[T, ID]) *MemoryRepository[T, ID] {
	return &MemoryRepository[T, ID]{
		config: config,
		items:  map[ID]T{},
	}
}

// List returns a searched, filtered, sorted, and paginated page.
func (m *MemoryRepository[T, ID]) List(_ context.Context, query Query) (Page[T], error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	items := make([]T, 0, len(m.items))
	for _, item := range m.items {
		if query.Search != "" && m.config.Search != nil && !m.config.Search(item, query.Search) {
			continue
		}
		if !m.matchesFilters(item, query.Filters) {
			continue
		}
		items = append(items, item)
	}

	if len(query.Sort) > 0 && m.config.Less != nil {
		sortField := query.Sort[0]
		sort.SliceStable(items, func(i, j int) bool {
			if sortField.Desc {
				return m.config.Less(items[j], items[i], sortField.Field)
			}
			return m.config.Less(items[i], items[j], sortField.Field)
		})
	}

	page := firstPositive(query.Page, 1)
	perPage := firstPositive(query.PerPage, defaultPerPage)
	total := len(items)
	start := (page - 1) * perPage
	if start > total {
		start = total
	}
	end := start + perPage
	if end > total {
		end = total
	}

	return Page[T]{
		Items:   append([]T(nil), items[start:end]...),
		Total:   total,
		Page:    page,
		PerPage: perPage,
	}, nil
}

// Get returns one object by ID.
func (m *MemoryRepository[T, ID]) Get(_ context.Context, id ID) (T, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	item, ok := m.items[id]
	if !ok {
		var zero T
		return zero, ErrNotFound
	}
	return item, nil
}

// Create stores a new object and assigns an ID when configured.
func (m *MemoryRepository[T, ID]) Create(_ context.Context, item T) (T, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	id := m.config.GetID(item)
	var zero ID
	if id == zero && m.config.NextID != nil && m.config.SetID != nil {
		id = m.config.NextID()
		m.config.SetID(&item, id)
	}
	m.items[id] = item
	return item, nil
}

// Update replaces an existing object.
func (m *MemoryRepository[T, ID]) Update(_ context.Context, id ID, item T) (T, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.items[id]; !ok {
		var zero T
		return zero, ErrNotFound
	}
	if m.config.SetID != nil {
		m.config.SetID(&item, id)
	}
	m.items[id] = item
	return item, nil
}

// Delete removes an object by ID.
func (m *MemoryRepository[T, ID]) Delete(_ context.Context, id ID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.items[id]; !ok {
		return ErrNotFound
	}
	delete(m.items, id)
	return nil
}

func (m *MemoryRepository[T, ID]) matchesFilters(item T, filters map[string][]string) bool {
	if len(filters) == 0 || m.config.Filter == nil {
		return true
	}
	for name, values := range filters {
		if !m.config.Filter(item, name, values) {
			return false
		}
	}
	return true
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
