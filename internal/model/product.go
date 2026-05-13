package model

import "time"

// Product represents an item in inventory.
type Product struct {
    ID          uint      `gorm:"primaryKey" json:"id"`
    Name        string    `gorm:"size:255;not null" json:"name"`
    Description string    `gorm:"type:text" json:"description"`
    Category    string    `gorm:"size:100;index" json:"category"`
    Quantity    int       `gorm:"not null;default:0" json:"quantity"`
    Price       float64   `gorm:"type:numeric" json:"price"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}
