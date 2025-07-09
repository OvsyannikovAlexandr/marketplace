package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/OvsyannikovAlexandr/marketplace/product-service/internal/domain"
	"github.com/OvsyannikovAlexandr/marketplace/product-service/internal/service"
	"github.com/gorilla/mux"
)

type ProductHandler struct {
	service service.ProductServiceInterface
}

func NewProductHandler(service service.ProductServiceInterface) *ProductHandler {
	return &ProductHandler{service: service}
}

// @Summary      Создать продукт
// @Description  Добавляет новый продукт в базу
// @Tags         products
// @Accept       json
// @Produce      json
// @Param        product  body      domain.Product  true  "Продукт"
// @Success      201
// @Failure      400  {string}  string "invalid body"
// @Failure      500  {string}  string "internal error"
// @Router       /products [post]
func (h *ProductHandler) Create(w http.ResponseWriter, r *http.Request) {
	var p domain.Product
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	if err := h.service.Create(r.Context(), p); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

// @Summary      Получить все продукты
// @Description  Получает все продукты из базы
// @Tags         products
// @Produce      json
// @Success      200  {array}   domain.Product
// @Failure      500  {string}  string "internal error"
// @Router       /products [get]
func (h *ProductHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	products, err := h.service.GetAll(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(products)
}

// @Summary      Получить продукт по ID
// @Description  Получает продукт по ID
// @Tags         products
// @Produce      json
// @Param        id   path      int  true  "ID продукта"
// @Success      200  {object}  domain.Product
// @Failure      400  {string}  string "invalid ID"
// @Failure      404  {string}  string "not found"
// @Router       /products/{id} [get]
func (h *ProductHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid ID", http.StatusBadRequest)
		return
	}

	product, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(product)
}

// @Summary      Удаление продукта
// @Description  Удаляет продукт по ID
// @Tags         products
// @Param        id   path      int  true  "ID продукта"
// @Success      204
// @Failure      400  {string}  string "invalid ID"
// @Failure      500  {string}  string "internal error"
// @Router       /products/{id} [delete]
func (h *ProductHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid ID", http.StatusBadRequest)
		return
	}
	if err := h.service.Delete(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
