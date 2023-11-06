-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';

CREATE TABLE users(
    id UUID PRIMARY KEY,
    email TEXT NOT NULL UNIQUE,
    first_name TEXT NOT NULL,
    last_name TEXT NOT NULL,
    password TEXT NOT NULL,
    role TEXT NOT NULL,
    CHECK (role IN ('regular', 'admin'))
);


CREATE TABLE folders(
    id UUID PRIMARY KEY,
    folder_name varchar(255) NOT NULL,
    owner_id UUID NOT NULL REFERENCES users(id), 
    "created_at" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE files(
    id UUID PRIMARY KEY,
    file_name varchar(255) NOT NULL,
    owner_id UUID NOT NULL REFERENCES users(id), 
    folder_id UUID NOT NULL REFERENCES folders(id), 
    file_store_key varchar(255) NOT NULL,
    file_size INTEGER NOT NULL,
    is_unsafe BOOL DEFAULT 'f',
    "created_at" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
DROP TABLE users;
DROP TABLE folders;
DROP TABLE files;

-- +goose StatementEnd
