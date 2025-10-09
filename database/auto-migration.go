package database

import (
	"fmt"

	"github.com/techmaster-vietnam/dd_goshare/pkg/models"
	"gorm.io/gorm"
)

func DBMigrator(db *gorm.DB) error {
	if db == nil {
		return fmt.Errorf("db is nil")
	}

	if err := db.AutoMigrate(
		&models.Customer{},
		&models.Employee{},
		&models.Topic{},
		&models.Dialog{},
		&models.Audio{},
		&models.Image{},
		&models.Word{},
		&models.WordInDialog{},
		&models.FillInBlank{},
		&models.Tag{},
		&models.DialogTag{},
		&models.Comment{},
		&models.CustomerDialogProgress{},
		&models.CustomerDialogExercise{},
		&models.Achievement{},
		&models.CustomerAchievement{},
		&models.Subscription{},
		&models.CustomerSubscription{},
		&models.Payment{},
		&models.PaymentLog{},
		&models.HistoryLog{},
		&models.SystemStatistics{},
		&models.UserStatistics{},
		&models.DialogStatistics{},
		// Authentication models
		&models.AuthProvider{},
		// RBAC models
		&models.Role{},
		&models.Rule{},
		&models.UserRole{},
		&models.RuleRole{},
	); err != nil {
		return fmt.Errorf("automigrate failed: %w", err)
	}

	// Seed business function rules only (no roles or assignments)
	// if err := SeedBusinessFunctionRules(db); err != nil {
	// 	log.Printf("Warning: Failed to seed business function rules: %v", err)
	// }

	return nil
}
