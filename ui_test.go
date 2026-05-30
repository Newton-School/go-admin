package admin

import (
	"os"
	"strings"
	"testing"
)

func TestAdminUsesPlainDjangoLikeMarkup(t *testing.T) {
	site, _ := newHandlerTestSite(t)

	index := performRequest(site.Handler(), "GET", "/admin/", nil, nil)
	if index.Code != 200 {
		t.Fatalf("index status = %d; body=%s", index.Code, index.Body.String())
	}
	assertBodyContains(t, index.Body.String(),
		`id="header"`,
		`id="branding"`,
		`id="content"`,
		`class="module"`,
		`<caption>`,
	)
	assertBodyNotContains(t, index.Body.String(), `ga-grid`, `ga-resource`, `ga-title-row`)

	list := performRequest(site.Handler(), "GET", "/admin/catalog/products/", nil, nil)
	if list.Code != 200 {
		t.Fatalf("list status = %d; body=%s", list.Code, list.Body.String())
	}
	assertBodyContains(t, list.Body.String(),
		`class="breadcrumbs"`,
		`class="object-tools"`,
		`id="changelist"`,
		`class="results"`,
	)
	assertBodyNotContains(t, list.Body.String(), `ga-button`, `ga-toolbar`, `ga-table`)
}

func TestAdminCSSAvoidsFancyLayoutStyling(t *testing.T) {
	css, err := os.ReadFile("internal/static/admin.css")
	if err != nil {
		t.Fatalf("read css: %v", err)
	}
	body := string(css)
	for _, forbidden := range []string{
		"border-radius",
		"box-shadow",
		"linear-gradient",
		"grid-template",
		"--ga-",
	} {
		if strings.Contains(body, forbidden) {
			t.Fatalf("expected simple css without %q; css=%s", forbidden, body)
		}
	}
}

func assertBodyNotContains(t *testing.T, body string, values ...string) {
	t.Helper()
	for _, value := range values {
		if strings.Contains(body, value) {
			t.Fatalf("expected body not to contain %q; body=%s", value, body)
		}
	}
}
