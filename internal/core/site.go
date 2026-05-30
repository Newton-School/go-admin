package core

import (
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"strings"
	"sync"
)

// SiteConfig controls a mounted admin site.
type SiteConfig struct {
	Title       string
	BasePath    string
	DisableCSRF bool
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
	Name      string         `json:"name"`
	Label     string         `json:"label"`
	Resources []ResourceMeta `json:"resources"`
}

// ResourceMeta is a stable description of a registered resource.
type ResourceMeta struct {
	Name  string `json:"name"`
	Label string `json:"label"`
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

// ServeHTTP routes admin UI requests.
func (s *Site) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == s.config.BasePath {
		http.Redirect(w, r, s.config.BasePath+"/", http.StatusMovedPermanently)
		return
	}
	prefix := s.config.BasePath + "/"
	if !strings.HasPrefix(r.URL.Path, prefix) {
		http.NotFound(w, r)
		return
	}

	relative := strings.Trim(strings.TrimPrefix(r.URL.Path, prefix), "/")
	if relative == "" {
		s.handleIndex(w, r)
		return
	}

	parts := strings.Split(relative, "/")
	if parts[0] == "static" {
		s.handleStatic(w, r)
		return
	}
	if parts[0] == "api" {
		s.handleAPI(w, r, parts[1:])
		return
	}
	if len(parts) == 1 {
		s.handleApp(w, r, parts[0])
		return
	}

	app, resource, ok := s.lookup(parts[0], parts[1])
	if !ok {
		http.NotFound(w, r)
		return
	}
	switch {
	case len(parts) == 2:
		s.handleList(w, r, app, resource)
	case len(parts) == 3 && parts[2] == "new":
		s.handleCreate(w, r, app, resource)
	case len(parts) == 3:
		s.handleDetail(w, r, app, resource, parts[2])
	case len(parts) == 4 && parts[2] == "actions":
		s.handleHTMLAction(w, r, app, resource, parts[3])
	case len(parts) == 4 && parts[3] == "delete":
		s.handleDelete(w, r, app, resource, parts[2])
	default:
		http.NotFound(w, r)
	}
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

func (s *Site) handleStatic(w http.ResponseWriter, r *http.Request) {
	staticFS, err := fs.Sub(embeddedFiles, "assets/static")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.StripPrefix(s.config.BasePath+"/static/", http.FileServer(http.FS(staticFS))).ServeHTTP(w, r)
}

func (s *Site) handleIndex(w http.ResponseWriter, r *http.Request) {
	data := s.basePage(r, s.config.Title)
	s.render(w, r, "index", data)
}

func (s *Site) handleApp(w http.ResponseWriter, r *http.Request, appName string) {
	apps := s.Apps()
	for _, app := range apps {
		if app.Name == appName {
			data := s.basePage(r, app.Label)
			data.Apps = []AppMeta{app}
			s.render(w, r, "index", data)
			return
		}
	}
	http.NotFound(w, r)
}

func (s *Site) handleList(w http.ResponseWriter, r *http.Request, app *App, resource resourceRuntime) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	listConfig := resource.listConfig()
	query := QueryFromRequest(r, QueryConfig{
		DefaultPerPage: defaultPerPage,
		MaxPerPage:     defaultMaxPage,
		AllowedSorts:   allowedSorts(listConfig),
		FilterNames:    filterNames(listConfig.Filters),
	})
	if len(query.Sort) == 0 {
		query.Sort = append(query.Sort, listConfig.Sort...)
	}
	page, err := resource.list(r.Context(), query)
	if err != nil {
		s.writeError(w, err)
		return
	}

	meta := resource.meta()
	columns := fieldsByNames(resource.fields(), listConfig.Columns)
	rows := make([]listRow, 0, len(page.Items))
	for _, item := range page.Items {
		id := resource.idString(item)
		row := listRow{ID: id, Detail: s.resourceURL(app.name, meta.Name, id)}
		for _, column := range columns {
			row.Cells = append(row.Cells, cellView{Field: column, Value: resource.fieldValue(item, column.name())})
		}
		rows = append(rows, row)
	}

	base := s.basePage(r, meta.Label)
	actions := resource.actions()
	if len(actions) > 0 {
		base.CSRFToken = s.ensureCSRF(w, r)
	}
	data := listPageData{
		pageData: base,
		App:      AppMeta{Name: app.name, Label: app.label},
		Resource: ResourceMeta{
			Name:  meta.Name,
			Label: meta.Label,
		},
		Columns: columns,
		Rows:    rows,
		Search:  query.Search,
		Filters: filterViews(listConfig.Filters, query.Filters),
		Page: pageView{
			Total:   page.Total,
			Page:    page.Page,
			PerPage: page.PerPage,
		},
		NewURL:  s.resourceURL(app.name, meta.Name, "new"),
		Actions: actions,
	}
	if len(actions) > 0 {
		data.ActionURL = s.resourceURL(app.name, meta.Name, "actions/"+actions[0].Name)
	}
	s.render(w, r, "list", data)
}

