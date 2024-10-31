package models

type App struct {
	ID     int    `gorm:"primaryKey;autoIncrement"`
	Name   string `gorm:"unique;not null"`
	Secret string `gorm:"not null"`
}
