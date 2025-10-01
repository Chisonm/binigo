package binigo

// Version of the framework
const Version = "1.0.0"

// Map is a shorthand for map[string]interface{}
type Map map[string]interface{}

// HandlerFunc is the function signature for route handlers
type HandlerFunc func(*Context) error

// MiddlewareFunc wraps a handler with middleware functionality
type MiddlewareFunc func(HandlerFunc) HandlerFunc
