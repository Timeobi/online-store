
CREATE TABLE users (
    id            SERIAL PRIMARY KEY,
    email         VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    role          VARCHAR(20) NOT NULL DEFAULT 'customer',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE categories (
    id         SERIAL PRIMARY KEY,
    name       VARCHAR(100) NOT NULL,
    slug       VARCHAR(120) NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE products (
    id             SERIAL PRIMARY KEY,
    category_id    INTEGER REFERENCES categories(id) ON DELETE SET NULL,
    name           VARCHAR(255) NOT NULL,
    slug           VARCHAR(280) NOT NULL UNIQUE,
    description    TEXT,
    price          INTEGER NOT NULL CHECK (price >= 0),
    stock_quantity INTEGER NOT NULL DEFAULT 0 CHECK (stock_quantity >= 0),
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE carts (
    id         SERIAL PRIMARY KEY,
    user_id    INTEGER NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE cart_items (
    id         SERIAL PRIMARY KEY,
    cart_id    INTEGER NOT NULL REFERENCES carts(id) ON DELETE CASCADE,
    product_id INTEGER NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    quantity   INTEGER NOT NULL CHECK (quantity > 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (cart_id, product_id)
);

CREATE TABLE orders (
    id           SERIAL PRIMARY KEY,
    user_id      INTEGER NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    status       VARCHAR(20) NOT NULL DEFAULT 'pending',
    total_amount INTEGER NOT NULL CHECK (total_amount >= 0),
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE order_items (
    id                 SERIAL PRIMARY KEY,
    order_id           INTEGER NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    product_id         INTEGER NOT NULL REFERENCES products(id) ON DELETE RESTRICT,
    quantity           INTEGER NOT NULL CHECK (quantity > 0),
    price_at_purchase  INTEGER NOT NULL CHECK (price_at_purchase >= 0),
    created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Индексы для ускорения частых запросов
CREATE INDEX idx_products_category_id ON products(category_id);
CREATE INDEX idx_cart_items_cart_id ON cart_items(cart_id);
CREATE INDEX idx_orders_user_id ON orders(user_id);
CREATE INDEX idx_order_items_order_id ON order_items(order_id);