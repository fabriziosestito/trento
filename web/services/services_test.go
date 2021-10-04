package services

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupTestDatabase() *gorm.DB {
	dsn := "host=localhost user=postgres password=postgres port=32432 dbname=trento_test sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	return db
}
