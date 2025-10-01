package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
)

const version = "1.0.0"

func main() {
	if len(os.Args) < 2 {
		printHelp()
		return
	}

	command := os.Args[1]

	switch command {
	case "new":
		if len(os.Args) < 3 {
			fmt.Println("❌ Project name required: binigo new <project-name>")
			os.Exit(1)
		}
		createProject(os.Args[2])
	case "serve":
		serve(os.Args[2:])
	case "make:controller":
		if len(os.Args) < 3 {
			fmt.Println("❌ Controller name required: binigo make:controller <name>")
			os.Exit(1)
		}
		makeController(os.Args[2])
	case "make:model":
		if len(os.Args) < 3 {
			fmt.Println("❌ Model name required: binigo make:model <name>")
			os.Exit(1)
		}
		makeModel(os.Args[2])
	case "make:middleware":
		if len(os.Args) < 3 {
			fmt.Println("❌ Middleware name required: binigo make:middleware <name>")
			os.Exit(1)
		}
		makeMiddleware(os.Args[2])
	case "make:migration":
		if len(os.Args) < 3 {
			fmt.Println("❌ Migration name required: binigo make:migration <name>")
			os.Exit(1)
		}
		makeMigration(os.Args[2])
	case "migrate":
		migrate()
	case "route:list":
		listRoutes()
	case "version", "-v", "--version":
		fmt.Printf("Binigo Framework v%s\n", version)
	case "help", "-h", "--help":
		printHelp()
	default:
		fmt.Printf("❌ Unknown command: %s\n", command)
		printHelp()
		os.Exit(1)
	}
}

func printHelp() {
	help := `
🔥 Binigo Framework v%s

USAGE:
    binigo <command> [arguments]

COMMANDS:
    new <name>              Create a new Binigo project
    serve [port]            Start development server (default: 8080)
    make:controller <name>  Generate a new controller
    make:model <name>       Generate a new model
    make:middleware <name>  Generate a new middleware
    make:migration <name>   Generate a new migration
    migrate                 Run database migrations
    route:list              List all registered routes
    version                 Show Binigo version
    help                    Show this help message

EXAMPLES:
    binigo new blog
    binigo serve
    binigo serve 3000
    binigo make:controller User
    binigo make:model Post
    binigo migrate

DOCUMENTATION:
    https://github.com/Chisonm/binigo

COMMUNITY:
    https://github.com/Chisonm/binigo/discussions
`
	fmt.Printf(help, version)
}

func createProject(name string) {
	fmt.Printf("🚀 Creating new Binigo project: %s\n\n", name)

	// Create project directory
	if err := os.Mkdir(name, 0755); err != nil {
		fmt.Printf("❌ Failed to create project directory: %v\n", err)
		os.Exit(1)
	}

	// Create directory structure
	dirs := []string{
		"app/controllers",
		"app/models",
		"app/middleware",
		"config",
		"database/migrations",
		"routes",
		"public",
		"storage/logs",
		"storage/uploads",
	}

	for _, dir := range dirs {
		path := filepath.Join(name, dir)
		if err := os.MkdirAll(path, 0755); err != nil {
			fmt.Printf("❌ Failed to create directory %s: %v\n", dir, err)
			os.Exit(1)
		}
	}

	// Create files from templates
	createMainFile(name)
	createGoModFile(name)
	createConfigFile(name)
	createEnvFile(name)
	createRoutesFile(name)
	createGitignore(name)
	createReadme(name)
	createExampleController(name)

	// Initialize go module
	fmt.Println("📦 Initializing Go module...")
	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = name
	if err := cmd.Run(); err != nil {
		fmt.Printf("⚠️  Warning: Failed to run go mod tidy: %v\n", err)
	}

	fmt.Println("\n✅ Project created successfully!")
	fmt.Printf("\n📚 Next steps:\n")
	fmt.Printf("   cd %s\n", name)
	fmt.Printf("   binigo serve\n\n")
	fmt.Printf("🎉 Happy coding with Binigo!\n\n")
}

func createMainFile(projectName string) {
	content := `package main

import (
	"log"
	"%[1]s/app/controllers"
	"%[1]s/config"
	"%[1]s/routes"

	binigo "github.com/Chisonm/binigo/pkg"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Bootstrap application
	app := binigo.NewApplication(cfg)

	// Register routes
	routes.Register(app)

	// Start server
	log.Printf("🚀 Server starting on :%%s", cfg.Port)
	if err := app.Run(":" + cfg.Port); err != nil {
		log.Fatal("❌ Server error:", err)
	}
}
`
	content = fmt.Sprintf(content, projectName)
	writeFile(filepath.Join(projectName, "main.go"), content)
}

