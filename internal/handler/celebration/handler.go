package celebration

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/signout/signout-api/internal/domain"
	"github.com/signout/signout-api/internal/middleware"
	celebSvc "github.com/signout/signout-api/internal/service/celebration"
	"github.com/signout/signout-api/pkg/response"
	apivalidator "github.com/signout/signout-api/pkg/validator"
)

type Handler struct {
	svc *celebSvc.Service
}

func NewHandler(svc *celebSvc.Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) ListShirts(w http.ResponseWriter, r *http.Request) {
	shirts, err := h.svc.ListShirts(r.Context())
	if err != nil {
		mapError(w, err)
		return
	}
	if shirts == nil {
		shirts = []domain.ShirtTemplate{}
	}
	response.JSON(w, http.StatusOK, shirts)
}

type createRequest struct {
	Slug            string `json:"slug" validate:"required,min=3,max=48"`
	DisplayName     string `json:"display_name" validate:"required,min=2,max=120"`
	School          string `json:"school" validate:"required,min=2,max=160"`
	Faculty         string `json:"faculty" validate:"required,min=2,max=120"`
	ClassOf         string `json:"class_of" validate:"required,min=2,max=64"`
	CelebrationType string `json:"celebration_type" validate:"required"`
	ShirtTemplateID string `json:"shirt_template_id" validate:"required,uuid"`
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.Fail(w, http.StatusUnauthorized, "unauthorized", "missing user context")
		return
	}

	var req createRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	if details := apivalidator.Struct(req); details != nil {
		response.FailDetails(w, http.StatusUnprocessableEntity, "validation_error", "invalid request", details)
		return
	}

	shirtID, err := uuid.Parse(req.ShirtTemplateID)
	if err != nil {
		response.Fail(w, http.StatusBadRequest, "bad_request", "invalid shirt_template_id")
		return
	}

	result, err := h.svc.Create(r.Context(), celebSvc.CreateInput{
		UserID:          userID,
		Slug:            req.Slug,
		DisplayName:     req.DisplayName,
		School:          req.School,
		Faculty:         req.Faculty,
		ClassOf:         req.ClassOf,
		CelebrationType: req.CelebrationType,
		ShirtTemplateID: shirtID,
	})
	if err != nil {
		mapError(w, err)
		return
	}
	response.JSON(w, http.StatusCreated, result)
}

func (h *Handler) GetBySlug(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	result, err := h.svc.GetBySlug(r.Context(), slug)
	if err != nil {
		mapError(w, err)
		return
	}
	response.JSON(w, http.StatusOK, result)
}

func (h *Handler) ListMine(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.Fail(w, http.StatusUnauthorized, "unauthorized", "missing user context")
		return
	}
	list, err := h.svc.ListMine(r.Context(), userID)
	if err != nil {
		mapError(w, err)
		return
	}
	if list == nil {
		list = []domain.CelebrationPublic{}
	}
	response.JSON(w, http.StatusOK, list)
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

func mapError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		response.Fail(w, http.StatusNotFound, "not_found", "resource not found")
	case errors.Is(err, domain.ErrConflict):
		response.Fail(w, http.StatusConflict, "conflict", "slug already taken")
	case errors.Is(err, domain.ErrValidation):
		response.Fail(w, http.StatusUnprocessableEntity, "validation_error", err.Error())
	default:
		response.Fail(w, http.StatusInternalServerError, "internal_error", "unexpected server error")
	}
}
