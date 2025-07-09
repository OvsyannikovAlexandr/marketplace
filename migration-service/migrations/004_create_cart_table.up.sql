CREATE SCHEMA IF NOT EXISTS cart_service;

CREATE TABLE IF NOT EXISTS cart_service.cart_items (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    product_id INTEGER NOT NULL,
    quantity INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    CONSTRAINT cart_items_user_product_unique UNIQUE (user_id, product_id)
)