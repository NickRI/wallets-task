-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied

INSERT INTO accounts (id, user_name, balance, currency, created_at, updated_at) VALUES
    (DEFAULT, 'alice123', 1000.334, 'USD', DEFAULT, DEFAULT),
    (DEFAULT, 'bob456', 1000.5324425, 'USD', DEFAULT, DEFAULT);

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back

;