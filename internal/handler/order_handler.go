package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	authmw "github.com/Timeobi/go-ecommerce/internal/middleware"
	"github.com/Timeobi/go-ecommerce/internal/model"
	"github.com/Timeobi/go-ecommerce/internal/service"
)

type OrderHandler struct {
	service *service.OrderService
}

func NewOrderHandler(service *service.OrderService) *OrderHandler {
	return &OrderHandler{service: service}
}

func (h *OrderHandler) Checkout(w http.ResponseWriter, r *http.Request) {
	userID, ok := getUserID(r)
	if !ok {
		respondError(w, http.StatusUnauthorized, "не удалось определить пользователя")
		return
	}

	result, err := h.service.Checkout(r.Context(), userID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrEmptyCart):
			respondError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, service.ErrInsufficientStock):
			respondError(w, http.StatusConflict, err.Error())
		default:
			respondError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	respondJSON(w, http.StatusCreated, result)
}

func (h *OrderHandler) ListMyOrders(w http.ResponseWriter, r *http.Request) {
	userID, ok := getUserID(r)
	if !ok {
		respondError(w, http.StatusUnauthorized, "не удалось определить пользователя")
		return
	}

	orders, err := h.service.ListMyOrders(r.Context(), userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "не удалось получить список заказов")
		return
	}

	respondJSON(w, http.StatusOK, orders)
}

func (h *OrderHandler) GetOrderByID(w http.ResponseWriter, r *http.Request) {
	userID, ok := getUserID(r)
	if !ok {
		respondError(w, http.StatusUnauthorized, "не удалось определить пользователя")
		return
	}

	orderID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "некорректный ID заказа")
		return
	}

	role, _ := r.Context().Value(authmw.UserRoleContextKey).(model.Role)
	isAdmin := role == model.RoleAdmin

	result, err := h.service.GetOrderByID(r.Context(), orderID, userID, isAdmin)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrOrderNotFound):
			respondError(w, http.StatusNotFound, err.Error())
		case errors.Is(err, service.ErrOrderAccessDenied):
			respondError(w, http.StatusForbidden, err.Error())
		default:
			respondError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	respondJSON(w, http.StatusOK, result)
}
