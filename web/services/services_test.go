package services

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupTestDatabase() *gorm.DB {
	dsn := "host=localhost user=postgres password=postgres dbname=trento port=32432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	return db
}
