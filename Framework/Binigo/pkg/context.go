package binigo

import (
	"encoding/json"
	"fmt"
	"mime/multipart"

	"github.com/valyala/fasthttp"
)

// Context wraps fasthttp context with helper methods
type Context struct {
	fastCtx *fasthttp.RequestCtx
	app     *Application
	params  map[string]string
	route   *Route
	store   map[string]interface{} // For storing data during request lifecycle
}

// NewContext creates a new context instance
func NewContext(ctx *fasthttp.RequestCtx, app *Application) *Context {
	return &Context{
		fastCtx: ctx,
		app:     app,
		params:  make(map[string]string),
		store:   make(map[string]interface{}),
	}
}

// Request methods

// Param gets a route parameter
func (c *Context) Param(name string) string {
	return c.params[name]
}

// Query gets a query parameter
func (c *Context) Query(name string) string {
	return string(c.fastCtx.QueryArgs().Peek(name))
}

// QueryDefault gets a query parameter with default value
func (c *Context) QueryDefault(name, defaultValue string) string {
	value := c.Query(name)
	if value == "" {
		return defaultValue
	}
	return value
}

// Input gets input from request body (JSON)
func (c *Context) Input(name string) interface{} {
	var data map[string]interface{}
	if err := json.Unmarshal(c.fastCtx.PostBody(), &data); err != nil {
		return nil
	}
	return data[name]
}

// Bind binds request body to a struct
func (c *Context) Bind(v interface{}) error {
	contentType := string(c.fastCtx.Request.Header.ContentType())
	
	if contentType == "application/json" || len(contentType) == 0 {
		return json.Unmarshal(c.fastCtx.PostBody(), v)
	}
	
	return fmt.Errorf("unsupported content type: %s", contentType)
}

// FormValue gets a form value
func (c *Context) FormValue(name string) string {
	return string(c.fastCtx.FormValue(name))
}

// File gets uploaded file
func (c *Context) File(name string) (*multipart.FileHeader, error) {
	file, err := c.fastCtx.FormFile(name)
	if err != nil {
		return nil, err
	}
	return file, nil
}

// Header gets a request header
func (c *Context) Header(name string) string {
	return string(c.fastCtx.Request.Header.Peek(name))
}

// Method gets the HTTP method
func (c *Context) Method() string {
	return string(c.fastCtx.Method())
}

// Path gets the request path
func (c *Context) Path() string {
	return string(c.fastCtx.Path())
}

// IP gets client IP address
func (c *Context) IP() string {
	return c.fastCtx.RemoteIP().String()
}

// Response methods

// Status sets the HTTP status code
func (c *Context) Status(code int) *Context {
	c.fastCtx.SetStatusCode(code)
	return c
}

// JSON sends JSON response
func (c *Context) JSON(data interface{}) error {
	c.fastCtx.Response.Header.SetContentType("application/json")
	
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	
	c.fastCtx.SetBody(jsonData)
	return nil
}

// String sends plain text response
func (c *Context) String(format string, values ...interface{}) error {
	c.fastCtx.Response.Header.SetContentType("text/plain")
	c.fastCtx.SetBodyString(fmt.Sprintf(format, values...))
	return nil
}

// HTML sends HTML response
func (c *Context) HTML(html string) error {
	c.fastCtx.Response.Header.SetContentType("text/html")
	c.fastCtx.SetBodyString(html)
	return nil
}

// Redirect performs HTTP redirect
func (c *Context) Redirect(url string, statusCode ...int) error {
	code := 302
	if len(statusCode) > 0 {
		code = statusCode[0]
	}
	c.fastCtx.Redirect(url, code)
	return nil
}

// SetHeader sets a response header
func (c *Context) SetHeader(key, value string) *Context {
	c.fastCtx.Response.Header.Set(key, value)
	return c
}

// Cookie sets a cookie
func (c *Context) Cookie(name, value string, maxAge ...int) *Context {
	cookie := &fasthttp.Cookie{}
	cookie.SetKey(name)
	cookie.SetValue(value)
	
	if len(maxAge) > 0 {
		cookie.SetMaxAge(maxAge[0])
	}
	
	c.fastCtx.Response.Header.SetCookie(cookie)
	return c
}

// GetCookie gets a cookie value
func (c *Context) GetCookie(name string) string {
	return string(c.fastCtx.Request.Header.Cookie(name))
}

// Store methods (for passing data between middleware)

// Set stores a value in the context
func (c *Context) Set(key string, value interface{}) {
	c.store[key] = value
}

// Get retrieves a value from the context
func (c *Context) Get(key string) interface{} {
	return c.store[key]
}

// GetString retrieves a string value from the context
func (c *Context) GetString(key string) string {
	if val, ok := c.store[key].(string); ok {
		return val
	}
	return ""
}

// MustGet retrieves a value or panics if not found
func (c *Context) MustGet(key string) interface{} {
	if val, exists := c.store[key]; exists {
		return val
	}
	panic(fmt.Sprintf("key '%s' does not exist", key))
}

// App returns the application instance
func (c *Context) App() *Application {
	return c.app
}

// Container returns the service container
func (c *Context) Container() *Container {
	return c.app.Container()
}

// FastHTTP returns the underlying fasthttp context
func (c *Context) FastHTTP() *fasthttp.RequestCtx {
	return c.fastCtx
}

// Abort stops the handler chain execution
func (c *Context) Abort() {
	c.fastCtx.SetStatusCode(fasthttp.StatusNotImplemented)
}

// AbortWithStatus stops execution and sets status code
func (c *Context) AbortWithStatus(code int) {
	c.fastCtx.SetStatusCode(code)
}

// AbortWithJSON stops execution and returns JSON
func (c *Context) AbortWithJSON(code int, data interface{}) error {
	c.Status(code)
	return c.JSON(data)
}

// Validation helpers

// Validate validates the request using a validator
func (c *Context) Validate(rules map[string]string) error {
	// Implementation would integrate with a validation library
	return nil
}

// Helper response methods

// Success returns a success JSON response
func (c *Context) Success(data interface{}, message ...string) error {
	response := Map{
		"success": true,
		"data":    data,
	}
	
	if len(message) > 0 {
		response["message"] = message[0]
	}
	
	return c.JSON(response)
}

// Error returns an error JSON response
func (c *Context) Error(message string, code ...int) error {
	statusCode := 400
	if len(code) > 0 {
		statusCode = code[0]
	}
	
	c.Status(statusCode)
	return c.JSON(Map{
		"success": false,
		"error":   message,
	})
}