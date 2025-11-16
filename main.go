package main

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v3"
)

func main() {
	app := fiber.New()

	app.Get("/", func(c fiber.Ctx) error {
		return c.SendString(":3")
	})

	app.Get("/session/:usernames", func(c fiber.Ctx) error {
		username := strings.ReplaceAll(c.Params("usernames"), " ", "_")

		if username == "" {
			return errors.New("Go fuck yourself")
		}

		session := GetSession(username)
		if err := session.Fetch(); err != nil {
			return err
		}

		return c.JSON(session)
	})

	app.Get("/sessions/:id", func(c fiber.Ctx) error {
		return c.SendString(".")
	})

	app.Get("/track", func(c fiber.Ctx) error {
		return c.SendString("here is your code")
	})

	app.Listen(":8080")
}
