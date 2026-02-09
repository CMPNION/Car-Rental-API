package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/CMPNION/Car-Rental-API.git/internal/entity"
	authhttp "github.com/CMPNION/Car-Rental-API.git/internal/interface/http/auth"
	"gorm.io/gorm"
)

// RentalRequest describes the structure of the incoming JSON request
type RentalRequest struct {
	CarID     uint      `json:"car_id"`
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
}

// rentalsHandler handles GET/POST /api/v1/rentals
func (s *Server) rentalsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.listRentals(w, r)
	case http.MethodPost:
		s.createRental(w, r)
	default:
		RespondWithError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

// rentalActionHandler handles POST /api/v1/rentals/{id}/pay|finish|cancel
func (s *Server) rentalActionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		RespondWithError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	id, action, err := parseRentalAction(r.URL.Path)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "invalid rental path")
		return
	}

	switch action {
	case "pay":
		s.payRental(w, r, id)
	case "finish":
		s.finishRental(w, r, id)
	case "cancel":
		s.cancelRental(w, r, id)
	default:
		RespondWithError(w, http.StatusNotFound, "unknown action")
	}
}

func (s *Server) createRental(w http.ResponseWriter, r *http.Request) {
	userID, ok := authhttp.UserIDFromContext(r.Context())
	if !ok {
		RespondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req RentalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondWithError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	now := time.Now().UTC()
	if req.StartDate.Before(now) {
		RespondWithError(w, http.StatusBadRequest, "pickup date cannot be in the past")
		return
	}
	if req.EndDate.Sub(req.StartDate) < time.Hour {
		RespondWithError(w, http.StatusBadRequest, "minimum rental duration is 1 hour")
		return
	}

	var created entity.Rental
	err := s.db.Transaction(func(tx *gorm.DB) error {
		var car entity.Car
		if err := tx.First(&car, req.CarID).Error; err != nil {
			return err
		}
		if car.Status != entity.CarStatusAvailable {
			return errors.New("car not available")
		}

		var user entity.User
		if err := tx.First(&user, userID).Error; err != nil {
			return err
		}

		available, err := checkAvailabilityWithDB(tx, req.CarID, req.StartDate, req.EndDate)
		if err != nil {
			return err
		}
		if !available {
			return errors.New("car already booked")
		}

		finalPrice := CalculatePrice(car.PricePerHour, req.StartDate, req.EndDate, user.Rating)
		created = entity.Rental{
			UserID:     userID,
			CarID:      req.CarID,
			StartDate:  req.StartDate.UTC(),
			EndDate:    req.EndDate.UTC(),
			TotalPrice: finalPrice,
			Status:     entity.RentalStatusPending,
		}

		if err := tx.Create(&created).Error; err != nil {
			return err
		}

		if err := tx.Model(&entity.Car{}).Where("id = ?", req.CarID).
			Update("status", entity.CarStatusBooked).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			RespondWithError(w, http.StatusNotFound, "car or user not found")
		case err.Error() == "car already booked":
			RespondWithError(w, http.StatusBadRequest, "car already booked for these dates")
		case err.Error() == "car not available":
			RespondWithError(w, http.StatusBadRequest, "car not available")
		default:
			RespondWithError(w, http.StatusInternalServerError, "could not create rental")
		}
		return
	}

	RespondWithJSON(w, http.StatusCreated, map[string]any{
		"rental_id":   created.ID,
		"total_price": created.TotalPrice,
		"status":      created.Status,
		"message":     "Rental created. Please proceed to payment.",
	})
}

func (s *Server) listRentals(w http.ResponseWriter, r *http.Request) {
	userID, ok := authhttp.UserIDFromContext(r.Context())
	if !ok {
		RespondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	role := getRoleFromContext(r)
	q := s.db.Model(&entity.Rental{})
	if role != entity.UserRoleAdmin {
		q = q.Where("user_id = ?", userID)
	} else if v := r.URL.Query().Get("user_id"); v != "" {
		id, err := strconv.Atoi(v)
		if err != nil || id <= 0 {
			RespondWithError(w, http.StatusBadRequest, "invalid user_id")
			return
		}
		q = q.Where("user_id = ?", id)
	}

	var rentals []entity.Rental
	if err := q.Find(&rentals).Error; err != nil {
		RespondWithError(w, http.StatusInternalServerError, "database error")
		return
	}

	RespondWithJSON(w, http.StatusOK, rentals)
}

func (s *Server) payRental(w http.ResponseWriter, r *http.Request, rentalID uint) {
	userID, ok := authhttp.UserIDFromContext(r.Context())
	if !ok {
		RespondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	role := getRoleFromContext(r)

	err := s.db.Transaction(func(tx *gorm.DB) error {
		var rental entity.Rental
		if err := tx.First(&rental, rentalID).Error; err != nil {
			return err
		}

		if role != entity.UserRoleAdmin && rental.UserID != userID {
			return errors.New("forbidden")
		}
		if rental.Status != entity.RentalStatusPending {
			return errors.New("invalid status")
		}

		res := tx.Model(&entity.User{}).
			Where("id = ? AND balance >= ?", rental.UserID, rental.TotalPrice).
			Update("balance", gorm.Expr("balance - ?", rental.TotalPrice))
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return errors.New("insufficient balance")
		}

		if err := tx.Model(&entity.Rental{}).
			Where("id = ? AND status = ?", rentalID, entity.RentalStatusPending).
			Update("status", entity.RentalStatusActive).Error; err != nil {
			return err
		}

			transaction := entity.Transaction{
				UserID:   rental.UserID,
				RentalID: &rentalID,
				Type:     "payment",
				Amount:   rental.TotalPrice,
				Status:   entity.TransactionStatusSuccess,
			}
			if err := tx.Create(&transaction).Error; err != nil {
				return err
			}

		return nil
	})

	if err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			RespondWithError(w, http.StatusNotFound, "rental not found")
		case err.Error() == "forbidden":
			RespondWithError(w, http.StatusForbidden, "forbidden")
		case err.Error() == "invalid status":
			RespondWithError(w, http.StatusBadRequest, "rental is not pending")
		case err.Error() == "insufficient balance":
			RespondWithError(w, http.StatusBadRequest, "insufficient balance")
		default:
			RespondWithError(w, http.StatusInternalServerError, "payment failed")
		}
		return
	}

	RespondWithJSON(w, http.StatusOK, map[string]any{"message": "payment successful"})
}

