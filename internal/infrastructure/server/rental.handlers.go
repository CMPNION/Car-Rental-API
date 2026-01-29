package server

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/CMPNION/Car-Rental-API.git/internal/auth/middleware"
	"github.com/CMPNION/Car-Rental-API.git/internal/models"
)

// RentalRequest describes the structure of the incoming JSON request
type RentalRequest struct {
	CarID     uint      `json:"car_id"`
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
}

// createRentalHandler handles POST /api/v1/rentals
func (s *Server) createRentalHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httpErr(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// 1. Get UserID from JWT context
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		httpErr(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req RentalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpErr(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	// 2. Date validation (Date Validator)
	now := time.Now().UTC()
	if req.StartDate.Before(now) {
		httpErr(w, http.StatusBadRequest, "The pickup date cannot be in the past. Please select a future date.")
		return
	}
	if req.EndDate.Sub(req.StartDate) < time.Hour {
		httpErr(w, http.StatusBadRequest, "Minimum rental duration is 1 hour. Please adjust your return time.")
		return
	}

	// 3. Load car and user data for price calculation
	var car models.Car
	if err := s.db.First(&car, req.CarID).Error; err != nil {
		httpErr(w, http.StatusNotFound, "Car not found")
		return
	}

	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		httpErr(w, http.StatusNotFound, "User not found")
		return
	}

	// 4. Overbooking check (call from rental_logic.go)
	available, err := s.CheckAvailability(req.CarID, req.StartDate, req.EndDate)
	if err != nil {
		httpErr(w, http.StatusInternalServerError, "database error")
		return
	}
	if !available {
		httpErr(w, http.StatusBadRequest, "Sorry, the car is already booked for these dates")
		return
	}

	// 5. Pricing Engine (call from rental_logic.go)
	finalPrice := CalculatePrice(car.PricePerHour, req.StartDate, req.EndDate, user.Rating)

	// 6. Create rental record with Pending status
	rental := models.Rental{
		UserID:     userID,
		CarID:      req.CarID,
		StartDate:  req.StartDate.UTC(),
		EndDate:    req.EndDate.UTC(),
		TotalPrice: finalPrice,
		Status:     models.RentalStatusPending,
	}

	if err := s.db.Create(&rental).Error; err != nil {
		httpErr(w, http.StatusInternalServerError, "could not create rental")
		return
	}

	// 7. Successful response
	writeJSON(w, http.StatusCreated, map[string]any{
		"rental_id":   rental.ID,
		"total_price": rental.TotalPrice,
		"status":      rental.Status,
		"message":     "Rental created. Please proceed to payment.",
	})
}
