package admin

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

func (s *Site) handleAPI(w http.ResponseWriter, r *http.Request, parts []string) {
	if len(parts) == 0 || parts[0] != "v1" {
		writeJSONError(w, http.StatusNotFound, "api version not found")
		return
	}
	parts = parts[1:]
	if len(parts) == 1 && parts[0] == "apps" {
		if r.Method != http.MethodGet {
			writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"apps": s.Apps()})
		return
	}
	if len(parts) < 2 {
		writeJSONError(w, http.StatusNotFound, "resource not found")
		return
	}

	app, resource, ok := s.lookup(parts[0], parts[1])
	if !ok {
		writeJSONError(w, http.StatusNotFound, "resource not found")
		return
	}

	switch {
	case len(parts) == 2:
		s.handleAPICollection(w, r, app, resource)
	case len(parts) == 4 && parts[2] == "actions":
		s.handleAPIAction(w, r, resource, parts[3])
	case len(parts) == 4 && parts[2] == "lookup":
		s.handleAPILookup(w, r, resource, parts[3])
	case len(parts) == 3:
		s.handleAPIObject(w, r, resource, parts[2])
	default:
		writeJSONError(w, http.StatusNotFound, "route not found")
	}
	_ = app
}

func (s *Site) handleAPICollection(w http.ResponseWriter, r *http.Request, _ *App, resource resourceRuntime) {
	switch r.Method {
	case http.MethodGet:
		config := resource.listConfig()
		query := QueryFromRequest(r, QueryConfig{
			DefaultPerPage: defaultPerPage,
			MaxPerPage:     defaultMaxPage,
			AllowedSorts:   allowedSorts(config),
			FilterNames:    filterNames(config.Filters),
		})
		page, err := resource.list(r.Context(), query)
		if err != nil {
			writeJSONRepositoryError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, page)
	case http.MethodPost:
		values, ok := readJSONObject(w, r)
		if !ok {
			return
		}
		obj, errs, err := resource.createJSON(r.Context(), values)
		if err != nil {
			writeJSONRepositoryError(w, err)
			return
		}
		if !errs.Empty() {
			writeJSON(w, http.StatusUnprocessableEntity, map[string]any{"errors": errs})
			return
		}
		writeJSON(w, http.StatusCreated, obj)
	default:
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (s *Site) handleAPIObject(w http.ResponseWriter, r *http.Request, resource resourceRuntime, rawID string) {
	switch r.Method {
	case http.MethodGet:
		obj, err := resource.get(r.Context(), rawID)
		if err != nil {
			writeJSONRepositoryError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, obj)
	case http.MethodPatch, http.MethodPut:
		values, ok := readJSONObject(w, r)
		if !ok {
			return
		}
		obj, errs, err := resource.updateJSON(r.Context(), rawID, values)
		if err != nil {
			writeJSONRepositoryError(w, err)
			return
		}
		if !errs.Empty() {
			writeJSON(w, http.StatusUnprocessableEntity, map[string]any{"errors": errs})
			return
		}
		writeJSON(w, http.StatusOK, obj)
	case http.MethodDelete:
		if err := resource.delete(r.Context(), rawID); err != nil {
			writeJSONRepositoryError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	default:
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (s *Site) handleAPIAction(w http.ResponseWriter, r *http.Request, resource resourceRuntime, actionName string) {
	if r.Method != http.MethodPost {
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var payload struct {
		IDs []any `json:"ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid json")
		return
	}
	ids := make([]string, 0, len(payload.IDs))
	for _, id := range payload.IDs {
		ids = append(ids, stringifyID(id))
	}
	result, err := resource.runAction(r.Context(), actionName, ids)
	if err != nil {
		writeJSONRepositoryError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (s *Site) handleAPILookup(w http.ResponseWriter, r *http.Request, resource resourceRuntime, fieldName string) {
	if r.Method != http.MethodGet {
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	query := QueryFromRequest(r, QueryConfig{DefaultPerPage: 20, MaxPerPage: 50})
	choices, err := resource.lookup(r.Context(), fieldName, query)
	if err != nil {
		writeJSONRepositoryError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"results": choices})
}

func readJSONObject(w http.ResponseWriter, r *http.Request) (map[string]any, bool) {
	defer r.Body.Close()
	var values map[string]any
	if err := json.NewDecoder(r.Body).Decode(&values); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid json")
		return nil, false
	}
	if values == nil {
		values = map[string]any{}
	}
	return values, true
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if status == http.StatusNoContent {
		return
	}
	_ = json.NewEncoder(w).Encode(body)
}

func writeJSONError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func writeJSONRepositoryError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrNotFound):
		writeJSONError(w, http.StatusNotFound, "not found")
	default:
		writeJSONError(w, http.StatusInternalServerError, err.Error())
	}
}

func stringifyID(value any) string {
	switch typed := value.(type) {
	case string:
		return typed
	case float64:
		if typed == float64(int64(typed)) {
			return strconv.FormatInt(int64(typed), 10)
		}
		return strconv.FormatFloat(typed, 'f', -1, 64)
	default:
		return strings.TrimSpace(fmt.Sprint(value))
	}
}
