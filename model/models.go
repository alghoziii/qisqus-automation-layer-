package model

import "time"

type Customer struct {
	ID        uint      `gorm:"primaryKey"`
	Name      string    `gorm:"not null"`
	RoomID    string    `gorm:"not null;unique"`
	Email     string    `gorm:"not null"`
	Status    string    `gorm:"default:waiting"`
	CreatedAt time.Time
}

type Queue struct {
	ID         uint      `gorm:"primaryKey"`
	CustomerID uint      `gorm:"not null"`
	RoomID     string    `gorm:"not null"`
	Assigned   bool      `gorm:"default:false"`
	AgentID    int64     `gorm:"default:0"`
	IsResolved bool      `gorm:"default:false"`
	Customer   Customer  `gorm:"foreignKey:CustomerID"` 
	CreatedAt  time.Time
}

type Agent struct {
	ID               int64  `json:"id"`
	Name             string `json:"name"`
	Email            string `json:"email"`
	IsAvailable      bool   `json:"is_available"`
	CurrentCustomers int    `json:"current_customer_count"`
	TypeAsString     string `json:"type_as_string"`
}
