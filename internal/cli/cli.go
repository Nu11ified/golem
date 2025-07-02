package cli

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/Nu11ified/golem/internal/build"
	"github.com/Nu11ified/golem/internal/config"
	"github.com/Nu11ified/golem/internal/dev"
	"github.com/Nu11ified/golem/internal/server"
)

// RunDev starts the development server with hot reload
func RunDev() {
	fmt.Println("🚀 Starting Golem development server...")

	config, err := loadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	devServer := dev.NewServer(config)
	if err := devServer.Start(); err != nil {
		log.Fatalf("Failed to start dev server: %v", err)
	}
}

// RunBuild builds the production-ready application
func RunBuild() {
	fmt.Println("🔨 Building Golem application...")

	config, err := loadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	builder := build.NewBuilder(config)
	if err := builder.Build(); err != nil {
		log.Fatalf("Build failed: %v", err)
	}

	fmt.Println("✅ Build completed successfully!")
}

// RunStart starts the production server
func RunStart() {
	fmt.Println("🌟 Starting Golem production server...")

	config, err := loadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	prodServer := server.NewServer(config)
	if err := prodServer.Start(); err != nil {
		log.Fatalf("Failed to start production server: %v", err)
	}
}

// RunNew creates a new Golem project
func RunNew(projectName string) {
	fmt.Printf("✨ Creating new Golem project: %s\n", projectName)

	if err := createProject(projectName); err != nil {
		log.Fatalf("Failed to create project: %v", err)
	}

	fmt.Printf("✅ Project '%s' created successfully!\n", projectName)
	fmt.Printf("   cd %s\n", projectName)
	fmt.Printf("   golem dev\n")
}

func loadConfig() (*config.Config, error) {
	configPath := "golem.config.json"
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("golem.config.json not found. Run 'golem new <project-name>' to create a new project")
	}

	return config.Load(configPath)
}

func createProject(projectName string) error {
	// Check if directory already exists
	if _, err := os.Stat(projectName); !os.IsNotExist(err) {
		return fmt.Errorf("directory '%s' already exists", projectName)
	}

	// Create directory structure
	dirs := []string{
		projectName,
		filepath.Join(projectName, ".golem", "build"),
		filepath.Join(projectName, ".golem", "dev"),
		filepath.Join(projectName, ".golem", "types"),
		filepath.Join(projectName, "src", "app"),
		filepath.Join(projectName, "src", "components"),
		filepath.Join(projectName, "src", "server"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %v", dir, err)
		}
	}

	// Create template files
	if err := createTemplateFiles(projectName); err != nil {
		return fmt.Errorf("failed to create template files: %v", err)
	}

	return nil
}
