package admin

import (
	"net/http"
	"strconv"
	"strings"
)

const (
	defaultPerPage = 25
	defaultMaxPage = 100
)

// SortField describes one requested sort column.
type SortField struct {
	Field string
	Desc  bool
}

// Query is the normalized list request passed to repositories.
type Query struct {
	Search  string
	Filters map[string][]string
	Sort    []SortField
	Page    int
	PerPage int
}

// Page is a repository result page.
type Page[T any] struct {
	Items   []T `json:"items"`
	Total   int `json:"total"`
	Page    int `json:"page"`
	PerPage int `json:"per_page"`
}

// QueryConfig controls how HTTP query parameters are accepted.
type QueryConfig struct {
	DefaultPerPage int
	MaxPerPage     int
	AllowedSorts   []string
	FilterNames    []string
}

// QueryFromRequest parses a list request into a safe repository query.
func QueryFromRequest(r *http.Request, cfg QueryConfig) Query {
	values := r.URL.Query()
	perPage := parsePositiveInt(values.Get("per_page"), firstPositive(cfg.DefaultPerPage, defaultPerPage))
	maxPerPage := firstPositive(cfg.MaxPerPage, defaultMaxPage)
	if perPage > maxPerPage {
		perPage = maxPerPage
	}

	query := Query{
		Search:  strings.TrimSpace(values.Get("q")),
		Filters: map[string][]string{},
		Page:    parsePositiveInt(values.Get("page"), 1),
		PerPage: perPage,
	}

	allowedSorts := stringSet(cfg.AllowedSorts)
	if rawSort := strings.TrimSpace(values.Get("sort")); rawSort != "" {
		desc := strings.HasPrefix(rawSort, "-")
		field := strings.TrimPrefix(rawSort, "-")
		if allowedSorts[field] {
			query.Sort = append(query.Sort, SortField{Field: field, Desc: desc})
		}
	}

	for _, name := range cfg.FilterNames {
		if selected, ok := values[name]; ok {
			query.Filters[name] = append([]string(nil), selected...)
		}
	}

	return query
}

func parsePositiveInt(raw string, fallback int) int {
	if raw == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(raw)
	if err != nil || parsed < 1 {
		return fallback
	}
	return parsed
}

func firstPositive(values ...int) int {
	for _, value := range values {
		if value > 0 {
			return value
		}
	}
	return 1
}

func stringSet(values []string) map[string]bool {
	set := make(map[string]bool, len(values))
	for _, value := range values {
		set[value] = true
	}
	return set
}
