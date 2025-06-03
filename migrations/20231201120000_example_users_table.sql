-- +goose Up
-- Create users table
CREATE TABLE users (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    age INT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL
);

-- Create index on email for faster lookups
CREATE INDEX idx_users_email ON users(email);

-- Create index on deleted_at for soft delete queries
CREATE INDEX idx_users_deleted_at ON users(deleted_at);

-- +goose Down
-- Drop indexes first
DROP INDEX idx_users_deleted_at ON users;
DROP INDEX idx_users_email ON users;

-- Drop the table
DROP TABLE users; 