package authhttp

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-playground/validator/v10"

	"github.com/CMPNION/Car-Rental-API.git/internal/entity"
	authuc "github.com/CMPNION/Car-Rental-API.git/internal/usecase/auth"
)

type Handler struct {
	svc *authuc.AuthService
}

func NewHandler(svc *authuc.AuthService) *Handler {
	return &Handler{svc: svc}
}

func RegisterHandlers(mux *http.ServeMux, svc *authuc.AuthService, jwtSecret string) {
	h := NewHandler(svc)
	mux.HandleFunc("/auth/register", h.register)
	mux.HandleFunc("/auth/login", h.login)
	mux.Handle("/auth/me", JWTAuthMiddleware(jwtSecret)(http.HandlerFunc(h.me)))
}

func (h *Handler) register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req authuc.RegisterRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.svc.Register(req)
	if err != nil {
		switch {
		case errors.Is(err, authuc.ErrEmailTaken):
			writeError(w, http.StatusBadRequest, "email already taken")
		case isValidationError(err):
			writeError(w, http.StatusBadRequest, err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	writeJSON(w, http.StatusCreated, resp)
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req authuc.LoginRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.svc.Login(req)
	if err != nil {
		switch {
		case errors.Is(err, authuc.ErrInvalidCredentials):
			writeError(w, http.StatusUnauthorized, "invalid credentials")
		case isValidationError(err):
			writeError(w, http.StatusBadRequest, err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) me(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	role, ok := RoleFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"role":     role,
		"is_admin": role == entity.UserRoleAdmin,
	})
}

func decodeJSON(r *http.Request, dst any) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(dst)
}

type apiResponse struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(apiResponse{Status: "ok", Data: payload})
}

func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(apiResponse{Status: "error", Message: message})
}

func isValidationError(err error) bool {
	var validationErrors validator.ValidationErrors
	return errors.As(err, &validationErrors)
}
