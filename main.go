package main

import (
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/joho/godotenv/autoload"
)

func main() {
	app := gin.Default()
	app.GET("/session/:username", func(ctx *gin.Context) {
		username := ctx.Param("username")

		if username == "" {
			ctx.JSON(400, Error("No username provided"))
		}

		user := users.Get(username)
		if user == 0 {
			s, err := CreateSession(username)
			if err != nil {
				ctx.JSON(500, Error(err.Error()))
			}
			ctx.JSON(200, s)
			return
		}
		session := sessions.Get(user)
		if session == nil {
			s, err := CreateSession(username)
			if err != nil {
				ctx.JSON(500, Error(err.Error()))
				return
			}

			ctx.JSON(200, s)
			return
		}

		ctx.JSON(200, session)
	})

	go func() {
		ticker := time.NewTicker(15 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			GetScores()
		}
	}()

	if err := app.Run(os.Getenv("PORT")); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
}
