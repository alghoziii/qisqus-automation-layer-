package controller

import (
	"net/http"
	"os"
	"qiscus-agent-allocator/utils"
	"gorm.io/gorm"
	"github.com/gin-gonic/gin"
	"qiscus-agent-allocator/model"
	"log"
)

func ManualResolveHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var payload struct {
			RoomID string `json:"room_id" binding:"required"`
		}
		if err := c.ShouldBindJSON(&payload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload"})
			return
		}

		secretKey := os.Getenv("QISCUS_SECRET_KEY")
		appID := os.Getenv("QISCUS_APP_ID")

		// Panggil endpoint mark_as_resolved Qiscus
		if err := utils.MarkRoomAsResolved(payload.RoomID, secretKey, appID); err != nil {
			log.Println("Gagal mark resolved ke Qiscus:", err)
			// Masih lanjut update lokal
		}

		// Update status di DB lokal
		db.Model(&model.Queue{}).Where("room_id = ?", payload.RoomID).Update("is_resolved", true)
		db.Model(&model.Customer{}).Where("room_id = ?", payload.RoomID).Update("status", "resolved")

		// Proses antrean
		go utils.ProcessQueue(db)

		c.JSON(http.StatusOK, gin.H{"message": "Room resolved successfully and queue processed"})
	}
}
