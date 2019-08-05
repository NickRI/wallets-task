-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied

CREATE TABLE IF NOT EXISTS accounts (
  id         serial PRIMARY KEY,
  user_name  varchar(64) NOT NULL,
  balance    decimal NOT NULL DEFAULT 0,
  currency   varchar(4) NOT NULL DEFAULT 0,
  created_at timestamp with time zone NOT NULL DEFAULT NOW(),
  updated_at timestamp with time zone NOT NULL DEFAULT NOW(),
  UNIQUE (user_name)
);



-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back

DROP TABLE IF EXISTS accounts;