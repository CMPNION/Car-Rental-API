package server

import (
	"net/http"

	"github.com/CMPNION/Car-Rental-API.git/internal/models"
)

func (s *Server) adminMetricsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		RespondWithError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	if getRoleFromContext(r) != models.UserRoleAdmin {
		RespondWithError(w, http.StatusForbidden, "admin only")
		return
	}

	var totalRevenue float64
	_ = s.db.Model(&models.Transaction{}).
		Where("type = ? AND status = ?", "payment", models.TransactionStatusSuccess).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&totalRevenue).Error

	var activeRentals int64
	_ = s.db.Model(&models.Rental{}).
		Where("status = ?", models.RentalStatusActive).
		Count(&activeRentals).Error

	var totalCars int64
	_ = s.db.Model(&models.Car{}).Count(&totalCars).Error
	var bookedCars int64
	_ = s.db.Model(&models.Car{}).Where("status = ?", models.CarStatusBooked).Count(&bookedCars).Error

	fleetLoad := 0.0
	if totalCars > 0 {
		fleetLoad = (float64(bookedCars) / float64(totalCars)) * 100
	}

	RespondWithJSON(w, http.StatusOK, map[string]any{
		"total_revenue": totalRevenue,
		"active_rentals": activeRentals,
		"fleet_load":    fleetLoad,
		"total_cars":    totalCars,
	})
}
