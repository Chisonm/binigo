package binigo

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
	"time"
)

// ANSI color codes
const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorCyan   = "\033[36m"
	ColorWhite  = "\033[37m"
	ColorGray   = "\033[90m"
)

// LoggerMiddleware logs incoming requests with colors and file logging
func LoggerMiddleware() MiddlewareFunc {
	// Setup file logger
	logDir := "storage/logs"
	os.MkdirAll(logDir, 0755)

	logFile, err := os.OpenFile(
		filepath.Join(logDir, "app.log"),
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		0644,
	)
	if err != nil {
		log.Printf("Warning: Could not open log file: %v", err)
	}

	// Create multi-writer for both console and file
	var writers []io.Writer
	writers = append(writers, os.Stdout)
	if logFile != nil {
		writers = append(writers, logFile)
	}
	multiWriter := io.MultiWriter(writers...)

	return func(next HandlerFunc) HandlerFunc {
		return func(ctx *Context) error {
			start := time.Now()

			method := ctx.Method()
			path := ctx.Path()
			ip := ctx.IP()

			// Process request
			err := next(ctx)

			// Log request details
			duration := time.Since(start)
			status := ctx.fastCtx.Response.StatusCode()

			// Color code based on status
			statusColor := getStatusColor(status)
			methodColor := getMethodColor(method)

			// Format timestamp
			timestamp := time.Now().Format("2006/01/02 15:04:05")

			// Console log (with colors)
			consoleLog := fmt.Sprintf(
				"%s[%s]%s %s%s%s %s%-7s%s %s%3d%s %s%-6v%s %s%s%s",
				ColorGray, timestamp, ColorReset,
				methodColor, method, ColorReset,
				ColorCyan, ip, ColorReset,
				statusColor, status, ColorReset,
				ColorBlue, duration.Round(time.Millisecond), ColorReset,
				ColorWhite, path, ColorReset,
			)

			// File log (without colors)
			fileLog := fmt.Sprintf(
				"[%s] %s %-7s %3d %-6v %s",
				timestamp, method, ip, status, duration.Round(time.Millisecond), path,
			)

			// Log error if present
			if err != nil {
				consoleLog += fmt.Sprintf(" %sERROR: %v%s", ColorRed, err, ColorReset)
				fileLog += fmt.Sprintf(" ERROR: %v", err)
			}

			// Write to console
			fmt.Fprintln(os.Stdout, consoleLog)

			// Write to file (without colors)
			if logFile != nil {
				fmt.Fprintln(logFile, fileLog)
			}

			return err
		}
	}
}

// getStatusColor returns color based on HTTP status code
func getStatusColor(status int) string {
	switch {
	case status >= 200 && status < 300:
		return ColorGreen
	case status >= 300 && status < 400:
		return ColorCyan
	case status >= 400 && status < 500:
		return ColorYellow
	case status >= 500:
		return ColorRed
	default:
		return ColorWhite
	}
}

// getMethodColor returns color based on HTTP method
func getMethodColor(method string) string {
	switch method {
	case "GET":
		return ColorBlue
	case "POST":
		return ColorGreen
	case "PUT":
		return ColorYellow
	case "DELETE":
		return ColorRed
	case "PATCH":
		return ColorCyan
	default:
		return ColorWhite
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
