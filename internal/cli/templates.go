package cli

import (
	"os"
	"path/filepath"
)

// createTemplateFiles creates the initial template files for a new project
func createTemplateFiles(projectName string) error {
	templates := map[string]string{
		"golem.config.json":        getConfigTemplate(projectName),
		"go.mod":                   getGoModTemplate(projectName),
		"package.json":             getPackageJSONTemplate(projectName),
		"src/app/main.go":          getMainGolemTemplate(projectName),
		"src/components/Button.go": getButtonComponentTemplate(),
		"src/server/hello.go":      getServerFunctionTemplate(),
		".gitignore":               getGitignoreTemplate(),
		"README.md":                getReadmeTemplate(projectName),
		"LICENSE":                  getLicenseTemplate(),
	}

	for path, content := range templates {
		fullPath := filepath.Join(projectName, path)
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			return err
		}
	}

	return nil
}

func getConfigTemplate(projectName string) string {
	return `{
  "projectName": "` + projectName + `",
  "version": "0.1.0",
  "entry": "src/app/main.go",
  "output": ".golem/build",
  "dev": {
    "port": 3000,
    "hotReload": true,
    "watch": ["src/**/*.golem", "src/**/*.go"]
  },
  "build": {
    "minify": true,
    "target": "es2020",
    "sourcemap": true
  },
  "server": {
    "grpc": {
      "port": 50051,
      "reflection": true
    },
    "functions": "src/server"
  },
  "wasm": {
    "optimizeSize": true,
    "enableFeatures": ["bulk-memory", "mutable-globals"]
  }
}`
}

func getGoModTemplate(projectName string) string {
	return `module github.com/` + "user/" + projectName + `

go 1.22

require (
	github.com/Nu11ified/golem v0.1.0
)`
}

func getPackageJSONTemplate(projectName string) string {
	return `{
  "name": "` + projectName + `",
  "version": "0.1.0",
  "description": "A Golem application",
  "scripts": {
    "dev": "golem dev",
    "build": "golem build",
    "start": "golem start"
  },
  "devDependencies": {
    "nodemon": "^3.0.2"
  }
}`
}

func getMainGolemTemplate(projectName string) string {
	return `package main

import (
	"github.com/Nu11ified/golem/dom"
	"` + projectName + `/src/components"
)

type AppState struct {
	Count int ` + "`json:\"count\"`" + `
}

func (s *AppState) Increment() {
	s.Count++
}

func (s *AppState) Decrement() {
	s.Count--
}

func App() *dom.Element {
	state := &AppState{Count: 0}
	
	return dom.Div(
		dom.Class("app"),
		dom.H1("Welcome to Golem! 🚀"),
		dom.P("Build reactive web apps with pure Go"),
		
		dom.Div(
			dom.Class("counter"),
			dom.H2("Counter Example"),
			dom.P("Count: ", dom.Text(state.Count)),
			
			dom.Button(
				dom.Text("Increment"),
				dom.OnClick(state.Increment),
			),
			dom.Button(
				dom.Text("Decrement"), 
				dom.OnClick(state.Decrement),
			),
		),
		
		components.Button(components.ButtonProps{
			Text: "Demo Button",
			Variant: "primary",
			OnClick: func() {
				dom.Alert("Hello from Golem!")
			},
		}),
	)
}

func main() {
	dom.Render(App(), "#app")
}`
}

func getButtonComponentTemplate() string {
	return `package components

import "github.com/Nu11ified/golem/dom"

type ButtonProps struct {
	Text     string
	OnClick  func()
	Variant  string // "primary", "secondary", "danger"
	Disabled bool
}

func Button(props ButtonProps) *dom.Element {
	class := "btn"
	if props.Variant != "" {
		class += " btn-" + props.Variant
	}
	if props.Disabled {
		class += " btn-disabled"
	}
	
	return dom.Button(
		dom.Class(class),
		dom.Text(props.Text),
		dom.OnClick(props.OnClick),
		dom.If(props.Disabled, dom.Disabled(true)),
	)
}`
}

func getServerFunctionTemplate() string {
	return `package server

import (
	"context"
	"fmt"
)

// Hello is a server function that can be called from the client
func Hello(name string) string {
	return fmt.Sprintf("Hello, %s! This message is from the Go server.", name)
}

// GetUserProfile fetches user profile data
func GetUserProfile(ctx context.Context, userID int) (*UserProfile, error) {
	// Simulate database lookup
	return &UserProfile{
		ID:   userID,
		Name: "John Doe",
		Email: "john@example.com",
	}, nil
}

type UserProfile struct {
	ID    int    ` + "`json:\"id\"`" + `
	Name  string ` + "`json:\"name\"`" + `
	Email string ` + "`json:\"email\"`" + `
}`
}

func getLicenseTemplate() string {
	return `MIT License

Copyright (c) 2024 Your Name

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
`
}

func getGitignoreTemplate() string {
	return `# Golem build outputs
.golem/
golem

# Go
vendor/
*.exe
*.exe~
*.dll
*.so
*.dylib
*.test
*.out
go.work

# Node
node_modules/
npm-debug.log*
yarn-debug.log*
yarn-error.log*
pnpm-debug.log*

# OS
.DS_Store
.DS_Store?
._*
.Spotlight-V100
.Trashes
ehthumbs.db
Thumbs.db

# IDE
.vscode/
.idea/
*.swp
*.swo

# Logs
*.log`
}

func getReadmeTemplate(projectName string) string {
	return `# ` + projectName + `

A Golem application - reactive web apps built with pure Go and WebAssembly.

This project was generated by the [Golem Framework](https://github.com/Nu11ified/golem).

## Getting Started

1. **Run the development server:**
   ` + "```sh" + `
   golem dev
   ` + "```" + `

2. **Open your browser** to [http://localhost:3000](http://localhost:3000) to see your app.

## Available Commands

- ` + "`golem dev`" + ` - Start the development server with hot-reloading.
- ` + "`golem build`" + ` - Build the application for production.
- ` + "`golem start`" + ` - Run the production server.
`
}
