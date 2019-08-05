-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied

INSERT INTO payments (id, guid, account_id, amount, created_at, updated_at) VALUES
    (DEFAULT, uuid_send('e9b0c72f-c08f-4e00-b158-42ae88f0c18e'::uuid), 1, 3.33, DEFAULT, DEFAULT),
    (DEFAULT,  uuid_send('e9b0c72f-c08f-4e00-b158-42ae88f0c18e'::uuid), 2, -3.33, DEFAULT, DEFAULT);

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
