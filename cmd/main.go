package main

import (
	"log"
	"qiscus-agent-allocator/config"
	"qiscus-agent-allocator/controller"

	"github.com/gin-gonic/gin"
)

func main() {
	if err := config.InitDB(); err != nil {
		log.Fatal("Failed to connect to DB:", err)
	}

	router := gin.Default()
	router.POST("/webhook/agent_allocation", controller.WebhookHandler(config.DB))
	router.Run(":8080")
}
