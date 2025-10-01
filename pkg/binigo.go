package binigo

// Version of the framework (set by build flags)
var Version = "dev"

// Map is a shorthand for map[string]interface{}
type Map map[string]interface{}

// HandlerFunc is the function signature for route handlers
type HandlerFunc func(*Context) error

// MiddlewareFunc wraps a handler with middleware functionality
type MiddlewareFunc func(HandlerFunc) HandlerFunc
