-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied

CREATE TABLE IF NOT EXISTS payments (
  id                     serial PRIMARY KEY,
  guid                   bytea NOT NULL,
  account_id             integer NOT NULL,
  amount                 decimal NOT NULL DEFAULT 0,
  created_at             timestamp with time zone NOT NULL DEFAULT NOW(),
  updated_at             timestamp with time zone NOT NULL DEFAULT NOW(),
  FOREIGN KEY (account_id) REFERENCES accounts (id)
);

CREATE INDEX ON payments(account_id);

CREATE INDEX ON payments(amount) WHERE amount > 0;

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back

DROP TABLE IF EXISTS payments;
