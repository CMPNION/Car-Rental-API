package database

import (
	"log"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/CMPNION/Car-Rental-API.git/internal/models"
)

// InitDB инициализирует файл базы данных и запускает миграции
func InitDB(filepath string) *gorm.DB {
	if filepath == "" {
		filepath = "car_rental.db"
	}

	db, err := gorm.Open(sqlite.Open(filepath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info), // Логируем все SQL запросы
	})

	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Автоматическое создание таблиц на основе структур (Auto-Migration)
	// Добавляйте сюда все ваши модели
	err = db.AutoMigrate(
		&models.User{},
		&models.Car{},
		&models.Rental{},
		&models.Transaction{},
	)

	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	log.Println("Database connection established and migrated")
	return db
}
