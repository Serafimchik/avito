-- +goose Up
-- +goose StatementBegin
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    coins BIGINT NOT NULL DEFAULT 1000
);

CREATE TABLE items (
    type TEXT PRIMARY KEY,
    price BIGINT NOT NULL CHECK (price >= 0)
);

INSERT INTO items (type, price) VALUES
    ('t-shirt', 80),
    ('cup', 20),
    ('book', 50),
    ('pen', 10),
    ('powerbank', 200),
    ('hoody', 300),
    ('umbrella', 200),
    ('socks', 10),
    ('wallet', 50),
    ('pink-hoody', 500);

CREATE TABLE transactions (
    id SERIAL PRIMARY KEY,
    from_user_id INT REFERENCES users(id),
    to_user_id INT REFERENCES users(id),
    amount BIGINT NOT NULL CHECK (amount > 0),
    timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE inventory (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id),
    item_type TEXT NOT NULL REFERENCES items(type),
    quantity BIGINT NOT NULL DEFAULT 0,
    UNIQUE (user_id, item_type)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS inventory;
DROP TABLE IF EXISTS transactions;
DROP TABLE IF EXISTS items;
DROP TABLE IF EXISTS users;
-- +goose StatementEnd