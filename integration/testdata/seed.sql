-- Test data for GuoceDB integration tests
-- This file contains seed data for testing various SQL operations

USE testdb;

-- Insert test users
INSERT INTO users (id, name, email, age) VALUES
    (1, 'Alice', 'alice@example.com', 25),
    (2, 'Bob', 'bob@example.com', 30),
    (3, 'Charlie', 'charlie@example.com', 35),
    (4, 'Diana', 'diana@example.com', 28),
    (5, 'Eve', 'eve@example.com', 22);

-- Insert test products
INSERT INTO products (id, name, price, stock) VALUES
    (1, 'Laptop', 999.99, 50),
    (2, 'Mouse', 29.99, 200),
    (3, 'Keyboard', 79.99, 150),
    (4, 'Monitor', 299.99, 75),
    (5, 'Headphones', 149.99, 100);

-- Insert test categories
INSERT INTO categories (id, name, description) VALUES
    (1, 'Electronics', 'Electronic devices and accessories'),
    (2, 'Computers', 'Computer hardware and peripherals'),
    (3, 'Audio', 'Audio equipment and accessories');

-- Insert product category mappings
INSERT INTO product_categories (product_id, category_id) VALUES
    (1, 1), (1, 2),  -- Laptop: Electronics, Computers
    (2, 1), (2, 2),  -- Mouse: Electronics, Computers
    (3, 1), (3, 2),  -- Keyboard: Electronics, Computers
    (4, 1), (4, 2),  -- Monitor: Electronics, Computers
    (5, 1), (5, 3);  -- Headphones: Electronics, Audio

-- Insert test orders
INSERT INTO orders (id, user_id, amount, status) VALUES
    (1, 1, 1029.98, 'completed'),
    (2, 2, 79.99, 'pending'),
    (3, 1, 299.99, 'shipped'),
    (4, 3, 29.99, 'completed'),
    (5, 4, 149.99, 'pending');

-- Insert test order items
INSERT INTO order_items (id, order_id, product_id, quantity, price) VALUES
    (1, 1, 1, 1, 999.99),   -- Alice's laptop
    (2, 1, 2, 1, 29.99),    -- Alice's mouse
    (3, 2, 3, 1, 79.99),    -- Bob's keyboard
    (4, 3, 4, 1, 299.99),   -- Alice's monitor
    (5, 4, 2, 1, 29.99),    -- Charlie's mouse
    (6, 5, 5, 1, 149.99);   -- Diana's headphones