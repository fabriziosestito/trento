package services

import (
	"fmt"

	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupTestDatabase() *gorm.DB {
	viper.AutomaticEnv()
	viper.SetDefault("dbhost", "localhost")
	viper.SetDefault("dbport", "32432")

	dsn := fmt.Sprintf("host=%s port=%s user=postgres password=postgres dbname=trento_test sslmode=disable", viper.GetString("dbhost"), viper.GetString("dbport"))
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	return db
}
