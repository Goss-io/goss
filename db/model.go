package db

import "time"

//Model .
type Model struct {
	ID         int       `gorm:"primary_key" json:"id"`
	CreatedAt  time.Time `gorm:"created_at"`
	UpdatedAt  time.Time `gorm:"created_at"`
	CreateTime string    `gorm:"-"`
	UpdateTime string    `gorm:"-"`
}
