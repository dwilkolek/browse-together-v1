package main

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/dwilkolek/browse-together/internal"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
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

	app := fiber.New(fiber.Config{
		Views: engine,
	})

	internal.SetupWebsockets(app)

	if env == "dev" {
		app.Static("/static", "./static", fiber.Static{
			Compress:      false,
			CacheDuration: 1 * time.Second,
		})
		engine = html.New("./views", ".html")
		engine.Reload(true)
	} else {
		app.Use("/static", filesystem.New(filesystem.Config{
			Root:       http.FS(embedFs),
			PathPrefix: "static",
		}))
	}

	app.Use(compress.New())
	app.Use(func(c *fiber.Ctx) error {
		host := string(c.Request().Host())
		uri := string(c.Request().RequestURI())
		protocol := c.Protocol()
		redirect := false
		if env != "dev" && protocol == "http" {
			redirect = true
			protocol = "https"
		}

		if strings.HasPrefix(host, "www.") {
			redirect = true
			host = host[4:]
		}

		c.Response().Header.Add("X-XSS-Protection", "1; mode=block")
		c.Response().Header.Add("X-Frame-Options", "DENY")
		c.Response().Header.Add("X-Content-Type-Options", "nosniff")
		c.Response().Header.Add("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")

		if redirect {
			target := fmt.Sprintf("%s://%s%s", protocol, host, uri)
			log.Println("Redirecting to: " + target)
			return c.Redirect(target, http.StatusMovedPermanently)
		}

		return c.Next()
	})

	app.Get("/robots.txt", func(c *fiber.Ctx) error {
		content, _ := embedFs.ReadFile("static/robots.txt")
		return c.Send(content)
	})

	app.Get("/", func(c *fiber.Ctx) error {
		return c.Render("index", fiber.Map{})
	})

	internal.SetupSessionApi(app)

	log.Fatal(app.Listen(fmt.Sprintf(":%s", port)))
}
