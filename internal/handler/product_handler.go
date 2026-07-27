package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/Timeobi/go-ecommerce/internal/service"
	"github.com/go-chi/chi/v5"
)

type ProductHandler struct {
	service *service.ProductService
}

func NewProductHandler(service *service.ProductService) *ProductHandler {
	return &ProductHandler{service: service}
}

type createProductRequest struct {
	CategoryID    *int   `json:"category_id"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	Price         int    `json:"price"`
	StockQuantity int    `json:"stock_quantity"`
}

func (h *ProductHandler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	var req createProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "невалидный JSON в теле запроса")
		return
	}

	product, err := h.service.CreateProduct(r.Context(), service.CreateProductInput{
		CategoryID:    req.CategoryID,
		Name:          req.Name,
		Description:   req.Description,
		Price:         req.Price,
		StockQuantity: req.StockQuantity,
	})
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, product)
}

// listProductsResponse — обёртка ответа со списком и метаданными пагинации.
type listProductsResponse struct {
	Products interface{} `json:"products"`
	Total    int         `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
}

func (h *ProductHandler) GetAllProducts(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	input := service.ListProductsInput{
		Search:   query.Get("search"),
		Page:     parseIntOrDefault(query.Get("page"), 1),
		PageSize: parseIntOrDefault(query.Get("page_size"), 0),
	}

	if categoryIDStr := query.Get("category_id"); categoryIDStr != "" {
		if categoryID, err := strconv.Atoi(categoryIDStr); err == nil {
			input.CategoryID = &categoryID
		}
	}

	result, err := h.service.ListProducts(r.Context(), input)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "не удалось получить список товаров")
		return
	}

	respondJSON(w, http.StatusOK, listProductsResponse{
		Products: result.Products,
		Total:    result.Total,
		Page:     result.Page,
		PageSize: result.PageSize,
	})
}

func (h *ProductHandler) GetProductByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "некорректный ID товара")
		return
	}

	product, err := h.service.GetProductByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrProductNotFound) {
			respondError(w, http.StatusNotFound, "товар не найден")
			return
		}
		respondError(w, http.StatusInternalServerError, "внутренняя ошибка сервера")
		return
	}

	respondJSON(w, http.StatusOK, product)
}

type updateProductRequest struct {
	Name          string `json:"name"`
	Description   string `json:"description"`
	Price         int    `json:"price"`
	StockQuantity int    `json:"stock_quantity"`
}

func (h *ProductHandler) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "некорректный ID товара")
		return
	}

	var req updateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "невалидный JSON в теле запроса")
		return
	}

	product, err := h.service.UpdateProduct(r.Context(), id, service.UpdateProductInput{
		Name:          req.Name,
		Description:   req.Description,
		Price:         req.Price,
		StockQuantity: req.StockQuantity,
	})
	if err != nil {
		if errors.Is(err, service.ErrProductNotFound) {
			respondError(w, http.StatusNotFound, "товар не найден")
			return
		}
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, product)
}

func (h *ProductHandler) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "некорректный ID товара")
		return
	}

	if err := h.service.DeleteProduct(r.Context(), id); err != nil {
		if errors.Is(err, service.ErrProductNotFound) {
			respondError(w, http.StatusNotFound, "товар не найден")
			return
		}
		respondError(w, http.StatusInternalServerError, "внутренняя ошибка сервера")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// parseIntOrDefault пытается распарсить строку в int, при ошибке возвращает значение по умолчанию.
func parseIntOrDefault(s string, def int) int {
	if s == "" {
		return def
	}
	val, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return val
}
