package server

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/CMPNION/Car-Rental-API.git/internal/auth/middleware"
	"github.com/CMPNION/Car-Rental-API.git/internal/models"
	"gorm.io/gorm"
)

func getRoleFromContext(r *http.Request) string {
	role, ok := middleware.RoleFromContext(r.Context())
	if !ok {
		return ""
	}
	return role
}

func (s *Server) adminOnly(next http.HandlerFunc) http.HandlerFunc {
	jwtMw := middleware.JWTAuthMiddleware(s.jwtSecret)

	return func(w http.ResponseWriter, r *http.Request) {
		jwtMw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if getRoleFromContext(r) != models.UserRoleAdmin {
				RespondWithError(w, http.StatusForbidden, "admin only")
				return
			}
			next(w, r)
		})).ServeHTTP(w, r)
	}
}

func (s *Server) carsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {

	case http.MethodGet:
		q := s.db.Model(&models.Car{})

		if v := r.URL.Query().Get("mark"); v != "" {
			q = q.Where("mark = ?", v)
		}
		if v := r.URL.Query().Get("category"); v != "" {
			q = q.Where("category = ?", strings.ToLower(v))
		}
		if v := r.URL.Query().Get("status"); v != "" {
			q = q.Where("status = ?", strings.ToLower(v))
		}
		if v := r.URL.Query().Get("min_price"); v != "" {
			p, err := strconv.ParseFloat(v, 64)
			if err != nil {
				RespondWithError(w, http.StatusBadRequest, "invalid min_price")
				return
			}
			q = q.Where("price_per_hour >= ?", p)
		}
		if v := r.URL.Query().Get("max_price"); v != "" {
			p, err := strconv.ParseFloat(v, 64)
			if err != nil {
				RespondWithError(w, http.StatusBadRequest, "invalid max_price")
				return
			}
			q = q.Where("price_per_hour <= ?", p)
		}

		sort := r.URL.Query().Get("sort")
		order := strings.ToLower(r.URL.Query().Get("order"))
		if order != "desc" {
			order = "asc"
		}
		if sort != "" {
			switch sort {
			case "price_per_hour", "rating", "created_at":
				q = q.Order(sort + " " + order)
			default:
				RespondWithError(w, http.StatusBadRequest, "invalid sort field")
				return
			}
		}

		if v := r.URL.Query().Get("limit"); v != "" {
			lim, err := strconv.Atoi(v)
			if err != nil || lim <= 0 || lim > 200 {
				RespondWithError(w, http.StatusBadRequest, "invalid limit")
				return
			}
			q = q.Limit(lim)
		}
		if v := r.URL.Query().Get("offset"); v != "" {
			off, err := strconv.Atoi(v)
			if err != nil || off < 0 {
				RespondWithError(w, http.StatusBadRequest, "invalid offset")
				return
			}
			q = q.Offset(off)
		}

		var cars []models.Car
		if err := q.Find(&cars).Error; err != nil {
			RespondWithError(w, http.StatusInternalServerError, "database error")
			return
		}
		RespondWithJSON(w, http.StatusOK, cars)

	case http.MethodPost:
		s.adminOnly(func(w http.ResponseWriter, r *http.Request) {
			var payload models.Car
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				RespondWithError(w, http.StatusBadRequest, "invalid JSON")
				return
			}

			payload.Mark = strings.TrimSpace(payload.Mark)
			payload.CarModel = strings.TrimSpace(payload.CarModel)
			payload.Category = strings.ToLower(strings.TrimSpace(payload.Category))
			payload.Status = strings.ToLower(strings.TrimSpace(payload.Status))

			if payload.Mark == "" || payload.CarModel == "" {
				RespondWithError(w, http.StatusBadRequest, "mark and model are required")
				return
			}
			if payload.PricePerHour <= 0 {
				RespondWithError(w, http.StatusBadRequest, "price_per_hour must be > 0")
				return
			}

			if payload.Category != models.CarCategoryEconomy &&
				payload.Category != models.CarCategoryBusiness &&
				payload.Category != models.CarCategoryLuxury {
				RespondWithError(w, http.StatusBadRequest, "invalid category")
				return
			}

			if payload.Status == "" {
				payload.Status = models.CarStatusAvailable
			}
			if payload.Status != models.CarStatusAvailable &&
				payload.Status != models.CarStatusBooked &&
				payload.Status != models.CarStatusMaintenance {
				RespondWithError(w, http.StatusBadRequest, "invalid status")
				return
			}

			if err := s.db.Create(&payload).Error; err != nil {
				RespondWithError(w, http.StatusInternalServerError, "database error")
				return
			}
			RespondWithJSON(w, http.StatusCreated, payload)
		})(w, r)

	default:
		RespondWithError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (s *Server) carByIDHandler(w http.ResponseWriter, r *http.Request) {
	if strings.HasSuffix(r.URL.Path, "/bookings") {
		s.carBookingsHandler(w, r)
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/api/v1/cars/")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		RespondWithError(w, http.StatusBadRequest, "invalid id")
		return
	}

	switch r.Method {

	case http.MethodGet:
		var car models.Car
		if err := s.db.First(&car, id).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				RespondWithError(w, http.StatusNotFound, "car not found")
				return
			}
			RespondWithError(w, http.StatusInternalServerError, "database error")
			return
		}
		RespondWithJSON(w, http.StatusOK, car)

	case http.MethodPut:
		s.adminOnly(func(w http.ResponseWriter, r *http.Request) {
			var payload struct {
				Mark         *string  `json:"mark"`
				CarModel     *string  `json:"model"`
				Category     *string  `json:"category"`
				Status       *string  `json:"status"`
				PricePerHour *float64 `json:"price_per_hour"`
				Metadata     *string  `json:"metadata"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				RespondWithError(w, http.StatusBadRequest, "invalid JSON")
				return
			}

			updates := map[string]any{}

			if payload.Mark != nil {
				v := strings.TrimSpace(*payload.Mark)
				if v == "" {
					RespondWithError(w, http.StatusBadRequest, "mark cannot be empty")
					return
				}
				updates["mark"] = v
			}
			if payload.CarModel != nil {
				v := strings.TrimSpace(*payload.CarModel)
				if v == "" {
					RespondWithError(w, http.StatusBadRequest, "model cannot be empty")
					return
				}
				updates["model"] = v
			}
			if payload.Category != nil {
				v := strings.ToLower(strings.TrimSpace(*payload.Category))
				if v != models.CarCategoryEconomy && v != models.CarCategoryBusiness && v != models.CarCategoryLuxury {
					RespondWithError(w, http.StatusBadRequest, "invalid category")
					return
				}
				updates["category"] = v
			}
			if payload.Status != nil {
				v := strings.ToLower(strings.TrimSpace(*payload.Status))
				if v != models.CarStatusAvailable && v != models.CarStatusBooked && v != models.CarStatusMaintenance {
					RespondWithError(w, http.StatusBadRequest, "invalid status")
					return
				}
				updates["status"] = v
			}
			if payload.PricePerHour != nil {
				if *payload.PricePerHour <= 0 {
					RespondWithError(w, http.StatusBadRequest, "price_per_hour must be > 0")
					return
				}
				updates["price_per_hour"] = *payload.PricePerHour
			}
			if payload.Metadata != nil {
				updates["metadata"] = *payload.Metadata
			}

			if len(updates) == 0 {
				RespondWithError(w, http.StatusBadRequest, "no fields to update")
				return
			}

			res := s.db.Model(&models.Car{}).Where("id = ?", id).Updates(updates)
			if res.Error != nil {
				RespondWithError(w, http.StatusInternalServerError, "database error")
				return
			}
			if res.RowsAffected == 0 {
				RespondWithError(w, http.StatusNotFound, "car not found")
				return
			}

			var updated models.Car
			_ = s.db.First(&updated, id).Error
			RespondWithJSON(w, http.StatusOK, updated)
		})(w, r)

	case http.MethodDelete:
		s.adminOnly(func(w http.ResponseWriter, r *http.Request) {
			res := s.db.Delete(&models.Car{}, id)
			if res.Error != nil {
				RespondWithError(w, http.StatusInternalServerError, "database error")
				return
			}
			if res.RowsAffected == 0 {
				RespondWithError(w, http.StatusNotFound, "car not found")
				return
			}
			w.WriteHeader(http.StatusNoContent)
		})(w, r)

	default:
		RespondWithError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (s *Server) carBookingsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		RespondWithError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	trimmed := strings.TrimPrefix(r.URL.Path, "/api/v1/cars/")
	trimmed = strings.TrimSuffix(trimmed, "/bookings")
	trimmed = strings.Trim(trimmed, "/")
	id, err := strconv.Atoi(trimmed)
	if err != nil || id <= 0 {
		RespondWithError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var rentals []models.Rental
	if err := s.db.
		Where("car_id = ? AND status != ?", id, models.RentalStatusCancelled).
		Order("start_date asc").
		Find(&rentals).Error; err != nil {
		RespondWithError(w, http.StatusInternalServerError, "database error")
		return
	}

	bookings := make([]map[string]any, 0, len(rentals))
	for _, rental := range rentals {
		bookings = append(bookings, map[string]any{
			"start_date": rental.StartDate,
			"end_date":   rental.EndDate,
			"status":     rental.Status,
		})
	}

	RespondWithJSON(w, http.StatusOK, bookings)
}
