package admin

import (
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"
)

// SiteConfig controls a mounted admin site.
type SiteConfig struct {
	Title    string
	BasePath string
}

// Site owns the admin registry and exposes the http.Handler.
type Site struct {
	config SiteConfig
	mu     sync.RWMutex
	apps   map[string]*App
	order  []string
}

// App groups related resources.
type App struct {
	site      *Site
	name      string
	label     string
	resources map[string]resourceRuntime
	order     []string
}

// AppMeta is a stable description of a registered app.
type AppMeta struct {
	Name      string
	Label     string
	Resources []ResourceMeta
}

// ResourceMeta is a stable description of a registered resource.
type ResourceMeta struct {
	Name  string
	Label string
}

// New creates an admin site.
func New(config SiteConfig) *Site {
	config.BasePath = normalizeBasePath(config.BasePath)
	if config.Title == "" {
		config.Title = "Admin"
	}
	return &Site{
		config: config,
		apps:   map[string]*App{},
	}
}

// BasePath returns the normalized mount path.
func (s *Site) BasePath() string {
	return s.config.BasePath
}

// App returns an existing app or registers a new app.
func (s *Site) App(name, label string) *App {
	s.mu.Lock()
	defer s.mu.Unlock()

	if app, ok := s.apps[name]; ok {
		if label != "" {
			app.label = label
		}
		return app
	}

	app := &App{
		site:      s,
		name:      name,
		label:     displayLabel(name, label),
		resources: map[string]resourceRuntime{},
	}
	s.apps[name] = app
	s.order = append(s.order, name)
	return app
}

// Apps returns registered app metadata in registration order.
func (s *Site) Apps() []AppMeta {
	s.mu.RLock()
	defer s.mu.RUnlock()

	apps := make([]AppMeta, 0, len(s.order))
	for _, name := range s.order {
		app := s.apps[name]
		meta := AppMeta{Name: app.name, Label: app.label}
		for _, resourceName := range app.order {
			resource := app.resources[resourceName].meta()
			meta.Resources = append(meta.Resources, ResourceMeta{
				Name:  resource.Name,
				Label: resource.Label,
			})
		}
		apps = append(apps, meta)
	}
	return apps
}

// Handler returns the admin HTTP handler.
func (s *Site) Handler() http.Handler {
	return s
}

// ServeHTTP currently exposes a minimal placeholder until UI routes are added.
func (s *Site) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = w.Write([]byte(s.config.Title))
}

// Register adds a resource to an app.
func Register[T any, ID comparable](a *App, resource Resource[T, ID]) error {
	if !validSlug(a.name) {
		return fmt.Errorf("invalid app name %q", a.name)
	}
	if err := validateResource(resource); err != nil {
		return err
	}

	a.site.mu.Lock()
	defer a.site.mu.Unlock()

	if _, exists := a.resources[resource.Name]; exists {
		return fmt.Errorf("resource %q already registered in app %q", resource.Name, a.name)
	}
	a.resources[resource.Name] = &typedResource[T, ID]{app: a, resource: resource}
	a.order = append(a.order, resource.Name)
	return nil
}

type typedResource[T any, ID comparable] struct {
	app      *App
	resource Resource[T, ID]
}

func (r *typedResource[T, ID]) meta() resourceMeta {
	return resourceMeta{
		Name:  r.resource.Name,
		Label: displayLabel(r.resource.Name, r.resource.Label),
	}
}

func normalizeBasePath(path string) string {
	path = strings.TrimSpace(path)
	if path == "" || path == "/" {
		return "/admin"
	}
	path = "/" + strings.Trim(path, "/")
	if path == "" {
		return "/admin"
	}
	return path
}

func sortResourceMeta(resources []ResourceMeta) {
	sort.SliceStable(resources, func(i, j int) bool {
		return resources[i].Name < resources[j].Name
	})
}