func (s *Site) handleCreate(w http.ResponseWriter, r *http.Request, app *App, resource resourceRuntime) {
	switch r.Method {
	case http.MethodGet:
		s.renderResourceForm(w, r, app, resource, nil, nil, true)
	case http.MethodPost:
		if !s.verifyCSRF(r) {
			http.Error(w, "csrf token invalid", http.StatusForbidden)
			return
		}
		if err := r.ParseForm(); err != nil {
			http.Error(w, "invalid form", http.StatusBadRequest)
			return
		}
		obj, errs, err := resource.create(r.Context(), r.PostForm)
		if err != nil {
			s.writeError(w, err)
			return
		}
		if !errs.Empty() {
			w.WriteHeader(http.StatusUnprocessableEntity)
			s.renderResourceForm(w, r, app, resource, obj, errs, true)
			return
		}
		http.Redirect(w, r, s.resourceURL(app.name, resource.meta().Name, resource.idString(obj)), http.StatusSeeOther)
	default:
		w.Header().Set("Allow", strings.Join([]string{http.MethodGet, http.MethodPost}, ", "))
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Site) handleDetail(w http.ResponseWriter, r *http.Request, app *App, resource resourceRuntime, rawID string) {
	switch r.Method {
	case http.MethodGet:
		obj, err := resource.get(r.Context(), rawID)
		if err != nil {
			s.writeError(w, err)
			return
		}
		s.renderResourceForm(w, r, app, resource, obj, nil, false)
	case http.MethodPost:
		if !s.verifyCSRF(r) {
			http.Error(w, "csrf token invalid", http.StatusForbidden)
			return
		}
		if err := r.ParseForm(); err != nil {
			http.Error(w, "invalid form", http.StatusBadRequest)
			return
		}
		obj, errs, err := resource.update(r.Context(), rawID, r.PostForm)
		if err != nil {
			s.writeError(w, err)
			return
		}
		if !errs.Empty() {
			w.WriteHeader(http.StatusUnprocessableEntity)
			s.renderResourceForm(w, r, app, resource, obj, errs, false)
			return
		}
		http.Redirect(w, r, s.resourceURL(app.name, resource.meta().Name, rawID), http.StatusSeeOther)
	default:
		w.Header().Set("Allow", strings.Join([]string{http.MethodGet, http.MethodPost}, ", "))
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Site) handleDelete(w http.ResponseWriter, r *http.Request, app *App, resource resourceRuntime, rawID string) {
	switch r.Method {
	case http.MethodGet:
		if _, err := resource.get(r.Context(), rawID); err != nil {
			s.writeError(w, err)
			return
		}
		meta := resource.meta()
		base := s.basePage(r, "Delete "+meta.Label)
		base.CSRFToken = s.ensureCSRF(w, r)
		data := deletePageData{
			pageData:  base,
			App:       AppMeta{Name: app.name, Label: app.label},
			Resource:  ResourceMeta{Name: meta.Name, Label: meta.Label},
			ActionURL: s.resourceURL(app.name, meta.Name, rawID) + "/delete",
			BackURL:   s.resourceURL(app.name, meta.Name, rawID),
			ObjectID:  rawID,
		}
		s.render(w, r, "delete", data)
	case http.MethodPost:
		if !s.verifyCSRF(r) {
			http.Error(w, "csrf token invalid", http.StatusForbidden)
			return
		}
		if err := resource.delete(r.Context(), rawID); err != nil {
			s.writeError(w, err)
			return
		}
		http.Redirect(w, r, s.resourceListURL(app.name, resource.meta().Name), http.StatusSeeOther)
	default:
		w.Header().Set("Allow", strings.Join([]string{http.MethodGet, http.MethodPost}, ", "))
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Site) handleHTMLAction(w http.ResponseWriter, r *http.Request, app *App, resource resourceRuntime, actionName string) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !s.verifyCSRF(r) {
		http.Error(w, "csrf token invalid", http.StatusForbidden)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}
	ids := r.PostForm["ids"]
	if len(ids) == 0 {
		http.Error(w, "select at least one row", http.StatusBadRequest)
		return
	}
	if _, err := resource.runAction(r.Context(), actionName, ids); err != nil {
		s.writeError(w, err)
		return
	}
	http.Redirect(w, r, s.resourceListURL(app.name, resource.meta().Name), http.StatusSeeOther)
}

func (s *Site) renderResourceForm(w http.ResponseWriter, r *http.Request, app *App, resource resourceRuntime, obj any, errs ValidationErrors, isNew bool) {
	meta := resource.meta()
	title := "Edit " + meta.Label
	actionURL := s.resourceURL(app.name, meta.Name, resource.idString(obj))
	deleteURL := actionURL + "/delete"
	if isNew {
		title = "New " + meta.Label
		actionURL = s.resourceURL(app.name, meta.Name, "new")
		deleteURL = ""
	}
	base := s.basePage(r, title)
	base.CSRFToken = s.ensureCSRF(w, r)
	data := formPageData{
		pageData:  base,
		App:       AppMeta{Name: app.name, Label: app.label},
		Resource:  ResourceMeta{Name: meta.Name, Label: meta.Label},
		ActionURL: actionURL,
		BackURL:   s.resourceListURL(app.name, meta.Name),
		DeleteURL: deleteURL,
		IsNew:     isNew,
		Errors:    errs,
		Fieldsets: buildFieldsets(resource, obj, errs),
	}
	s.render(w, r, "form", data)
}

func (s *Site) lookup(appName, resourceName string) (*App, resourceRuntime, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	app := s.apps[appName]
	if app == nil {
		return nil, nil, false
	}
	resource := app.resources[resourceName]
	if resource == nil {
		return nil, nil, false
	}
	return app, resource, true
}

func (s *Site) writeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrNotFound):
		http.Error(w, "not found", http.StatusNotFound)
	default:
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Site) resourceListURL(appName, resourceName string) string {
	return s.config.BasePath + "/" + appName + "/" + resourceName + "/"
}

