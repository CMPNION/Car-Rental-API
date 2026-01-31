package server

import (
	"net/http"
	"strconv"

	"github.com/CMPNION/Car-Rental-API.git/internal/auth/middleware"
	"github.com/CMPNION/Car-Rental-API.git/internal/models"
)

func (s *Server) transactionsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		RespondWithError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		RespondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	role := getRoleFromContext(r)
	q := s.db.Model(&models.Transaction{})
	if role != models.UserRoleAdmin {
		q = q.Where("user_id = ?", userID)
	} else if v := r.URL.Query().Get("user_id"); v != "" {
		id, err := strconv.Atoi(v)
		if err != nil || id <= 0 {
			RespondWithError(w, http.StatusBadRequest, "invalid user_id")
			return
		}
		q = q.Where("user_id = ?", id)
	}

	var transactions []models.Transaction
	if err := q.Order("created_at desc").Find(&transactions).Error; err != nil {
		RespondWithError(w, http.StatusInternalServerError, "database error")
		return
	}

	RespondWithJSON(w, http.StatusOK, transactions)
}
