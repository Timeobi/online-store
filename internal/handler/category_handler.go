package handler

import (
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

func (h *CategoryHandler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	// ... (код из Этапа 3, без изменений)
}

func (h *CategoryHandler) GetAllCategories(w http.ResponseWriter, r *http.Request) {
	// ... (код из Этапа 3, без изменений)
}

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
