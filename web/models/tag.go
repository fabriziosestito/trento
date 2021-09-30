package models

type Tag struct {
	Value        string `gorm:"primaryKey"`
	ResourceType string `gorm:"primaryKey"`
	ResourceId   string `gorm:"primaryKey"`
}
