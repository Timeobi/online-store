package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/Timeobi/go-ecommerce/internal/service"
	"github.com/go-chi/chi/v5"
)

type CategoryHandler struct {
	service *service.CategoryService
}

func NewCategoryHandler(service *service.CategoryService) *CategoryHandler {
	return &CategoryHandler{service: service}
}

type createCategoryRequest struct {
	Name string `json:"name"`
}

// CreateCategory обрабатывает POST /categories
func (h *CategoryHandler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	var req createCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "невалидный JSON в теле запроса")
		return
	}

	category, err := h.service.CreateCategory(r.Context(), req.Name)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, category)
}

// GetAllCategories обрабатывает GET /categories
func (h *CategoryHandler) GetAllCategories(w http.ResponseWriter, r *http.Request) {

	categories, err := h.service.GetAllCategories(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, "не удалось получить список категорий")
		return
	}

	respondJSON(w, http.StatusOK, categories)
}

// GetCategoryByID обрабатывает GET /categories/{id}
func (h *CategoryHandler) GetCategoryByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "некорректный ID категории")
		return
	}

	category, err := h.service.GetCategoryByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrCategoryNotFound) {
			respondError(w, http.StatusNotFound, "категория не найдена")
			return
		}
		respondError(w, http.StatusInternalServerError, "внутренняя ошибка сервера")
		return
	}

	respondJSON(w, http.StatusOK, category)
}
