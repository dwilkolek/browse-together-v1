package main

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/dwilkolek/browse-together/internal/server"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/gofiber/template/html/v2"
)

//go:embed static views
var embedFs embed.FS

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	env := os.Getenv("ENV")
	if env == "" {
		env = "prod"
	}

	log.Printf("Running %s mode on port %s", env, port)

	viewsFs, _ := fs.Sub(embedFs, "views")
	engine := html.NewFileSystem(http.FS(viewsFs), ".html")
	server := server.New(engine)
	if env != "dev" {
		server.RedirectToHttpsWww()
	}
	if env == "dev" {
		server.Static("/static", "./static", fiber.Static{
			Compress:      false,
			CacheDuration: 1 * time.Second,
		})
		engine = html.New("./views", ".html")
		engine.Reload(true)
	} else {
		server.Use("/static", filesystem.New(filesystem.Config{
			Root:       http.FS(embedFs),
			PathPrefix: "static",
		}))
	}

	server.Get("/robots.txt", func(c *fiber.Ctx) error {
		content, _ := embedFs.ReadFile("static/robots.txt")
		return c.Send(content)
	})

	server.Get("/", func(c *fiber.Ctx) error {
		return c.Render("index", fiber.Map{})
	})

	log.Fatal(server.Listen(fmt.Sprintf(":%s", port)))
}
