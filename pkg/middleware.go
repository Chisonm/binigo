package binigo

import (
	"fmt"
	"log"
	"runtime/debug"
	"strings"
	"time"
)

// LoggerMiddleware logs incoming requests
func LoggerMiddleware() MiddlewareFunc {
	return func(next HandlerFunc) HandlerFunc {
		return func(ctx *Context) error {
			start := time.Now()

			method := ctx.Method()
			path := ctx.Path()

			// Process request
			err := next(ctx)

			// Log request details
			duration := time.Since(start)
			status := ctx.fastCtx.Response.StatusCode()

			log.Printf("[%s] %s %d - %v", method, path, status, duration)

			return err
		}
	}
}

// RecoveryMiddleware recovers from panics
func RecoveryMiddleware() MiddlewareFunc {
	return func(next HandlerFunc) HandlerFunc {
		return func(ctx *Context) error {
			defer func() {
				if err := recover(); err != nil {
					log.Printf("PANIC: %v\n%s", err, debug.Stack())

					_ = ctx.Status(500).JSON(Map{
						"error": "Internal Server Error",
					})
				}
			}()

			return next(ctx)
		}
	}
}

// CORSMiddleware handles CORS headers
func CORSMiddleware(allowedOrigins ...string) MiddlewareFunc {
	origins := "*"
	if len(allowedOrigins) > 0 {
		origins = strings.Join(allowedOrigins, ",")
	}

	return func(next HandlerFunc) HandlerFunc {
		return func(ctx *Context) error {
			ctx.SetHeader("Access-Control-Allow-Origin", origins)
			ctx.SetHeader("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
			ctx.SetHeader("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
			ctx.SetHeader("Access-Control-Allow-Credentials", "true")

			// Handle preflight
			if ctx.Method() == "OPTIONS" {
				ctx.Status(204)
				return nil
			}

			return next(ctx)
		}
	}
}

// AuthMiddleware checks for authentication
func AuthMiddleware() MiddlewareFunc {
	return func(next HandlerFunc) HandlerFunc {
		return func(ctx *Context) error {
			token := ctx.Header("Authorization")

			if token == "" {
				token = ctx.GetCookie("auth_token")
			}

			if token == "" {
				return ctx.AbortWithJSON(401, Map{
					"error": "Unauthorized",
				})
			}

			// Remove "Bearer " prefix if present
			token = strings.TrimPrefix(token, "Bearer ")

			// Validate token (implement your logic here)
			user, err := validateToken(token)
			if err != nil {
				return ctx.AbortWithJSON(401, Map{
					"error": "Invalid token",
				})
			}

			// Store user in context
			ctx.Set("user", user)

			return next(ctx)
		}
	}
}

// validateToken is a placeholder for token validation
func validateToken(token string) (interface{}, error) {
	// Implement JWT validation or session validation here
	if token == "" {
		return nil, fmt.Errorf("invalid token")
	}

	// Return user object
	return Map{
		"id":    1,
		"email": "user@example.com",
		"name":  "John Doe",
	}, nil
}

// RateLimitMiddleware implements rate limiting
func RateLimitMiddleware(requestsPerMinute int) MiddlewareFunc {
	type client struct {
		requests  int
		resetTime time.Time
	}

	clients := make(map[string]*client)

	return func(next HandlerFunc) HandlerFunc {
		return func(ctx *Context) error {
			ip := ctx.IP()

			now := time.Now()

			if c, exists := clients[ip]; exists {
				if now.After(c.resetTime) {
					c.requests = 0
					c.resetTime = now.Add(time.Minute)
				}

				if c.requests >= requestsPerMinute {
					return ctx.AbortWithJSON(429, Map{
						"error": "Too many requests",
					})
				}

				c.requests++
			} else {
				clients[ip] = &client{
					requests:  1,
					resetTime: now.Add(time.Minute),
				}
			}

			return next(ctx)
		}
	}
}

// JSONOnlyMiddleware ensures requests are JSON
func JSONOnlyMiddleware() MiddlewareFunc {
	return func(next HandlerFunc) HandlerFunc {
		return func(ctx *Context) error {
			contentType := ctx.Header("Content-Type")

			if ctx.Method() != "GET" && ctx.Method() != "DELETE" {
				if !strings.Contains(contentType, "application/json") {
					return ctx.AbortWithJSON(415, Map{
						"error": "Content-Type must be application/json",
					})
				}
			}

			return next(ctx)
		}
	}
}

// TrimTrailingSlashMiddleware removes trailing slashes from URLs
func TrimTrailingSlashMiddleware() MiddlewareFunc {
	return func(next HandlerFunc) HandlerFunc {
		return func(ctx *Context) error {
			path := ctx.Path()

			if len(path) > 1 && path[len(path)-1] == '/' {
				newPath := path[:len(path)-1]
				return ctx.Redirect(newPath, 301)
			}

			return next(ctx)
		}
	}
}

// GuestMiddleware ensures user is NOT authenticated
func GuestMiddleware() MiddlewareFunc {
	return func(next HandlerFunc) HandlerFunc {
		return func(ctx *Context) error {
			token := ctx.Header("Authorization")

			if token == "" {
				token = ctx.GetCookie("auth_token")
			}

			if token != "" {
				return ctx.AbortWithJSON(403, Map{
					"error": "Already authenticated",
				})
			}

			return next(ctx)
		}
	}
}

// CompressionMiddleware enables gzip compression
func CompressionMiddleware() MiddlewareFunc {
	return func(next HandlerFunc) HandlerFunc {
		return func(ctx *Context) error {
			acceptEncoding := ctx.Header("Accept-Encoding")

			if strings.Contains(acceptEncoding, "gzip") {
				ctx.fastCtx.Response.Header.Set("Content-Encoding", "gzip")
			}

			return next(ctx)
		}
	}
}

// SecureHeadersMiddleware adds security headers
func SecureHeadersMiddleware() MiddlewareFunc {
	return func(next HandlerFunc) HandlerFunc {
		return func(ctx *Context) error {
			ctx.SetHeader("X-Content-Type-Options", "nosniff")
			ctx.SetHeader("X-Frame-Options", "DENY")
			ctx.SetHeader("X-XSS-Protection", "1; mode=block")
			ctx.SetHeader("Strict-Transport-Security", "max-age=31536000; includeSubDomains")

			return next(ctx)
		}
	}
}

// RequestIDMiddleware adds a unique request ID
func RequestIDMiddleware() MiddlewareFunc {
	return func(next HandlerFunc) HandlerFunc {
		return func(ctx *Context) error {
			requestID := ctx.Header("X-Request-ID")

			if requestID == "" {
				requestID = generateRequestID()
			}

			ctx.Set("request_id", requestID)
			ctx.SetHeader("X-Request-ID", requestID)

			return next(ctx)
		}
	}
}

// generateRequestID creates a unique request ID
func generateRequestID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
