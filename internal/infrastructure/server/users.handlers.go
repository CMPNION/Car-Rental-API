package server

import (
	"encoding/json"
	"errors"
	"net/http"

	"gorm.io/gorm"

	"github.com/CMPNION/Car-Rental-API.git/internal/auth/middleware"
	"github.com/CMPNION/Car-Rental-API.git/internal/models"
)

type balanceRequest struct {
	Amount float64 `json:"amount"`
}

func (s *Server) userBalanceHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		RespondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	switch r.Method {
	case http.MethodGet:
		var user models.User
		if err := s.db.First(&user, userID).Error; err != nil {
			RespondWithError(w, http.StatusInternalServerError, "database error")
			return
		}
		RespondWithJSON(w, http.StatusOK, map[string]any{
			"balance": user.Balance,
		})
		return
	case http.MethodPatch:
		// continue below
	default:
		RespondWithError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req balanceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondWithError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if req.Amount <= 0 {
		RespondWithError(w, http.StatusBadRequest, "amount must be > 0")
		return
	}

	var user models.User
	err := s.db.Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&models.User{}).Where("id = ?", userID).
			Update("balance", gorm.Expr("balance + ?", req.Amount))
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}

		if err := tx.First(&user, userID).Error; err != nil {
			return err
		}

		transaction := models.Transaction{
			UserID: userID,
			Type:   "topup",
			Amount: req.Amount,
			Status: models.TransactionStatusSuccess,
		}
		if err := tx.Create(&transaction).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			RespondWithError(w, http.StatusNotFound, "user not found")
			return
		}
		RespondWithError(w, http.StatusInternalServerError, "database error")
		return
	}

	RespondWithJSON(w, http.StatusOK, map[string]any{
		"balance": user.Balance,
	})
}
