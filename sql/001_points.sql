-- +goose Up
-- +goose StatementBegin
CREATE TABLE points (
-- id is any short string
-- TODO: maybe use autoinc id?
id       TEXT      NOT NULL  PRIMARY KEY,
name     TEXT      NOT NULL,
kind     INTEGER   NOT NULL,
url      TEXT      NOT NULL
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE points;
-- +goose StatementEnd
