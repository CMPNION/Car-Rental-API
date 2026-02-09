package server

import (
	"net/http"
	"time"

	"github.com/CMPNION/Car-Rental-API.git/internal/entity"
)

func (s *Server) adminMetricsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		RespondWithError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	if getRoleFromContext(r) != entity.UserRoleAdmin {
		RespondWithError(w, http.StatusForbidden, "admin only")
		return
	}

	var totalRevenue float64
	_ = s.db.Model(&entity.Transaction{}).
		Where("type = ? AND status = ?", "payment", entity.TransactionStatusSuccess).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&totalRevenue).Error

	var totalUsers int64
	_ = s.db.Model(&entity.User{}).Count(&totalUsers).Error

	var totalRentals int64
	_ = s.db.Model(&entity.Rental{}).Count(&totalRentals).Error

	var activeRentals int64
	_ = s.db.Model(&entity.Rental{}).
		Where("status = ?", entity.RentalStatusActive).
		Count(&activeRentals).Error

	var pendingRentals int64
	_ = s.db.Model(&entity.Rental{}).
		Where("status = ?", entity.RentalStatusPending).
		Count(&pendingRentals).Error

	var completedRentals int64
	_ = s.db.Model(&entity.Rental{}).
		Where("status = ?", entity.RentalStatusCompleted).
		Count(&completedRentals).Error

	var cancelledRentals int64
	_ = s.db.Model(&entity.Rental{}).
		Where("status = ?", entity.RentalStatusCancelled).
		Count(&cancelledRentals).Error

	var totalCars int64
	_ = s.db.Model(&entity.Car{}).Count(&totalCars).Error
	var bookedCars int64
	_ = s.db.Model(&entity.Car{}).Where("status = ?", entity.CarStatusBooked).Count(&bookedCars).Error

	var averageCarRating float64
	_ = s.db.Model(&entity.Car{}).
		Select("COALESCE(AVG(rating), 0)").
		Scan(&averageCarRating).Error

	var averageUserRating float64
	_ = s.db.Model(&entity.User{}).
		Select("COALESCE(AVG(rating), 0)").
		Scan(&averageUserRating).Error

	since30 := time.Now().UTC().AddDate(0, 0, -30)
	var revenueLast30 float64
	_ = s.db.Model(&entity.Transaction{}).
		Where("type = ? AND status = ? AND created_at >= ?", "payment", entity.TransactionStatusSuccess, since30).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&revenueLast30).Error

	since7 := time.Now().UTC().AddDate(0, 0, -7)
	type revenueByDay struct {
		Day     string  `json:"day"`
		Revenue float64 `json:"revenue"`
	}
	var revenueLast7 []revenueByDay
	_ = s.db.Table("transactions").
		Select("date(created_at) as day, COALESCE(SUM(amount), 0) as revenue").
		Where("type = ? AND status = ? AND created_at >= ?", "payment", entity.TransactionStatusSuccess, since7).
		Group("day").
		Order("day asc").
		Scan(&revenueLast7).Error

	type topCar struct {
		CarID   uint   `json:"car_id"`
		Mark    string `json:"mark"`
		Model   string `json:"model"`
		Rentals int64  `json:"rentals"`
	}
	var topCars []topCar
	_ = s.db.Table("rentals").
		Select("cars.id as car_id, cars.mark as mark, cars.model as model, COUNT(rentals.id) as rentals").
		Joins("JOIN cars ON cars.id = rentals.car_id").
		Where("rentals.status != ?", entity.RentalStatusCancelled).
		Group("cars.id, cars.mark, cars.model").
		Order("rentals desc").
		Limit(5).
		Scan(&topCars).Error

	type topUser struct {
		UserID uint    `json:"user_id"`
		Name   string  `json:"name"`
		Email  string  `json:"email"`
		Spend  float64 `json:"spend"`
	}
	var topUsers []topUser
	_ = s.db.Table("transactions").
		Select("users.id as user_id, users.first_name || ' ' || users.last_name as name, users.email as email, COALESCE(SUM(transactions.amount), 0) as spend").
		Joins("JOIN users ON users.id = transactions.user_id").
		Where("transactions.type = ? AND transactions.status = ?", "payment", entity.TransactionStatusSuccess).
		Group("users.id, users.first_name, users.last_name, users.email").
		Order("spend desc").
		Limit(5).
		Scan(&topUsers).Error

	fleetLoad := 0.0
	if totalCars > 0 {
		fleetLoad = (float64(bookedCars) / float64(totalCars)) * 100
	}

	RespondWithJSON(w, http.StatusOK, map[string]any{
		"total_revenue":        totalRevenue,
		"revenue_last_30_days": revenueLast30,
		"revenue_last_7_days":  revenueLast7,
		"total_users":          totalUsers,
		"total_cars":           totalCars,
		"total_rentals":        totalRentals,
		"rentals_by_status": map[string]int64{
			"pending":   pendingRentals,
			"active":    activeRentals,
			"completed": completedRentals,
			"cancelled": cancelledRentals,
		},
		"fleet_load":          fleetLoad,
		"average_car_rating":  averageCarRating,
		"average_user_rating": averageUserRating,
		"top_cars_by_rentals": topCars,
		"top_users_by_spend":  topUsers,
	})
}
