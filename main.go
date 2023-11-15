package main

import (
	"fmt"
	"log"
	"os"

	"github.com/dwilkolek/browse-together-api/db"
	"github.com/dwilkolek/browse-together-api/internal/server"
)

func main() {
	log.SetOutput(os.Stdout)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	env := os.Getenv("ENV")
	if env == "" {
		env = "prod"
	}

	db.GetDb()

	log.Printf("Running %s mode on port %s", env, port)

	server := server.New()

	log.Fatal(server.Listen(fmt.Sprintf(":%s", port)))
}
