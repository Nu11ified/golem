package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"go-orchestrator/routes"
	"go-orchestrator/schema"
)

func spawnNodeRenderer() (*exec.Cmd, error) {
	cmd := exec.Command("node", "../node-renderer/renderer.ts")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(), "NODE_ENV=production")
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	return cmd, nil
}

func main() {
	var nodeCmd *exec.Cmd
	var err error
	if os.Getenv("SPAWN_NODE_RENDERER") == "1" {
		nodeCmd, err = spawnNodeRenderer()
		if err != nil {
			log.Fatalf("Failed to start Node renderer: %v", err)
		}
		// Ensure Node renderer is killed on exit
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		go func() {
			<-c
			if nodeCmd != nil && nodeCmd.Process != nil {
				nodeCmd.Process.Kill()
			}
			os.Exit(1)
		}()
	}

	schemaJSON, err := schema.GetExampleSchema()
	if err != nil {
		log.Fatalf("Failed to generate UI schema: %v", err)
	}
	fmt.Println("Example Platform-Agnostic UI Schema:\n", string(schemaJSON))

	r := routes.SetupRouter()
	port := os.Getenv("GO_PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Go orchestrator listening on http://localhost:%s", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
