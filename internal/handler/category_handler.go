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
//
// @Summary      Создать категорию
// @Description  Создаёт новую категорию товаров. Требует роль admin.
// @Tags         categories
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        input  body      createCategoryRequest  true  "Данные категории"
// @Success      201    {object}  model.Category
// @Failure      400    {object}  errorResponse
// @Failure      401    {object}  errorResponse
// @Failure      403    {object}  errorResponse
// @Router       /categories [post]
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
//
// @Summary      Список категорий
// @Description  Возвращает все категории товаров. Публичный эндпоинт.
// @Tags         categories
// @Produce      json
// @Success      200  {array}   model.Category
// @Router       /categories [get]
func (h *CategoryHandler) GetAllCategories(w http.ResponseWriter, r *http.Request) {

	categories, err := h.service.GetAllCategories(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, "не удалось получить список категорий")
		return
	}

	respondJSON(w, http.StatusOK, categories)
}

// GetCategoryByID обрабатывает GET /categories/{id}
//
// @Summary      Получить категорию по ID
// @Description  Возвращает одну категорию. Публичный эндпоинт.
// @Tags         categories
// @Produce      json
// @Param        id   path      int  true  "ID категории"
// @Success      200  {object}  model.Category
// @Failure      404  {object}  errorResponse
// @Router       /categories/{id} [get]
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
