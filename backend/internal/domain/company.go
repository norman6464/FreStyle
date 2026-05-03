package domain

import "time"

type Company struct {
	ID        uint64    `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"column:name;not null" json:"name"`
	CreatedAt time.Time `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updatedAt"`
}

func (Company) TableName() string { return "companies" }
