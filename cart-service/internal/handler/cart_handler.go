package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/OvsyannikovAlexandr/marketplace/cart-service/internal/domain"
	"github.com/OvsyannikovAlexandr/marketplace/cart-service/internal/service"
	"github.com/gorilla/mux"
)

type CartHandler struct {
	svc service.CartServiceInterface
}

func NewCartHandler(svc service.CartServiceInterface) *CartHandler {
	return &CartHandler{svc: svc}
}

func (h *CartHandler) AddItem(w http.ResponseWriter, r *http.Request) {
	var item domain.CartItem
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	err := h.svc.AddItem(r.Context(), item)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (h *CartHandler) GetItems(w http.ResponseWriter, r *http.Request) {
	userIDStr := mux.Vars(r)["user_id"]
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid user_id", http.StatusBadRequest)
		return
	}
	items, err := h.svc.GetItems(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(items)
}

func (h *CartHandler) DeleteItem(w http.ResponseWriter, r *http.Request) {
	userIDStr := mux.Vars(r)["user_id"]
	productIDStr := mux.Vars(r)["product_id"]

	userID, err1 := strconv.ParseInt(userIDStr, 10, 64)
	productID, err2 := strconv.ParseInt(productIDStr, 10, 64)

	if err1 != nil || err2 != nil {
		http.Error(w, "invalid parameters", http.StatusBadRequest)
		return
	}

	if err := h.svc.DeleteItem(r.Context(), userID, productID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *CartHandler) ClearCart(w http.ResponseWriter, r *http.Request) {
	userIDStr := mux.Vars(r)["user_id"]
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid user_id", http.StatusBadRequest)
		return
	}

	if err := h.svc.ClearCart(r.Context(), userID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *CartHandler) GetCartDetailsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userIDStr := vars["user_id"]

	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}

	items, err := h.svc.GetCartWithDetails(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}
