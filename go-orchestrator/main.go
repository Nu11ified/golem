package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"go-orchestrator/routes"
	"go-orchestrator/schema"
)

func main() {
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
