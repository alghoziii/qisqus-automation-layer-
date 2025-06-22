package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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
			Where("assigned = true AND agent_id = ?", agent.ID).
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
