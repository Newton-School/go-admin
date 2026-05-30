package admin

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRepositoryKeepsPublicRootSmall(t *testing.T) {
	forbiddenRootFiles := []string{
		"api.go",
		"field.go",
		"id.go",
		"memory.go",
		"query.go",
		"resource.go",
		"runtime.go",
		"site.go",
		"templates.go",
	}
	for _, name := range forbiddenRootFiles {
		if _, err := os.Stat(name); err == nil {
			t.Fatalf("implementation file %s should live under internal/core, not the module root", name)
		}
	}

	requiredInternalFiles := []string{
		"internal/core/api.go",
		"internal/core/field.go",
		"internal/core/site.go",
		"internal/core/assets/templates/list.tmpl",
		"internal/core/assets/static/admin.css",
	}
	for _, name := range requiredInternalFiles {
		if _, err := os.Stat(filepath.Clean(name)); err != nil {
			t.Fatalf("expected professional internal layout file %s: %v", name, err)
		}
	}
}
