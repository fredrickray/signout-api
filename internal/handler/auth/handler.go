package auth

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/signout/signout-api/internal/domain"
	"github.com/signout/signout-api/internal/middleware"
	authsvc "github.com/signout/signout-api/internal/service/auth"
	"github.com/signout/signout-api/pkg/response"
	apivalidator "github.com/signout/signout-api/pkg/validator"
)

type Handler struct {
	svc *authsvc.Service
}

func NewHandler(svc *authsvc.Service) *Handler {
	return &Handler{svc: svc}
}

type registerRequest struct {
	Email    string `json:"email" validate:"required,email,max=255"`
	Password string `json:"password" validate:"required,min=8,max=128"`
	FullName string `json:"full_name" validate:"required,min=2,max=120"`
}

type loginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type logoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	if details := apivalidator.Struct(req); details != nil {
		response.FailDetails(w, http.StatusUnprocessableEntity, "validation_error", "invalid request", details)
		return
	}

	result, err := h.svc.Register(r.Context(), authsvc.RegisterInput{
		Email:    req.Email,
		Password: req.Password,
		FullName: req.FullName,
	})
	if err != nil {
		mapAuthError(w, err)
		return
	}
	response.JSON(w, http.StatusCreated, result)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	if details := apivalidator.Struct(req); details != nil {
		response.FailDetails(w, http.StatusUnprocessableEntity, "validation_error", "invalid request", details)
		return
	}

	result, err := h.svc.Login(r.Context(), authsvc.LoginInput{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		mapAuthError(w, err)
		return
	}
	response.JSON(w, http.StatusOK, result)
}

func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req refreshRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	if details := apivalidator.Struct(req); details != nil {
		response.FailDetails(w, http.StatusUnprocessableEntity, "validation_error", "invalid request", details)
		return
	}

	tokens, err := h.svc.Refresh(r.Context(), req.RefreshToken)
	if err != nil {
		mapAuthError(w, err)
		return
	}
	response.JSON(w, http.StatusOK, tokens)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	var req logoutRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	if err := h.svc.Logout(r.Context(), req.RefreshToken); err != nil {
		mapAuthError(w, err)
		return
	}
	response.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.Fail(w, http.StatusUnauthorized, "unauthorized", "missing user context")
		return
	}
	user, err := h.svc.Me(r.Context(), userID)
	if err != nil {
		mapAuthError(w, err)
		return
	}
	response.JSON(w, http.StatusOK, user)
}

func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		response.Fail(w, http.StatusBadRequest, "bad_request", "invalid JSON body")
		return false
	}
	return true
}

func mapAuthError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrConflict):
		response.Fail(w, http.StatusConflict, "conflict", "email already registered")
	case errors.Is(err, domain.ErrInvalidCredentials):
		response.Fail(w, http.StatusUnauthorized, "invalid_credentials", "email or password is incorrect")
	case errors.Is(err, domain.ErrRefreshTokenInvalid):
		response.Fail(w, http.StatusUnauthorized, "invalid_refresh_token", "refresh token is invalid or expired")
	case errors.Is(err, domain.ErrNotFound):
		response.Fail(w, http.StatusNotFound, "not_found", "resource not found")
	case errors.Is(err, domain.ErrUnauthorized):
		response.Fail(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
	default:
		response.Fail(w, http.StatusInternalServerError, "internal_error", "unexpected server error")
	}
}
