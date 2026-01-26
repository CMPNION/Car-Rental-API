package models

import (
	"time"

	"gorm.io/gorm"
)

const (
	UserRoleAdmin     = "admin"
	UserRoleClient    = "client"
	UserRoleCorporate = "corporate"
)

const (
	CarCategoryEconomy  = "economy"
	CarCategoryBusiness = "business"
	CarCategoryLuxury   = "luxury"
)

const (
	CarStatusAvailable   = "available"
	CarStatusBooked      = "booked"
	CarStatusMaintenance = "maintenance"
)

const (
	RentalStatusPending   = "pending"
	RentalStatusActive    = "active"
	RentalStatusCompleted = "completed"
	RentalStatusCancelled = "cancelled"
)

const (
	TransactionStatusSuccess = "success"
	TransactionStatusFailed  = "failed"
)

type User struct {
	gorm.Model
	FirstName    string   `json:"first_name" gorm:"column:first_name" validate:"required"`
	LastName     string   `json:"last_name" gorm:"column:last_name" validate:"required"`
	Email        string   `json:"email" gorm:"column:email;uniqueIndex" validate:"required,email"`
	PasswordHash string   `json:"password_hash" gorm:"column:password_hash" validate:"required"`
	Role         string   `json:"role" gorm:"column:role" validate:"required,oneof=admin client corporate"`
	Balance      float64  `json:"balance" gorm:"column:balance" validate:"gte=0"`
	Rating       float64  `json:"rating" gorm:"column:rating" validate:"gte=0,lte=5"`
	Rentals      []Rental `json:"rentals" gorm:"foreignKey:UserID"`
}

type Car struct {
	gorm.Model
	Mark         string   `json:"mark" gorm:"column:mark" validate:"required"`
	CarModel     string   `json:"model" gorm:"column:model" validate:"required"`
	Category     string   `json:"category" gorm:"column:category" validate:"required,oneof=economy business luxury"`
	Status       string   `json:"status" gorm:"column:status" validate:"required,oneof=available booked maintenance"`
	PricePerHour float64  `json:"price_per_hour" gorm:"column:price_per_hour" validate:"required,gt=0"`
	Metadata     string   `json:"metadata" gorm:"column:metadata;type:text"`
	Rating       float64  `json:"rating" gorm:"column:rating" validate:"gte=0,lte=5"`
	Rentals      []Rental `json:"rentals" gorm:"foreignKey:CarID"`
}

type Rental struct {
	gorm.Model
	UserID      uint        `json:"user_id" gorm:"column:user_id;index" validate:"required"`
	CarID       uint        `json:"car_id" gorm:"column:car_id;index" validate:"required"`
	StartDate   time.Time   `json:"start_date" gorm:"column:start_date" validate:"required"`
	EndDate     time.Time   `json:"end_date" gorm:"column:end_date" validate:"required,gtfield=StartDate"`
	TotalPrice  float64     `json:"total_price" gorm:"column:total_price" validate:"required,gt=0"`
	Status      string      `json:"status" gorm:"column:status" validate:"required,oneof=pending active completed cancelled"`
	User        *User       `json:"user" gorm:"foreignKey:UserID"`
	Car         *Car        `json:"car" gorm:"foreignKey:CarID"`
	Transaction *Transaction `json:"transaction" gorm:"foreignKey:RentalID"`
}

type Transaction struct {
	gorm.Model
	RentalID uint    `json:"rental_id" gorm:"column:rental_id;uniqueIndex" validate:"required"`
	Amount   float64 `json:"amount" gorm:"column:amount" validate:"required,gt=0"`
	Status   string  `json:"status" gorm:"column:status" validate:"required,oneof=success failed"`
	Rental   *Rental `json:"rental" gorm:"foreignKey:RentalID"`
}
