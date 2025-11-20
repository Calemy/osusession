package main

import (
	"errors"
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v3"
)

func main() {
	app := fiber.New()

	sess := sessions

	app.Use(func(c fiber.Ctx) error {
		if c.Request().URI().String() == "http://127.0.0.1:8080/favicon.ico" {
			return c.Next()
		}
		fmt.Println("REQUEST INCOMING: ", c.Request().URI())
		return c.Next()
	})

	app.Get("/", func(c fiber.Ctx) error {
		return c.SendString(":3")
	})

	app.Get("/session/:usernames", func(c fiber.Ctx) error {
		username := strings.ReplaceAll(c.Params("usernames"), " ", "_")

		if username == "" {
			return errors.New("Go fuck yourself")
		}

		session := sess.Get(username)
		if session == nil {
			fmt.Println("Creating new session")
			s, err := CreateSession(username)
			if err != nil {
				return err
			}
			return c.JSON(s)
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
