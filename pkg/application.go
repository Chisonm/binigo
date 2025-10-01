package binigo

import (
	"fmt"
	"log"
	"net"
	"strconv"
	"sync"

	"github.com/valyala/fasthttp"
)

// Application is the main framework instance
type Application struct {
	router     *Router
	container  *Container
	middleware []MiddlewareFunc
	config     *Config
	mu         sync.RWMutex
}

// NewApplication creates a new framework instance
func NewApplication(config *Config) *Application {
	app := &Application{
		router:     NewRouter(),
		container:  NewContainer(),
		middleware: make([]MiddlewareFunc, 0),
		config:     config,
	}

	// Register core services
	app.registerCoreServices()

	return app
}

// registerCoreServices binds core framework services
func (a *Application) registerCoreServices() {
	a.container.Singleton("app", func(c *Container) interface{} {
		return a
	})

	a.container.Singleton("router", func(c *Container) interface{} {
		return a.router
	})

	a.container.Singleton("config", func(c *Container) interface{} {
		return a.config
	})
}

// Use adds global middleware
func (a *Application) Use(middleware ...MiddlewareFunc) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.middleware = append(a.middleware, middleware...)
}

// Router returns the router instance
func (a *Application) Router() *Router {
	return a.router
}

// Container returns the service container
func (a *Application) Container() *Container {
	return a.container
}

// Run starts the HTTP server
func (a *Application) Run(addr string) error {
	handler := a.buildHandler()

	// Find available port if the specified one is in use
	finalAddr := a.findAvailablePort(addr)

	log.Printf("Server starting on %s", finalAddr)
	return fasthttp.ListenAndServe(finalAddr, handler)
}

// findAvailablePort checks if the port is available, if not, finds the next available one
func (a *Application) findAvailablePort(addr string) string {
	host, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		// If no host specified, assume localhost
		portStr = addr
		if portStr[0] == ':' {
			portStr = portStr[1:]
		}
		host = ""
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		log.Printf("Invalid port: %s, using default 8000", portStr)
		port = 8000
	}

	// Try up to 100 ports
	for i := 0; i < 100; i++ {
		testPort := port + i
		testAddr := fmt.Sprintf(":%d", testPort)
		if host != "" {
			testAddr = fmt.Sprintf("%s:%d", host, testPort)
		}

		// Try to listen on the port
		ln, err := net.Listen("tcp", testAddr)
		if err == nil {
			ln.Close()
			if i > 0 {
				log.Printf("Port %d is in use, using port %d instead", port, testPort)
			}
			return testAddr
		}
	}

	// If all ports are taken, return the original
	log.Printf("Warning: Could not find available port, attempting original: %s", addr)
	return addr
}

// buildHandler creates the main request handler with middleware chain
func (a *Application) buildHandler() fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		// Create framework context
		fctx := NewContext(ctx, a)

		// Build middleware chain
		handler := a.router.Handle

		// Apply middleware in reverse order
		for i := len(a.middleware) - 1; i >= 0; i-- {
			handler = a.middleware[i](handler)
		}

		// Execute handler chain
		if err := handler(fctx); err != nil {
			// Error already handled by handler
			_ = err
		}
	}
}

// Group creates a route group with shared middleware/prefix
func (a *Application) Group(prefix string, fn func(r *Router)) *Router {
	return a.router.Group(prefix, fn)
}

// Route registration helpers
func (a *Application) Get(path string, handler HandlerFunc) *Route {
	return a.router.Get(path, handler)
}

func (a *Application) Post(path string, handler HandlerFunc) *Route {
	return a.router.Post(path, handler)
}

func (a *Application) Put(path string, handler HandlerFunc) *Route {
	return a.router.Put(path, handler)
}

func (a *Application) Delete(path string, handler HandlerFunc) *Route {
	return a.router.Delete(path, handler)
}

func (a *Application) Patch(path string, handler HandlerFunc) *Route {
	return a.router.Patch(path, handler)
}

func (a *Application) Options(path string, handler HandlerFunc) *Route {
	return a.router.Options(path, handler)
}

// Any registers a route for all HTTP methods
func (a *Application) Any(path string, handler HandlerFunc) []*Route {
	methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS", "HEAD"}
	routes := make([]*Route, len(methods))

	for i, method := range methods {
		routes[i] = a.router.Add(method, path, handler)
	}

	return routes
}

// Config structure
type Config struct {
	AppName     string
	Environment string
	Debug       bool
	Port        string
	Database    DatabaseConfig
}

type DatabaseConfig struct {
	Driver   string
	Host     string
	Port     string
	Database string
	Username string
	Password string
}

// LoadConfig loads configuration from environment
func LoadConfig() *Config {
	// In a real implementation, load from .env or environment variables
	return &Config{
		AppName:     "MyApp",
		Environment: "development",
		Debug:       true,
		Port:        ":8080",
		Database: DatabaseConfig{
			Driver:   "postgres",
			Host:     "localhost",
			Port:     "5432",
			Database: "myapp",
			Username: "user",
			Password: "password",
		},
	}
}

// Helper to create and configure a new application
func Bootstrap() *Application {
	config := LoadConfig()
	app := NewApplication(config)

	// Add default middleware
	app.Use(RecoveryMiddleware())
	app.Use(LoggerMiddleware())

	return app
}
