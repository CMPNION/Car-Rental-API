package server

import (
	"math"
	"time"

	"github.com/CMPNION/Car-Rental-API.git/internal/models"
	"gorm.io/gorm"
)

// CalculatePrice is our isolated Pricing Engine
func CalculatePrice(basePricePerHour float64, start, end time.Time, userRating float64) float64 {
	durationHours := end.Sub(start).Hours()
	baseTotal := basePricePerHour * durationHours

	modifier := 1.0
	// 10% discount for high-rated users (rating > 4.5)
	if userRating > 4.5 {
		modifier = 0.90
	} else if userRating < 2.0 && userRating > 0 {
		// 20% surcharge for risky drivers
		modifier = 1.20
	}

	return math.Round(baseTotal*modifier*100) / 100
}

// CheckAvailability is an isolated check for overbooking
func (s *Server) CheckAvailability(carID uint, start, end time.Time) (bool, error) {
	var existingRental models.Rental
	// Formula for interval intersection
	err := s.db.Where("car_id = ? AND status != ? AND start_date < ? AND end_date > ?",
		carID, models.RentalStatusCancelled, end, start).First(&existingRental).Error

	if err == nil {
		return false, nil // Match found, car is occupied
	}
	if err == gorm.ErrRecordNotFound {
		return true, nil // No matches found, car is available
	}
	return false, err // Database error
}