func (s *Site) resourceURL(appName, resourceName, suffix string) string {
	url := s.config.BasePath + "/" + appName + "/" + resourceName
	if suffix != "" {
		url += "/" + suffix
	}
	return url
}

func allowedSorts(config ListConfig) []string {
	values := make([]string, 0, len(config.Sort)+len(config.Columns))
	for _, sortField := range config.Sort {
		values = append(values, sortField.Field)
	}
	values = append(values, config.Columns...)
	return values
}

func filterNames(filters []Filter) []string {
	values := make([]string, 0, len(filters))
	for _, filter := range filters {
		values = append(values, filter.Name)
	}
	return values
}

func filterViews(filters []Filter, selected map[string][]string) []filterView {
	views := make([]filterView, 0, len(filters))
	for _, filter := range filters {
		choices := filter.Choices
		if len(choices) == 0 {
			choices = []Choice{{Value: "true", Label: "Yes"}, {Value: "false", Label: "No"}}
		}
		views = append(views, filterView{
			Name:    filter.Name,
			Label:   displayLabel(filter.Name, filter.Label),
			Choices: choices,
			Values:  selected[filter.Name],
		})
	}
	return views
}

func fieldsByNames(fields []Field, names []string) []Field {
	if len(names) == 0 {
		return fields
	}
	byName := map[string]Field{}
	for _, field := range fields {
		byName[field.name()] = field
	}
	selected := make([]Field, 0, len(names))
	for _, name := range names {
		if field, ok := byName[name]; ok {
			selected = append(selected, field)
		}
	}
	return selected
}

func buildFieldsets(resource resourceRuntime, obj any, errs ValidationErrors) []fieldsetView {
	fields := resource.fields()
	fieldsets := resource.fieldsets()
	if len(fieldsets) == 0 {
		names := make([]string, 0, len(fields))
		for _, field := range fields {
			names = append(names, field.name())
		}
		fieldsets = []Fieldset{{Fields: names}}
	}
	byName := map[string]Field{}
	for _, field := range fields {
		byName[field.name()] = field
	}
	views := make([]fieldsetView, 0, len(fieldsets))
	for _, fieldset := range fieldsets {
		view := fieldsetView{
			Title:       fieldset.Title,
			Description: fieldset.Description,
			Collapsed:   fieldset.Collapsed,
		}
		names := fieldset.Fields
		if len(names) == 0 {
			for _, row := range fieldset.Rows {
				names = append(names, row...)
			}
		}
		for _, name := range names {
			field, ok := byName[name]
			if !ok {
				continue
			}
			value := resource.fieldValue(obj, name)
			view.Fields = append(view.Fields, fieldView{
				Field:    field,
				Value:    value,
				Widget:   RenderWidget(WidgetContext{Field: field, Value: value, Errors: errs}),
				Readonly: field.ReadonlyValue,
				Display:  formatValue(value),
			})
		}
		views = append(views, view)
	}
	return views
}
