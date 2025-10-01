package binigo

import (
	"fmt"
	"regexp"
	"strings"
)

// Router manages application routes
type Router struct {
	routes     map[string][]*Route
	middleware []MiddlewareFunc
	prefix     string
	parent     *Router
}

// Route represents a single route definition
type Route struct {
	method     string
	path       string
	handler    HandlerFunc
	middleware []MiddlewareFunc
	name       string
	pattern    *regexp.Regexp
	paramNames []string
}

// NewRouter creates a new router instance
func NewRouter() *Router {
	return &Router{
		routes:     make(map[string][]*Route),
		middleware: make([]MiddlewareFunc, 0),
	}
}

// Add registers a new route
func (r *Router) Add(method, path string, handler HandlerFunc) *Route {
	route := &Route{
		method:     strings.ToUpper(method),
		path:       r.prefix + path,
		handler:    handler,
		middleware: make([]MiddlewareFunc, 0),
	}
	
	// Compile route pattern
	route.compile()
	
	// Store route
	if r.routes[route.method] == nil {
		r.routes[route.method] = make([]*Route, 0)
	}
	r.routes[route.method] = append(r.routes[route.method], route)
	
	return route
}

// compile creates regex pattern from route path
func (route *Route) compile() {
	// Convert Laravel-style {param} to regex
	pattern := route.path
	pattern = regexp.MustCompile(`\{(\w+)\}`).ReplaceAllStringFunc(pattern, func(match string) string {
		// Extract parameter name
		paramName := strings.Trim(match, "{}")
		route.paramNames = append(route.paramNames, paramName)
		return `([^/]+)`
	})
	
	// Convert Laravel-style {param?} optional parameters
	pattern = regexp.MustCompile(`\{(\w+)\?\}`).ReplaceAllStringFunc(pattern, func(match string) string {
		paramName := strings.Trim(match, "{?}")
		route.paramNames = append(route.paramNames, paramName)
		return `(?:([^/]+))?`
	})
	
	// Compile the pattern
	pattern = "^" + pattern + "$"
	route.pattern = regexp.MustCompile(pattern)
}

// Middleware adds middleware to this specific route
func (route *Route) Middleware(middleware ...MiddlewareFunc) *Route {
	route.middleware = append(route.middleware, middleware...)
	return route
}

// Name sets the route name
func (route *Route) Name(name string) *Route {
	route.name = name
	return route
}

// Match checks if the route matches the given path
func (route *Route) Match(path string) (bool, map[string]string) {
	matches := route.pattern.FindStringSubmatch(path)
	if matches == nil {
		return false, nil
	}
	
	params := make(map[string]string)
	for i, name := range route.paramNames {
		if i+1 < len(matches) && matches[i+1] != "" {
			params[name] = matches[i+1]
		}
	}
	
	return true, params
}

// Handle processes the incoming request
func (r *Router) Handle(ctx *Context) error {
	method := string(ctx.fastCtx.Method())
	path := string(ctx.fastCtx.Path())
	
	// Find matching route
	routes, ok := r.routes[method]
	if !ok {
		return ctx.Status(404).JSON(Map{
			"error": "Not Found",
		})
	}
	
	for _, route := range routes {
		if matched, params := route.Match(path); matched {
			// Set route parameters
			ctx.params = params
			ctx.route = route
			
			// Build handler with route middleware
			handler := route.handler
			
			// Apply route-specific middleware
			for i := len(route.middleware) - 1; i >= 0; i-- {
				handler = route.middleware[i](handler)
			}
			
			// Apply router group middleware
			for i := len(r.middleware) - 1; i >= 0; i-- {
				handler = r.middleware[i](handler)
			}
			
			return handler(ctx)
		}
	}
	
	return ctx.Status(404).JSON(Map{
		"error": "Not Found",
	})
}

// Group creates a route group
func (r *Router) Group(prefix string, fn func(router *Router)) *Router {
	group := &Router{
		routes:     r.routes,
		middleware: make([]MiddlewareFunc, 0),
		prefix:     r.prefix + prefix,
		parent:     r,
	}
	
	fn(group)
	return group
}

// Middleware adds middleware to the router group
func (r *Router) Middleware(middleware ...MiddlewareFunc) *Router {
	r.middleware = append(r.middleware, middleware...)
	return r
}

// Route registration methods
func (r *Router) Get(path string, handler HandlerFunc) *Route {
	return r.Add("GET", path, handler)
}

func (r *Router) Post(path string, handler HandlerFunc) *Route {
	return r.Add("POST", path, handler)
}

func (r *Router) Put(path string, handler HandlerFunc) *Route {
	return r.Add("PUT", path, handler)
}

func (r *Router) Delete(path string, handler HandlerFunc) *Route {
	return r.Add("DELETE", path, handler)
}

func (r *Router) Patch(path string, handler HandlerFunc) *Route {
	return r.Add("PATCH", path, handler)
}

func (r *Router) Options(path string, handler HandlerFunc) *Route {
	return r.Add("OPTIONS", path, handler)
}

// Resource creates RESTful routes for a resource
func (r *Router) Resource(name string, controller interface{}) {
	basePath := "/" + name
	
	// Index - GET /resource
	r.Get(basePath, func(ctx *Context) error {
		return fmt.Errorf("index not implemented")
	}).Name(name + ".index")
	
	// Create - GET /resource/create
	r.Get(basePath+"/create", func(ctx *Context) error {
		return fmt.Errorf("create not implemented")
	}).Name(name + ".create")
	
	// Store - POST /resource
	r.Post(basePath, func(ctx *Context) error {
		return fmt.Errorf("store not implemented")
	}).Name(name + ".store")
	
	// Show - GET /resource/{id}
	r.Get(basePath+"/{id}", func(ctx *Context) error {
		return fmt.Errorf("show not implemented")
	}).Name(name + ".show")
	
	// Edit - GET /resource/{id}/edit
	r.Get(basePath+"/{id}/edit", func(ctx *Context) error {
		return fmt.Errorf("edit not implemented")
	}).Name(name + ".edit")
	
	// Update - PUT/PATCH /resource/{id}
	r.Put(basePath+"/{id}", func(ctx *Context) error {
		return fmt.Errorf("update not implemented")
	}).Name(name + ".update")
	
	// Destroy - DELETE /resource/{id}
	r.Delete(basePath+"/{id}", func(ctx *Context) error {
		return fmt.Errorf("destroy not implemented")
	}).Name(name + ".destroy")
}