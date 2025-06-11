package main

import (
	"fmt"
	"log"
	"net/http"

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
	log.Println("Go orchestrator listening on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
