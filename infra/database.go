package infra

import (
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func connectDatabase(dsn string) *gorm.DB {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("infra: failed to connect database: %v", err)
	}

	log.Println("infra: database connected")

	return db
}
