package database

import (
	"fmt"
	"log"
	"os"

	"github.com/techmaster-vietnam/dd_goshare/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var dbInstance *gorm.DB

func Init(config *config.DBConfig, DBMigrator func(*gorm.DB) error) (*gorm.DB, error) {
	uri := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=Asia/Ho_Chi_Minh",
		config.DBHost, config.DBUser, config.DBPassword, config.DBName, config.DBPort, config.DBSSLMode,
	)

	var err error
	dbInstance, err = gorm.Open(postgres.Open(uri), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
		return nil, err
	}

	// Sử dụng hàm DBMigrator để migrate
	if err := DBMigrator(dbInstance); err != nil {
		log.Fatalf("failed to migrate database: %v", err)
		return nil, err
	}

	log.Println("✅ Database connected & migrated successfully!")
	return dbInstance, nil
}

func GetDB() *gorm.DB {
	if dbInstance == nil {
		log.Println("Database not initialized. Call Init first.")
		os.Exit(1)
	}
	return dbInstance
}