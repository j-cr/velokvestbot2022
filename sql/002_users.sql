-- +goose Up
-- +goose StatementBegin
CREATE TABLE users (
-- id is telegram id
-- kind is the group the user is in (same constants as points/kind)

id               INTEGER      NOT NULL PRIMARY KEY,
name             TEXT         NOT NULL,
kind             INTEGER      NOT NULL,
currentPoint     TEXT         NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE users;
-- +goose StatementEnd
