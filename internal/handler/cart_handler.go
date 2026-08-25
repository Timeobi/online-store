package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	authmw "github.com/Timeobi/go-ecommerce/internal/middleware"
	"github.com/Timeobi/go-ecommerce/internal/service"
)

type CartHandler struct {
	service *service.CartService
}

func NewCartHandler(service *service.CartService) *CartHandler {
	return &CartHandler{service: service}
}

// getUserID достаёт ID текущего пользователя из контекста запроса — его туда
// кладёт middleware Auth (см. Этап 5) после успешной проверки JWT-токена.
// Функция общая для всех "защищённых" хендлеров, объявлена один раз здесь.
func getUserID(r *http.Request) (int, bool) {
	userID, ok := r.Context().Value(authmw.UserIDContextKey).(int)
	return userID, ok
}

func (h *CartHandler) GetCart(w http.ResponseWriter, r *http.Request) {
	userID, ok := getUserID(r)
	if !ok {
		respondError(w, http.StatusUnauthorized, "не удалось определить пользователя")
		return
	}

	cart, err := h.service.GetCart(r.Context(), userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "не удалось получить корзину")
		return
	}

	respondJSON(w, http.StatusOK, cart)
}

type addCartItemRequest struct {
	ProductID int `json:"product_id"`
	Quantity  int `json:"quantity"`
}

func (h *CartHandler) AddItem(w http.ResponseWriter, r *http.Request) {
	userID, ok := getUserID(r)
	if !ok {
		respondError(w, http.StatusUnauthorized, "не удалось определить пользователя")
		return
	}

	var req addCartItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "невалидный JSON в теле запроса")
		return
	}

	if err := h.service.AddItem(r.Context(), userID, req.ProductID, req.Quantity); err != nil {
		if errors.Is(err, service.ErrProductNotFound) {
			respondError(w, http.StatusNotFound, "товар не найден")
			return
		}
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

type updateCartItemRequest struct {
	Quantity int `json:"quantity"`
}

func (h *CartHandler) UpdateItemQuantity(w http.ResponseWriter, r *http.Request) {
	userID, ok := getUserID(r)
	if !ok {
		respondError(w, http.StatusUnauthorized, "не удалось определить пользователя")
		return
	}

	productID, err := strconv.Atoi(chi.URLParam(r, "productID"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "некорректный ID товара")
		return
	}

	var req updateCartItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "невалидный JSON в теле запроса")
		return
	}

	if err := h.service.UpdateItemQuantity(r.Context(), userID, productID, req.Quantity); err != nil {
		if errors.Is(err, service.ErrProductNotInCart) {
			respondError(w, http.StatusNotFound, err.Error())
			return
		}
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *CartHandler) RemoveItem(w http.ResponseWriter, r *http.Request) {
	userID, ok := getUserID(r)
	if !ok {
		respondError(w, http.StatusUnauthorized, "не удалось определить пользователя")
		return
	}

	productID, err := strconv.Atoi(chi.URLParam(r, "productID"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "некорректный ID товара")
		return
	}

	if err := h.service.RemoveItem(r.Context(), userID, productID); err != nil {
		if errors.Is(err, service.ErrProductNotInCart) {
			respondError(w, http.StatusNotFound, err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, "внутренняя ошибка сервера")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
