-- Create the 'users' table
CREATE TABLE users (
                       id SERIAL PRIMARY KEY,
                       name VARCHAR(100) NOT NULL,
                       email VARCHAR(100) NOT NULL UNIQUE,
                       created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create the 'orders' table
CREATE TABLE orders (
                        id SERIAL PRIMARY KEY,
                        user_id INT NOT NULL,
                        amount DECIMAL(10,2) NOT NULL,
                        status TEXT NOT NULL CHECK (status IN ('pending', 'completed', 'canceled')),
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
