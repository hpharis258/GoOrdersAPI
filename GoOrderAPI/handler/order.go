package handler

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/hpharis258/orders-api/model"
	"github.com/hpharis258/orders-api/repository/order"
)

type Order struct {
	Repo *order.RedisRepo
}

func (o *Order) Create(w http.ResponseWriter, r *http.Request) {
	var body struct {
		CustomerID uuid.UUID        `json:"customer_id"`
		LineItems  []model.LineItem `json:"line_items"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	now := time.Now().UTC()
	order := model.Order{
		OrderID:    uint64(rand.Intn(1000000)),
		CustomerID: body.CustomerID,
		LineItems:  body.LineItems,
		CreatedAt:  &now,
	}
	err := o.Repo.Insert(r.Context(), order)
	if err != nil {
		http.Error(w, "Failed to create order", http.StatusInternalServerError)
		return
	}
	res, err := json.Marshal(order)
	if err != nil {
		http.Error(w, "Failed to marshal order", http.StatusInternalServerError)
		return
	}
	w.Write(res)
	w.WriteHeader(http.StatusCreated)

}

func (o *Order) List(w http.ResponseWriter, r *http.Request) {
	cursorStr := r.URL.Query().Get("cursor")
	if cursorStr == "" {
		cursorStr = "0"
	}
	const decimal = 10
	const bitSize = 64
	cursor, err := strconv.ParseUint(cursorStr, decimal, bitSize)
	if err != nil {
		http.Error(w, "Invalid cursor", http.StatusBadRequest)
		return
	}
	const size = 50
	res, err := o.Repo.FindAll(r.Context(), order.FindAllPage{
		Offset: cursor,
		Size:   size,
	})
	if err != nil {
		http.Error(w, "Failed to list orders", http.StatusInternalServerError)
		return
	}
	var response struct {
		Items []model.Order `json:"items"`
		Next  uint64        `json:"next,omitempty"`
	}
	response.Items = res.Orders
	response.Next = res.Cursor
	data, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "Failed to marshal orders", http.StatusInternalServerError)
		return
	}
	w.Write(data)
	w.WriteHeader(http.StatusOK)

}

func (o *Order) GetById(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	const base = 10
	const bitSize = 64
	orderID, err := strconv.ParseUint(idParam, base, bitSize)
	if err != nil {
		http.Error(w, "Invalid order ID", http.StatusBadRequest)
		return
	}
	order, err := o.Repo.GetByID(r.Context(), orderID)
	if err != nil {
		http.Error(w, "Failed to get order", http.StatusInternalServerError)
		return
	}
	if order == nil {
		http.Error(w, "Order not found", http.StatusNotFound)
		return
	}
	res, err := json.Marshal(order)
	if err != nil {
		http.Error(w, "Failed to marshal order", http.StatusInternalServerError)
		return
	}
	w.Write(res)
	w.WriteHeader(http.StatusOK)

}
func (o *Order) UpdateById(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	idParam := chi.URLParam(r, "id")
	const base = 10
	const bitSize = 64
	orderID, err := strconv.ParseUint(idParam, base, bitSize)
	if err != nil {
		http.Error(w, "Invalid order ID", http.StatusBadRequest)
		return
	}
	existingOrder, err := o.Repo.GetByID(r.Context(), orderID)
	if err != nil {
		http.Error(w, "Failed to get order", http.StatusInternalServerError)
		return
	}
	if existingOrder == nil {
		http.Error(w, "Order not found", http.StatusNotFound)
		return
	}
	const completedStatus = "completed"
	const shippedStatus = "shipped"
	now := time.Now().UTC()
	switch body.Status {
	case shippedStatus:
		if existingOrder.ShippedAt != nil {
			http.Error(w, "Order already shipped", http.StatusBadRequest)
			return
		}
		existingOrder.ShippedAt = &now
	case completedStatus:
		if existingOrder.ShippedAt == nil {
			http.Error(w, "Order not shipped yet", http.StatusBadRequest)
			return
		}
	default:
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	err = o.Repo.Update(r.Context(), *existingOrder)
	if err != nil {
		http.Error(w, "Failed to update order", http.StatusInternalServerError)
		return
	}
	res, err := json.Marshal(existingOrder)
	if err != nil {
		http.Error(w, "Failed to marshal order", http.StatusInternalServerError)
		return
	}
	w.Write(res)
	w.WriteHeader(http.StatusOK)
}
func (o *Order) DeleteById(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	const base = 10
	const bitSize = 64
	orderID, err := strconv.ParseUint(idParam, base, bitSize)
	if err != nil {
		http.Error(w, "Invalid order ID", http.StatusBadRequest)
		return
	}
	err = o.Repo.DeleteById(r.Context(), orderID)
	if err != nil {
		http.Error(w, "Failed to delete order", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)

}
