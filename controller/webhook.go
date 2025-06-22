package controller

import (
	"log"
	"net/http"
	"os"
	"qiscus-agent-allocator/model"
	"qiscus-agent-allocator/utils"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type WebhookPayload struct {
	AppID     string `json:"app_id"`
	Source    string `json:"source"`
	Name      string `json:"name" binding:"required"`
	Email     string `json:"email" binding:"required,email"`
	RoomID    string `json:"room_id" binding:"required"`
}

func WebhookHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var payload WebhookPayload
		if err := c.ShouldBindJSON(&payload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
			return
		}

		secretKey := os.Getenv("QISCUS_SECRET_KEY")
		appID := os.Getenv("QISCUS_APP_ID")

		// Validasi room ID
		valid, err := utils.ValidateRoomID(payload.RoomID, secretKey, appID)
		if err != nil || !valid {
			c.JSON(http.StatusBadRequest, gin.H{"error": "room_id tidak valid"})
			return
		}

		// Cek jika customer sudah pernah masuk
		var existing model.Customer
		if err := db.Where("room_id = ?", payload.RoomID).First(&existing).Error; err == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "customer already exists"})
			return
		}

		// Simpan customer ke database
		customer := model.Customer{
			Name:      payload.Name,
			Email:     payload.Email,
			RoomID:    payload.RoomID,
			Status:    "waiting",
			CreatedAt: time.Now(),
		}
		if err := db.Create(&customer).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save customer"})
			return
		}

		// Cari agent yang tersedia
		agents, err := utils.GetAvailableAgents(secretKey, appID, 2, db)
		if err != nil {
			log.Println("Gagal ambil agent:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch agents"})
			return
		}

		// Jika tidak ada agent tersedia
		if len(agents) == 0 {
			log.Println("Tidak ada agent tersedia.")
			db.Create(&model.Queue{
				CustomerID: customer.ID,
				RoomID:     payload.RoomID,
				AgentID:    0,
				Assigned:   false,
				CreatedAt:  time.Now(),
			})
			c.JSON(http.StatusOK, gin.H{"message": "Customer queued. No agent available."})
			return
		}

		// Assign agent
		assignedAgent := agents[0]
		if err := utils.AssignAgentToRoom(payload.RoomID, assignedAgent.ID, secretKey, appID); err != nil {
			log.Println("Gagal assign agent:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "assign failed", "detail": err.Error()})
			return
		}

		// Simpan ke queue jika berhasil assign
		db.Create(&model.Queue{
			CustomerID: customer.ID,
			RoomID:     payload.RoomID,
			AgentID:    assignedAgent.ID,
			Assigned:   true,
			CreatedAt:  time.Now(),
		})

		db.Model(&customer).Update("status", "assigned")

		c.JSON(http.StatusOK, gin.H{"message": "Agent assigned", "agent": assignedAgent.Name})
	}
}
