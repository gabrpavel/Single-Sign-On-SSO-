package models

type User struct {
	ID       int64  `gorm:"primaryKey;autoIncrement"`
	Email    string `gorm:"unique;not null"`
	PassHash []byte `gorm:"not null"`
	IsAdmin  bool   `gorm:"default:false"`
}
