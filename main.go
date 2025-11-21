package main

import (
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/joho/godotenv/autoload"
)

func main() {
	gin.SetMode(gin.ReleaseMode)
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

	app.GET("/sessions", func(ctx *gin.Context) {
		ctx.JSON(200, users.m)
	})

	app.GET("/sessions/:usernames", func(ctx *gin.Context) {
		usernameParam := ctx.Param("usernames")

		if usernameParam == "" {
			ctx.JSON(400, Error("No username provided"))
		}

		usernames := strings.Split(usernameParam, ",")

		results := make([]*Session, len(usernames))

		var wg sync.WaitGroup

		wg.Add(len(usernames))

		for i, v := range usernames {
			go func(idx int) {
				defer wg.Done()
				user := users.Get(v)
				if user == 0 {
					s, err := CreateSession(v)
					if err != nil {
						return
					}
					results[idx] = s
					return
				}

				session := sessions.Get(user)
				if session == nil {
					s, err := CreateSession(v)
					if err != nil {
						return
					}
					results[idx] = s
					return
				}

				results[idx] = session

			}(i)
		}

		wg.Wait()

		ctx.JSON(200, results)
	})

	go func() {
		ticker := time.NewTicker(15 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			GetScores()
		}
	}()

	go func() {
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			ClearSessions()
		}
	}()

	if err := app.Run(os.Getenv("BIND")); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
}
