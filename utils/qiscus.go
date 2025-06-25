package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"qiscus-agent-allocator/model"

	"gorm.io/gorm"
)

var qiscusBaseURL = "https://omnichannel.qiscus.com"

func GetAvailableAgents(secretKey, appID string, maxCustomers int, db *gorm.DB) ([]model.Agent, error) {
	url := qiscusBaseURL + "/api/v2/admin/agents"

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Qiscus-Secret-Key", secretKey)
	req.Header.Set("Qiscus-App-Id", appID)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("get agents failed with status: %d", res.StatusCode)
	}

	var result struct {
		Data struct {
			Agents []model.Agent `json:"agents"`
		} `json:"data"`
	}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, err
	}

	available := []model.Agent{}
	for _, agent := range result.Data.Agents {
		// Syarat wajib: agent harus aktif dan bertipe "agent"
		if agent.TypeAsString != "agent" || !agent.IsAvailable {
			continue
		}

		// Hitung berapa customer aktif dari queue
		var count int64
		db.Model(&model.Queue{}).
		Where("assigned = true AND agent_id = ? AND is_resolved = false", agent.ID).
		Count(&count)

		agent.CurrentCustomers = int(count)
		if count < int64(maxCustomers) {
			available = append(available, agent)
		}
	}

	return available, nil
}

func AssignAgentToRoom(roomID string, agentID int64, secretKey, appID string) error {
	url := qiscusBaseURL + "/api/v1/admin/service/assign_agent"

	payload := map[string]interface{}{
		"room_id":  roomID,
		"agent_id": agentID,
	}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Qiscus-Secret-Key", secretKey)
	req.Header.Set("Qiscus-App-Id", appID)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		respBody, _ := io.ReadAll(res.Body)
		return fmt.Errorf("assign failed: %s", respBody)
	}

	return nil
}

func ValidateRoomID(roomID, secretKey, appID string) (bool, error) {
	url := qiscusBaseURL + "/api/v2/customer_rooms/" + roomID

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Qiscus-Secret-Key", secretKey)
	req.Header.Set("Qiscus-App-Id", appID)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, err
	}
	defer res.Body.Close()

	if res.StatusCode == 200 {
		return true, nil
	}
	return false, fmt.Errorf("room_id %s tidak valid (status: %d)", roomID, res.StatusCode)
}

func ProcessQueue(db *gorm.DB) {
	secretKey := os.Getenv("QISCUS_SECRET_KEY")
	appID := os.Getenv("QISCUS_APP_ID")

	var queues []model.Queue
	if err := db.Where("assigned = false AND is_resolved = false").
		Order("created_at ASC").
		Find(&queues).Error; err != nil {
		return
	}

	for _, q := range queues {
		var customer model.Customer
		db.First(&customer, q.CustomerID)

		agents, err := GetAvailableAgents(secretKey, appID, 2, db)
		if err != nil || len(agents) == 0 {
			continue
		}

		selectedAgent := agents[0]

		err = AssignAgentToRoom(q.RoomID, selectedAgent.ID, secretKey, appID)
		if err != nil {
			continue
		}

		db.Model(&q).Updates(map[string]interface{}{
			"assigned": true,
			"agent_id": selectedAgent.ID,
		})
		db.Model(&customer).Update("status", "assigned")
	}
}


func MarkRoomAsResolved(roomID, secretKey, appID string) error {
	url := qiscusBaseURL + "/api/v1/admin/service/mark_as_resolved"

	payload := map[string]interface{}{
		"room_id": roomID,
	}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Qiscus-Secret-Key", secretKey)
	req.Header.Set("Qiscus-App-Id", appID)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		respBody, _ := io.ReadAll(res.Body)
		return fmt.Errorf("mark as resolved failed: %s", respBody)
	}

	return nil
}
