package binigo

import (
	"fmt"
	"reflect"
	"sync"
)

// Container is the dependency injection container
type Container struct {
	bindings  map[string]*binding
	instances map[string]interface{}
	mu        sync.RWMutex
	aliases   map[string]string
}

// binding represents a service binding
type binding struct {
	resolver  func(*Container) interface{}
	singleton bool
}

// NewContainer creates a new service container
func NewContainer() *Container {
	return &Container{
		bindings:  make(map[string]*binding),
		instances: make(map[string]interface{}),
		aliases:   make(map[string]string),
	}
}

// Bind registers a binding in the container
func (c *Container) Bind(abstract string, resolver func(*Container) interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.bindings[abstract] = &binding{
		resolver:  resolver,
		singleton: false,
	}
}

// Singleton registers a singleton binding
func (c *Container) Singleton(abstract string, resolver func(*Container) interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.bindings[abstract] = &binding{
		resolver:  resolver,
		singleton: true,
	}
}

// Instance registers an existing instance as shared
func (c *Container) Instance(abstract string, instance interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.instances[abstract] = instance
}

// Make resolves a binding from the container
func (c *Container) Make(abstract string) (interface{}, error) {
	c.mu.RLock()

	// Check for alias
	if alias, ok := c.aliases[abstract]; ok {
		abstract = alias
	}

	// Check if instance already exists
	if instance, ok := c.instances[abstract]; ok {
		c.mu.RUnlock()
		return instance, nil
	}

	// Get binding
	bind, ok := c.bindings[abstract]
	c.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("binding not found: %s", abstract)
	}

	// Resolve the binding
	instance := bind.resolver(c)

	// Store singleton
	if bind.singleton {
		c.mu.Lock()
		c.instances[abstract] = instance
		c.mu.Unlock()
	}

	return instance, nil
}

// MustMake resolves a binding or panics
func (c *Container) MustMake(abstract string) interface{} {
	instance, err := c.Make(abstract)
	if err != nil {
		panic(err)
	}
	return instance
}

// Call invokes a function with dependency injection
func (c *Container) Call(fn interface{}, params ...interface{}) ([]interface{}, error) {
	fnValue := reflect.ValueOf(fn)
	fnType := fnValue.Type()

	if fnType.Kind() != reflect.Func {
		return nil, fmt.Errorf("not a function")
	}

	// Build arguments
	args := make([]reflect.Value, fnType.NumIn())
	paramIndex := 0

	for i := 0; i < fnType.NumIn(); i++ {
		argType := fnType.In(i)

		// Try to use provided parameters first
		if paramIndex < len(params) {
			args[i] = reflect.ValueOf(params[paramIndex])
			paramIndex++
			continue
		}

		// Try to resolve from container
		typeName := argType.String()
		instance, err := c.Make(typeName)
		if err == nil {
			args[i] = reflect.ValueOf(instance)
			continue
		}

		// Create zero value
		args[i] = reflect.Zero(argType)
	}

	// Call function
	results := fnValue.Call(args)

	// Convert results to []interface{}
	output := make([]interface{}, len(results))
	for i, result := range results {
		output[i] = result.Interface()
	}

	return output, nil
}

// Resolve builds an instance of the given type with dependency injection
func (c *Container) Resolve(target interface{}) error {
	targetValue := reflect.ValueOf(target)

	if targetValue.Kind() != reflect.Ptr {
		return fmt.Errorf("target must be a pointer")
	}

	targetType := targetValue.Elem().Type()

	if targetType.Kind() != reflect.Struct {
		return fmt.Errorf("target must be a pointer to a struct")
	}

	// Iterate through struct fields
	for i := 0; i < targetType.NumField(); i++ {
		field := targetType.Field(i)

		// Check for inject tag
		tag := field.Tag.Get("inject")
		if tag == "" {
			continue
		}

		// Resolve dependency
		instance, err := c.Make(tag)
		if err != nil {
			return fmt.Errorf("failed to inject %s: %v", tag, err)
		}

		// Set field value
		fieldValue := targetValue.Elem().Field(i)
		if fieldValue.CanSet() {
			fieldValue.Set(reflect.ValueOf(instance))
		}
	}

	return nil
}

// Alias creates an alias for a binding
func (c *Container) Alias(alias, abstract string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.aliases[alias] = abstract
}

// Bound checks if a binding exists
func (c *Container) Bound(abstract string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if _, ok := c.bindings[abstract]; ok {
		return true
	}

	if _, ok := c.instances[abstract]; ok {
		return true
	}

	return false
}

// Forget removes a binding from the container
func (c *Container) Forget(abstract string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.bindings, abstract)
	delete(c.instances, abstract)
}

// Flush clears all bindings and instances
func (c *Container) Flush() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.bindings = make(map[string]*binding)
	c.instances = make(map[string]interface{})
	c.aliases = make(map[string]string)
}

// ServiceProvider interface for registering services
type ServiceProvider interface {
	Register(*Container)
	Boot(*Application) error
}

// RegisterProvider registers a service provider
func (c *Container) RegisterProvider(provider ServiceProvider, app *Application) error {
	// Register bindings
	provider.Register(c)

	// Boot provider
	return provider.Boot(app)
}

// Example usage with struct tags:
// type UserController struct {
//     DB *Database `inject:"db"`
//     Cache *Cache `inject:"cache"`
// }
//
// controller := &UserController{}
// container.Resolve(controller)
