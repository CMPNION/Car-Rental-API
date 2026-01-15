package database

import (
	"log"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// InitDB инициализирует файл базы данных и запускает миграции
func InitDB(filepath string) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(filepath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info), // Логируем все SQL запросы
	})

	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Автоматическое создание таблиц на основе структур (Auto-Migration)
	// Добавляйте сюда все ваши модели
	err = db.AutoMigrate(
		&User{},
		&Car{},
		&Booking{},
	)

	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	log.Println("Database connection established and migrated")
	return db
}


type Car struct {
	gorm.Model
	Brand  string `json:"brand"`
    CarModel  string `json:"carModel"`
	Status string `json:"status" gorm:"default:'available'"` // available, rented, maintenance
}

type User struct {
	gorm.Model
	Email    string `gorm:"uniqueIndex"`
	Password string `json:"-"`
	Role     string `gorm:"default:'client'"`
}

type Booking struct {
	gorm.Model
	UserID uint
	CarID  uint
	Start  string
	End    string
}