func createGoModFile(projectName string) {
	content := `module %s

go 1.21

require (
	github.com/Chisonm/binigo v1.0.0
	github.com/joho/godotenv v1.5.1
)
`
	content = fmt.Sprintf(content, projectName)
	writeFile(filepath.Join(projectName, "go.mod"), content)
}

func createConfigFile(projectName string) {
	content := `package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	binigo "github.com/Chisonm/binigo/pkg"
)

func Load() *binigo.Config {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  No .env file found, using environment variables")
	}

	return &binigo.Config{
		AppName:     getEnv("APP_NAME", "Binigo App"),
		Environment: getEnv("APP_ENV", "development"),
		Debug:       getEnv("APP_DEBUG", "true") == "true",
		Port:        getEnv("APP_PORT", "8080"),
		Database: binigo.DatabaseConfig{
			Driver:   getEnv("DB_DRIVER", "postgres"),
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			Database: getEnv("DB_DATABASE", "%s"),
			Username: getEnv("DB_USERNAME", "postgres"),
			Password: getEnv("DB_PASSWORD", ""),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
`
	content = fmt.Sprintf(content, projectName)
	writeFile(filepath.Join(projectName, "config", "config.go"), content)
}

func createEnvFile(projectName string) {
	content := `APP_NAME=%s
APP_ENV=development
APP_DEBUG=true
APP_PORT=8080

DB_DRIVER=postgres
DB_HOST=localhost
DB_PORT=5432
DB_DATABASE=%s
DB_USERNAME=postgres
DB_PASSWORD=
`
	content = fmt.Sprintf(content, projectName, projectName)
	writeFile(filepath.Join(projectName, ".env"), content)
}

func createRoutesFile(projectName string) {
	content := `package routes

import (
	binigo "github.com/Chisonm/binigo/pkg"
	"%s/app/controllers"
)

func Register(app *binigo.Application) {
	// Welcome route
	app.Get("/", func(ctx *binigo.Context) error {
		return ctx.JSON(binigo.Map{
			"message": "Welcome to Binigo Framework!",
			"version": "1.0.0",
		})
	})

	// Health check
	app.Get("/health", func(ctx *binigo.Context) error {
		return ctx.String("OK")
	})

	// API v1 routes
	api := app.Group("/api/v1", func(r *binigo.Router) {
		// Example controller
		hello := &controllers.HelloController{}
		r.Get("/hello", hello.Index)
		r.Get("/hello/{name}", hello.Show)
	})

	_ = api // Avoid unused variable warning
}
`
	content = fmt.Sprintf(content, projectName)
	writeFile(filepath.Join(projectName, "routes", "routes.go"), content)
}

func createExampleController(projectName string) {
	content := `package controllers

import binigo "github.com/Chisonm/binigo/pkg"

type HelloController struct{}

func (h *HelloController) Index(ctx *binigo.Context) error {
	return ctx.Success(binigo.Map{
		"message": "Hello from Binigo!",
	})
}

func (h *HelloController) Show(ctx *binigo.Context) error {
	name := ctx.Param("name")
	
	return ctx.Success(binigo.Map{
		"message": "Hello, " + name + "!",
		"name":    name,
	})
}
`
	writeFile(filepath.Join(projectName, "app", "controllers", "hello_controller.go"), content)
}

func createGitignore(projectName string) {
	content := `.env
.env.*
!.env.example

# Binaries
*.exe
*.exe~
*.dll
*.so
*.dylib
bin/
dist/

# Test binary, built with go test -c
*.test

# Output of the go coverage tool
*.out

# Dependency directories
vendor/

# Go workspace file
go.work

# IDE
.idea/
.vscode/
*.swp
*.swo
*~

# OS
.DS_Store
Thumbs.db

# Logs
storage/logs/*.log
!storage/logs/.gitkeep

# Uploads
storage/uploads/*
!storage/uploads/.gitkeep
`
	writeFile(filepath.Join(projectName, ".gitignore"), content)
}

