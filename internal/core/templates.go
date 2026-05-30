package core

import (
	"crypto/rand"
	"embed"
	"encoding/hex"
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"time"
)

const csrfCookieName = "go_admin_csrf"

//go:embed assets/templates/*.tmpl assets/static/*
var embeddedFiles embed.FS

var adminTemplates = template.Must(template.New("admin").
	Funcs(template.FuncMap{
		"formatValue": formatValue,
		"safe": func(value template.HTML) template.HTML {
			return value
		},
	}).
	ParseFS(embeddedFiles, "assets/templates/*.tmpl"))

type pageData struct {
	SiteTitle string
	PageTitle string
	BasePath  string
	Apps      []AppMeta
	CSRFToken string
}

type listPageData struct {
	pageData
	App       AppMeta
	Resource  ResourceMeta
	Columns   []Field
	Rows      []listRow
	Search    string
	Filters   []filterView
	Page      pageView
	NewURL    string
	Actions   []ActionMeta
	ActionURL string
	SortValue string
}

type listRow struct {
	ID     string
	Detail string
	Cells  []cellView
}

type cellView struct {
	Field Field
	Value any
}

type filterView struct {
	Name    string
	Label   string
	Choices []Choice
	Values  []string
}

type pageView struct {
	Total   int
	Page    int
	PerPage int
}

type formPageData struct {
	pageData
	App       AppMeta
	Resource  ResourceMeta
	ActionURL string
	BackURL   string
	DeleteURL string
	IsNew     bool
	Fieldsets []fieldsetView
	Errors    ValidationErrors
}

type fieldsetView struct {
	Title       string
	Description string
	Collapsed   bool
	Fields      []fieldView
}

type fieldView struct {
	Field    Field
	Value    any
	Widget   template.HTML
	Readonly bool
	Display  string
}

type deletePageData struct {
	pageData
	App       AppMeta
	Resource  ResourceMeta
	ActionURL string
	BackURL   string
	ObjectID  string
}

func (s *Site) render(w http.ResponseWriter, r *http.Request, name string, data any) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := adminTemplates.ExecuteTemplate(w, name, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_ = r
}

func (s *Site) basePage(r *http.Request, title string) pageData {
	return pageData{
		SiteTitle: s.config.Title,
		PageTitle: title,
		BasePath:  s.config.BasePath,
		Apps:      s.Apps(),
	}
}

func (s *Site) ensureCSRF(w http.ResponseWriter, r *http.Request) string {
	if s.config.DisableCSRF {
		return ""
	}
	if cookie, err := r.Cookie(csrfCookieName); err == nil && cookie.Value != "" {
		return cookie.Value
	}
	token := randomToken()
	http.SetCookie(w, &http.Cookie{
		Name:     csrfCookieName,
		Value:    token,
		Path:     s.config.BasePath,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	return token
}

func (s *Site) verifyCSRF(r *http.Request) bool {
	if s.config.DisableCSRF {
		return true
	}
	cookie, err := r.Cookie(csrfCookieName)
	if err != nil || cookie.Value == "" {
		return false
	}
	token := r.FormValue(csrfCookieName)
	return token != "" && token == cookie.Value
}

func randomToken() string {
	var buf [32]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(buf[:])
}

func formatValue(value any) string {
	switch typed := value.(type) {
	case nil:
		return "-"
	case bool:
		if typed {
			return "Yes"
		}
		return "No"
	case time.Time:
		if typed.IsZero() {
			return "-"
		}
		return typed.Format("2006-01-02 15:04")
	default:
		text := strings.TrimSpace(fmt.Sprint(value))
		if text == "" {
			return "-"
		}
		return text
	}
}
