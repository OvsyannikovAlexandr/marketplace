package domain

import "time"

// Product представляет товар на маркетплейсе
// swagger:model
type Product struct {
	// ID уникальный идентификатор продукта
	ID int64 `json:"id"`
	// Name название продукта
	Name string `json:"name"`
	// Description описание продукта
	Description string `json:"description"`
	// Price цена продукта в валюте USD
	Price float64 `json:"price"`
	// CreatedAt дата и время создания продукта
	CreatedAt time.Time `json:"created_at"`
	// UpdatedAt дата и время последнего обновления продукта
	UpdatedAt time.Time `json:"updated_at"`
}
