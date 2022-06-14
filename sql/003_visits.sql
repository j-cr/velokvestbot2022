-- +goose Up
-- +goose StatementBegin
CREATE TABLE visits (

-- id is autogenerated (optional, probably can be deleted)
-- user+point is unique: each user visits each point at most once
-- photo is file_unique_id
-- see: https://core.telegram.org/bots/api#photosize

id         INTEGER      NOT NULL  PRIMARY KEY AUTOINCREMENT,
user       INTEGER      NOT NULL,
point      TEXT         NOT NULL,
photo      TEXT         NOT NULL  UNIQUE,
score      INTEGER      NOT NULL,
added      INTEGER      NOT NULL,
UNIQUE(user, point)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE visits;
-- +goose StatementEnd
