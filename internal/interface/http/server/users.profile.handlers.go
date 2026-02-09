package server

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/CMPNION/Car-Rental-API.git/internal/entity"
	authhttp "github.com/CMPNION/Car-Rental-API.git/internal/interface/http/auth"
)

type profileUpdateRequest struct {
	FirstName *string `json:"first_name"`
	LastName  *string `json:"last_name"`
}

func (s *Server) userProfileHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := authhttp.UserIDFromContext(r.Context())
	if !ok {
		RespondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	switch r.Method {
	case http.MethodGet:
		var user entity.User
		if err := s.db.First(&user, userID).Error; err != nil {
			RespondWithError(w, http.StatusInternalServerError, "database error")
			return
		}
		RespondWithJSON(w, http.StatusOK, map[string]any{
			"id":         user.ID,
			"first_name": user.FirstName,
			"last_name":  user.LastName,
			"email":      user.Email,
			"role":       user.Role,
			"rating":     user.Rating,
			"balance":    user.Balance,
		})
		return
	case http.MethodPatch:
		// continue
	default:
		RespondWithError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req profileUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondWithError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	updates := map[string]any{}
	if req.FirstName != nil {
		v := strings.TrimSpace(*req.FirstName)
		if v == "" {
			RespondWithError(w, http.StatusBadRequest, "first_name cannot be empty")
			return
		}
		updates["first_name"] = v
	}
	if req.LastName != nil {
		v := strings.TrimSpace(*req.LastName)
		if v == "" {
			RespondWithError(w, http.StatusBadRequest, "last_name cannot be empty")
			return
		}
		updates["last_name"] = v
	}

	if len(updates) == 0 {
		RespondWithError(w, http.StatusBadRequest, "no fields to update")
		return
	}

	if err := s.db.Model(&entity.User{}).Where("id = ?", userID).Updates(updates).Error; err != nil {
		RespondWithError(w, http.StatusInternalServerError, "database error")
		return
	}

	RespondWithJSON(w, http.StatusOK, map[string]any{"message": "profile updated"})
}
