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

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func httpErr(w http.ResponseWriter, status int, msg string) {
	http.Error(w, msg, status)
}

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
				httpErr(w, http.StatusForbidden, "admin only")
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
				httpErr(w, 400, "invalid min_price")
				return
			}
			q = q.Where("price_per_hour >= ?", p)
		}
		if v := r.URL.Query().Get("max_price"); v != "" {
			p, err := strconv.ParseFloat(v, 64)
			if err != nil {
				httpErr(w, 400, "invalid max_price")
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
				httpErr(w, 400, "invalid sort field")
				return
			}
		}

		if v := r.URL.Query().Get("limit"); v != "" {
			lim, err := strconv.Atoi(v)
			if err != nil || lim <= 0 || lim > 200 {
				httpErr(w, 400, "invalid limit")
				return
			}
			q = q.Limit(lim)
		}
		if v := r.URL.Query().Get("offset"); v != "" {
			off, err := strconv.Atoi(v)
			if err != nil || off < 0 {
				httpErr(w, 400, "invalid offset")
				return
			}
			q = q.Offset(off)
		}

		var cars []models.Car
		if err := q.Find(&cars).Error; err != nil {
			httpErr(w, 500, err.Error())
			return
		}
		writeJSON(w, 200, cars)

	case http.MethodPost:
		s.adminOnly(func(w http.ResponseWriter, r *http.Request) {
			var payload models.Car
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				httpErr(w, 400, "invalid JSON")
				return
			}

			payload.Mark = strings.TrimSpace(payload.Mark)
			payload.CarModel = strings.TrimSpace(payload.CarModel)
			payload.Category = strings.ToLower(strings.TrimSpace(payload.Category))
			payload.Status = strings.ToLower(strings.TrimSpace(payload.Status))

			if payload.Mark == "" || payload.CarModel == "" {
				httpErr(w, 400, "mark and model are required")
				return
			}
			if payload.PricePerHour <= 0 {
				httpErr(w, 400, "price_per_hour must be > 0")
				return
			}

			if payload.Category != models.CarCategoryEconomy &&
				payload.Category != models.CarCategoryBusiness &&
				payload.Category != models.CarCategoryLuxury {
				httpErr(w, 400, "invalid category")
				return
			}

			if payload.Status == "" {
				payload.Status = models.CarStatusAvailable
			}
			if payload.Status != models.CarStatusAvailable &&
				payload.Status != models.CarStatusBooked &&
				payload.Status != models.CarStatusMaintenance {
				httpErr(w, 400, "invalid status")
				return
			}

			if err := s.db.Create(&payload).Error; err != nil {
				httpErr(w, 500, err.Error())
				return
			}
			writeJSON(w, 201, payload)
		})(w, r)

	default:
		httpErr(w, 405, "method not allowed")
	}
}

func (s *Server) carByIDHandler(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/v1/cars/")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		httpErr(w, 400, "invalid id")
		return
	}

	switch r.Method {

	case http.MethodGet:
		var car models.Car
		if err := s.db.First(&car, id).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				httpErr(w, 404, "car not found")
				return
			}
			httpErr(w, 500, err.Error())
			return
		}
		writeJSON(w, 200, car)

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
				httpErr(w, 400, "invalid JSON")
				return
			}

			updates := map[string]any{}

			if payload.Mark != nil {
				v := strings.TrimSpace(*payload.Mark)
				if v == "" {
					httpErr(w, 400, "mark cannot be empty")
					return
				}
				updates["mark"] = v
			}
			if payload.CarModel != nil {
				v := strings.TrimSpace(*payload.CarModel)
				if v == "" {
					httpErr(w, 400, "model cannot be empty")
					return
				}
				updates["model"] = v
			}
			if payload.Category != nil {
				v := strings.ToLower(strings.TrimSpace(*payload.Category))
				if v != models.CarCategoryEconomy && v != models.CarCategoryBusiness && v != models.CarCategoryLuxury {
					httpErr(w, 400, "invalid category")
					return
				}
				updates["category"] = v
			}
			if payload.Status != nil {
				v := strings.ToLower(strings.TrimSpace(*payload.Status))
				if v != models.CarStatusAvailable && v != models.CarStatusBooked && v != models.CarStatusMaintenance {
					httpErr(w, 400, "invalid status")
					return
				}
				updates["status"] = v
			}
			if payload.PricePerHour != nil {
				if *payload.PricePerHour <= 0 {
					httpErr(w, 400, "price_per_hour must be > 0")
					return
				}
				updates["price_per_hour"] = *payload.PricePerHour
			}
			if payload.Metadata != nil {
				updates["metadata"] = *payload.Metadata
			}

			if len(updates) == 0 {
				httpErr(w, 400, "no fields to update")
				return
			}

			res := s.db.Model(&models.Car{}).Where("id = ?", id).Updates(updates)
			if res.Error != nil {
				httpErr(w, 500, res.Error.Error())
				return
			}
			if res.RowsAffected == 0 {
				httpErr(w, 404, "car not found")
				return
			}

			var updated models.Car
			_ = s.db.First(&updated, id).Error
			writeJSON(w, 200, updated)
		})(w, r)

	case http.MethodDelete:
		s.adminOnly(func(w http.ResponseWriter, r *http.Request) {
			res := s.db.Delete(&models.Car{}, id)
			if res.Error != nil {
				httpErr(w, 500, res.Error.Error())
				return
			}
			if res.RowsAffected == 0 {
				httpErr(w, 404, "car not found")
				return
			}
			w.WriteHeader(http.StatusNoContent)
		})(w, r)

	default:
		httpErr(w, 405, "method not allowed")
	}
}
