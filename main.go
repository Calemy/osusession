package main

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	app := gin.Default()
	app.GET("/session/:username", func(ctx *gin.Context) {
		username := ctx.Param("username")

		if username == "" {
			ctx.String(500, "go fuck yourself")
		}

		user := users.Get(username)
		if user == 0 {
			s, err := CreateSession(username)
			if err != nil {
				ctx.String(500, err.Error())
			}
			ctx.JSON(200, s)
			return
		}
		session := sessions.Get(user)
		if session == nil {
			s, err := CreateSession(username)
			if err != nil {
				ctx.String(500, "gay sex")
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

		log.Fatalf("failed to run server: %v", err)
	}
}
