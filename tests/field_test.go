package admin_test

import (
	. "github.com/Newton-School/go-admin"
	"html/template"
	"net/url"
	"strings"
	"testing"
	"time"
)

type fieldSample struct {
	ID          int64
	Name        string `json:"name"`
	Active      bool
	Count       int
	Price       float64
	PublishedAt time.Time `json:"published_at"`
	Metadata    map[string]any
	Tags        []string
	Status      string
}

func TestBindFormSetsSupportedFieldTypesAndSkipsReadonlyFields(t *testing.T) {
	obj := fieldSample{ID: 99}
	values := url.Values{
		"id":           {"1"},
		"name":         {"Chair"},
		"active":       {"on"},
		"count":        {"7"},
		"price":        {"12.5"},
		"published_at": {"2026-05-30T12:45"},
		"metadata":     {`{"sku":"C-1"}`},
		"tags":         {"new", "featured"},
		"status":       {"published"},
	}

	errs := BindForm([]Field{
		Int64("id", "ID").Readonly(),
		Text("name", "Name").Required(),
		Bool("active", "Active"),
		Int("count", "Count"),
		Float("price", "Price"),
		DateTime("published_at", "Published At"),
		JSON("metadata", "Metadata"),
		MultiSelect("tags", "Tags").Options([]Choice{{Value: "new", Label: "New"}, {Value: "featured", Label: "Featured"}}),
		Enum("status", "Status").Options([]Choice{{Value: "draft", Label: "Draft"}, {Value: "published", Label: "Published"}}),
	}, values, &obj)
	if !errs.Empty() {
		t.Fatalf("expected no validation errors, got %#v", errs)
	}

	if obj.ID != 99 {
		t.Fatalf("readonly id should stay 99, got %d", obj.ID)
	}
	if obj.Name != "Chair" || !obj.Active || obj.Count != 7 || obj.Price != 12.5 {
		t.Fatalf("unexpected scalar values: %#v", obj)
	}
	if obj.PublishedAt.Format("2006-01-02T15:04") != "2026-05-30T12:45" {
		t.Fatalf("unexpected published_at: %s", obj.PublishedAt.Format(time.RFC3339))
	}
	if obj.Metadata["sku"] != "C-1" {
		t.Fatalf("unexpected metadata: %#v", obj.Metadata)
	}
	if strings.Join(obj.Tags, ",") != "new,featured" {
		t.Fatalf("unexpected tags: %#v", obj.Tags)
	}
	if obj.Status != "published" {
		t.Fatalf("unexpected status: %q", obj.Status)
	}
}

func TestBindFormReturnsFieldErrors(t *testing.T) {
	var obj fieldSample
	errs := BindForm([]Field{
		Text("name", "Name").Required(),
		Int("count", "Count"),
		JSON("metadata", "Metadata"),
	}, url.Values{
		"name":     {""},
		"count":    {"nan"},
		"metadata": {"{"},
	}, &obj)

	if errs.Empty() {
		t.Fatal("expected validation errors")
	}
	if errs.Get("name") == "" {
		t.Fatal("expected required error for name")
	}
	if errs.Get("count") == "" {
		t.Fatal("expected parse error for count")
	}
	if errs.Get("metadata") == "" {
		t.Fatal("expected parse error for metadata")
	}
}

func TestRenderWidgetEscapesHTMLValues(t *testing.T) {
	html := RenderWidget(WidgetContext{
		Field: Text("name", "Name").Required().Placeholder("Product name").Help("Shown in the list"),
		Value: `<script>alert("x")</script>`,
	})

	if strings.Contains(string(html), "<script>") {
		t.Fatalf("expected escaped script tag, got %s", html)
	}
	if !strings.Contains(string(html), "Product name") {
		t.Fatalf("expected placeholder in widget html, got %s", html)
	}
	if _, ok := any(html).(template.HTML); !ok {
		t.Fatal("expected RenderWidget to return template.HTML")
	}
}
