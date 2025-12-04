package database

import (
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"youlai-gin/config"
)

var DB *gorm.DB

func Init() {
	dsn := config.Dsn()
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}
	DB = db
}
