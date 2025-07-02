package main

import (
	"fmt"
	"os"

	"github.com/Nu11ified/golem/internal/cli"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "dev":
		cli.RunDev()
	case "build":
		cli.RunBuild()
	case "start":
		cli.RunStart()
	case "new":
		if len(os.Args) < 3 {
			fmt.Println("Usage: golem new <project-name>")
			os.Exit(1)
		}
		cli.RunNew(os.Args[2])
	case "version", "-v", "--version":
		fmt.Println("Golem Framework v0.1.0")
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`Golem Framework - Pure Go WebAssembly Web Framework

Usage:
  golem <command> [options]

Commands:
  dev      Start development server with hot reload
  build    Build production-ready application  
  start    Start production server
  new      Create new Golem project
  version  Show version information
  help     Show this help message

Examples:
  golem new my-app
  golem dev
  golem build
  golem start`)
}