func (s *Server) finishRental(w http.ResponseWriter, r *http.Request, rentalID uint) {
	userID, ok := authhttp.UserIDFromContext(r.Context())
	if !ok {
		RespondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	role := getRoleFromContext(r)

	err := s.db.Transaction(func(tx *gorm.DB) error {
		var rental entity.Rental
		if err := tx.First(&rental, rentalID).Error; err != nil {
			return err
		}
		if role != entity.UserRoleAdmin && rental.UserID != userID {
			return errors.New("forbidden")
		}
		if rental.Status != entity.RentalStatusActive {
			return errors.New("invalid status")
		}

		if err := tx.Model(&entity.Rental{}).Where("id = ?", rentalID).
			Update("status", entity.RentalStatusCompleted).Error; err != nil {
			return err
		}
		if err := tx.Model(&entity.Car{}).Where("id = ?", rental.CarID).
			Update("status", entity.CarStatusAvailable).Error; err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			RespondWithError(w, http.StatusNotFound, "rental not found")
		case err.Error() == "forbidden":
			RespondWithError(w, http.StatusForbidden, "forbidden")
		case err.Error() == "invalid status":
			RespondWithError(w, http.StatusBadRequest, "rental is not active")
		default:
			RespondWithError(w, http.StatusInternalServerError, "could not finish rental")
		}
		return
	}

	RespondWithJSON(w, http.StatusOK, map[string]any{"message": "rental completed"})
}

func (s *Server) cancelRental(w http.ResponseWriter, r *http.Request, rentalID uint) {
	userID, ok := authhttp.UserIDFromContext(r.Context())
	if !ok {
		RespondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	role := getRoleFromContext(r)

	err := s.db.Transaction(func(tx *gorm.DB) error {
		var rental entity.Rental
		if err := tx.First(&rental, rentalID).Error; err != nil {
			return err
		}
		if role != entity.UserRoleAdmin && rental.UserID != userID {
			return errors.New("forbidden")
		}
		if rental.Status != entity.RentalStatusPending {
			return errors.New("invalid status")
		}
		if time.Now().UTC().After(rental.StartDate) {
			return errors.New("too late")
		}

		if err := tx.Model(&entity.Rental{}).Where("id = ?", rentalID).
			Update("status", entity.RentalStatusCancelled).Error; err != nil {
			return err
		}
		if err := tx.Model(&entity.Car{}).Where("id = ?", rental.CarID).
			Update("status", entity.CarStatusAvailable).Error; err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			RespondWithError(w, http.StatusNotFound, "rental not found")
		case err.Error() == "forbidden":
			RespondWithError(w, http.StatusForbidden, "forbidden")
		case err.Error() == "invalid status":
			RespondWithError(w, http.StatusBadRequest, "rental is not pending")
		case err.Error() == "too late":
			RespondWithError(w, http.StatusBadRequest, "cannot cancel after start")
		default:
			RespondWithError(w, http.StatusInternalServerError, "could not cancel rental")
		}
		return
	}

	RespondWithJSON(w, http.StatusOK, map[string]any{"message": "rental cancelled"})
}

func parseRentalAction(path string) (uint, string, error) {
	trimmed := strings.TrimPrefix(path, "/api/v1/rentals/")
	parts := strings.Split(strings.Trim(trimmed, "/"), "/")
	if len(parts) != 2 {
		return 0, "", errors.New("invalid path")
	}
	id, err := strconv.Atoi(parts[0])
	if err != nil || id <= 0 {
		return 0, "", errors.New("invalid id")
	}
	return uint(id), parts[1], nil
}