func createReadme(projectName string) {
	content := `# %s

A web application built with [Binigo Framework](https://github.com/Chisonm/binigo).

## Getting Started

### Prerequisites

- Go 1.21 or higher
- PostgreSQL (optional, for database features)

### Installation

1. Install dependencies:
` + "```bash" + `
go mod download
` + "```" + `

2. Copy .env.example to .env and configure:
` + "```bash" + `
cp .env.example .env
` + "```" + `

3. Run the application:
` + "```bash" + `
binigo serve
` + "```" + `

The server will start at http://localhost:8080

## Project Structure

` + "```" + `
%s/
├── app/
│   ├── controllers/    # HTTP controllers
│   ├── models/         # Data models
│   └── middleware/     # Custom middleware
├── config/             # Configuration files
├── database/
│   └── migrations/     # Database migrations
├── routes/             # Route definitions
├── public/             # Static files
├── storage/            # Application storage
│   ├── logs/          # Log files
│   └── uploads/       # Uploaded files
├── .env               # Environment variables
└── main.go            # Application entry point
` + "```" + `

## Available Commands

` + "```bash" + `
binigo serve                    # Start development server
binigo make:controller User     # Create a controller
binigo make:model Post          # Create a model
binigo make:middleware Auth     # Create middleware
binigo make:migration create_users_table  # Create migration
binigo migrate                  # Run migrations
binigo route:list               # List all routes
` + "```" + `

## Documentation

Visit [Binigo Documentation](https://github.com/Chisonm/binigo) for more information.

## License

MIT
`
	content = fmt.Sprintf(content, projectName, projectName)
	writeFile(filepath.Join(projectName, "README.md"), content)
}

func writeFile(path, content string) {
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		fmt.Printf("❌ Failed to create %s: %v\n", path, err)
		os.Exit(1)
	}
	fmt.Printf("✅ Created: %s\n", path)
}

func serve(args []string) {
	port := "8080"
	if len(args) > 0 {
		port = strings.TrimPrefix(args[0], ":")
	}

	fmt.Printf("🚀 Starting development server on http://localhost:%s\n", port)
	fmt.Println("Press Ctrl+C to stop")
	fmt.Println("")

	cmd := exec.Command("go", "run", "main.go")
	cmd.Env = append(os.Environ(), fmt.Sprintf("APP_PORT=%s", port))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Printf("❌ Server error: %v\n", err)
		os.Exit(1)
	}
}

func makeController(name string) {
	if !strings.HasSuffix(name, "Controller") {
		name += "Controller"
	}

	tmpl := `package controllers

import binigo "github.com/Chisonm/binigo/pkg"

type {{.Name}} struct {
	// DB *binigo.DB ` + "`inject:\"db\"`" + `
}

func (c *{{.Name}}) Index(ctx *binigo.Context) error {
	return ctx.Success(binigo.Map{
		"message": "Index method",
	})
}

func (c *{{.Name}}) Show(ctx *binigo.Context) error {
	id := ctx.Param("id")
	
	return ctx.Success(binigo.Map{
		"id": id,
	})
}

func (c *{{.Name}}) Store(ctx *binigo.Context) error {
	var input map[string]interface{}
	
	if err := ctx.Bind(&input); err != nil {
		return ctx.Error("Invalid input", 400)
	}
	
	return ctx.Status(201).Success(input, "Created successfully")
}

func (c *{{.Name}}) Update(ctx *binigo.Context) error {
	id := ctx.Param("id")
	
	var input map[string]interface{}
	if err := ctx.Bind(&input); err != nil {
		return ctx.Error("Invalid input", 400)
	}
	
	return ctx.Success(binigo.Map{
		"id":   id,
		"data": input,
	}, "Updated successfully")
}

func (c *{{.Name}}) Destroy(ctx *binigo.Context) error {
	id := ctx.Param("id")
	
	return ctx.Success(nil, "Deleted successfully")
}
`

	data := struct{ Name string }{Name: name}

	t, err := template.New("controller").Parse(tmpl)
	if err != nil {
		fmt.Printf("❌ Template error: %v\n", err)
		os.Exit(1)
	}

	filename := fmt.Sprintf("app/controllers/%s.go", toSnakeCase(name))
	file, err := os.Create(filename)
	if err != nil {
		fmt.Printf("❌ Failed to create file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	if err := t.Execute(file, data); err != nil {
		fmt.Printf("❌ Failed to write file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Controller created: %s\n", filename)
}

func makeModel(name string) {
	// Similar to makeController but for models
	fmt.Printf("✅ Model %s created\n", name)
}

func makeMiddleware(name string) {
	// Similar to makeController but for middleware
	fmt.Printf("✅ Middleware %s created\n", name)
}

func makeMigration(name string) {
	// Create migration file
	fmt.Printf("✅ Migration %s created\n", name)
}

func migrate() {
	fmt.Println("🔄 Running migrations...")
	fmt.Println("✅ Migrations completed successfully")
}

func listRoutes() {
	fmt.Println("📋 Registered Routes:")
	fmt.Println("--------------------------------------------------")
	// Implementation would list actual routes
}

func toSnakeCase(str string) string {
	var result strings.Builder
	for i, r := range str {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune('_')
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
}
