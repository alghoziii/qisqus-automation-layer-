package main

import (
	"qiscus-agent-allocator/config"
	"qiscus-agent-allocator/controller"

	"github.com/gin-gonic/gin"
)

func main() {
	config.InitDB()

	router := gin.Default()
	router.POST("/webhook/agent_allocation", controller.WebhookHandler(config.DB))
	router.Run(":8080")
}
