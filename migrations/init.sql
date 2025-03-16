-- Create the database testdb2 and testdb3
CREATE DATABASE IF NOT EXISTS testdb2;
CREATE DATABASE IF NOT EXISTS testdb3;
CREATE DATABASE IF NOT EXISTS fakedb;
USE testdb;

-- Create the 'users' table
CREATE TABLE users (
                       id INT AUTO_INCREMENT PRIMARY KEY,
                       name VARCHAR(100) NOT NULL,
                       email VARCHAR(100) NOT NULL UNIQUE,
                       created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create the 'orders' table
CREATE TABLE orders (
                        id INT AUTO_INCREMENT PRIMARY KEY,
                        user_id INT NOT NULL,
                        amount DECIMAL(10,2) NOT NULL,
                        status ENUM('pending', 'completed', 'canceled') NOT NULL DEFAULT 'pending',
                        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                        FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Insert fake users
INSERT INTO users (name, email) VALUES
                                    ('Alice Smith', 'alice@example.com'),
                                    ('Bob Johnson', 'bob@example.com'),
                                    ('Charlie Brown', 'charlie@example.com');

-- Insert fake orders
INSERT INTO orders (user_id, amount, status) VALUES
                                                 (1, 100.50, 'completed'),
                                                 (2, 200.75, 'pending'),
                                                 (3, 50.00, 'canceled');
