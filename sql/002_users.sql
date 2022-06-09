-- +goose Up
-- +goose StatementBegin
CREATE TABLE users (
-- id is telegram id
-- kind is the group the user is in (same constants as points/kind)

id               INTEGER      NOT NULL PRIMARY KEY,
name             VARCHAR(255) NOT NULL,
kind             INTEGER NOT NULL,
currentPoint     VARCHAR(10) NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE users;
-- +goose StatementEnd
