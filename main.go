package main

import (
	"fmt"
	service "github.com/asifrahaman13/bhagabad_gita/internal/core/services"
	"github.com/asifrahaman13/bhagabad_gita/internal/handlers"
	"github.com/asifrahaman13/bhagabad_gita/internal/repository"
	"github.com/asifrahaman13/bhagabad_gita/internal/routes"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"log"
	"time"
	"github.com/gorilla/websocket"
	"net/http"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
	parent_route := gin.Default()

	parent_route.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	routes.InitializeRoutes(parent_route)

	parent_route.GET("/ws", func(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		go routes.HandleWebSocketConnection(conn)
	})
	log.Fatal(parent_route.Run())
}

func run() error {
	db, err := repository.InitializeDB()
	if err != nil {
		panic(err)
	}
	fmt.Println(db)
	userRep := repository.UserRepo.Initialize(db)
	users := service.InitializeUserService(userRep)
	handlers.UserHandler.Initialize(users)
	return nil
}
