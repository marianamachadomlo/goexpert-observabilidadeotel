package main

import (
	"log"
	"os"

	"github.com/marianamachado/observabilidade-otel/internal/servicea"
	"github.com/marianamachado/observabilidade-otel/internal/serviceb"
)

func main() {
	switch os.Getenv("SERVICE") {
	case "a", "service-a":
		servicea.Run()
	case "b", "service-b":
		serviceb.Run()
	default:
		log.Fatal("SERVICE env is required: use 'a' for Service A or 'b' for Service B")
	}
}
