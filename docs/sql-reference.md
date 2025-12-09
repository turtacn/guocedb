# GuoceDB SQL Reference

GuoceDB aims for MySQL compatibility. This document covers the currently supported SQL features.

## Data Definition Language (DDL)

### Databases

```sql
-- Create database
CREATE DATABASE [IF NOT EXISTS] database_name;

-- Use database
USE database_name;

-- Drop database
DROP DATABASE [IF EXISTS] database_name;

-- Show databases
SHOW DATABASES;
```

### Tables

```sql
-- Create table
CREATE TABLE [IF NOT EXISTS] table_name (
    column_name data_type [constraints],
    ...
    [PRIMARY KEY (columns)],
    [INDEX index_name (columns)]
);

-- Drop table
DROP TABLE [IF EXISTS] table_name;

-- Show tables
SHOW TABLES [FROM database_name];

-- Describe table
DESCRIBE table_name;
DESC table_name;
SHOW CREATE TABLE table_name;
```

## Data Types

| Type         | Description            | Example                    |
| ------------ | ---------------------- | -------------------------- |
| INT          | 32-bit integer         | `age INT`                  |
| BIGINT       | 64-bit integer         | `user_id BIGINT`           |
| FLOAT        | Floating point         | `score FLOAT`              |
| DOUBLE       | Double precision       | `price DOUBLE`             |
| DECIMAL(M,D) | Fixed decimal          | `amount DECIMAL(10,2)`     |
| VARCHAR(N)   | Variable string        | `name VARCHAR(100)`        |
| TEXT         | Long text              | `description TEXT`         |
| BLOB         | Binary data            | `image BLOB`               |
| DATE         | Date only              | `birthday DATE`            |
| DATETIME     | Date and time          | `created_at DATETIME`      |
| TIMESTAMP    | Unix timestamp         | `updated_at TIMESTAMP`     |
| BOOLEAN      | TRUE/FALSE             | `is_active BOOLEAN`        |

## Column Constraints

```sql
NOT NULL                -- Disallow NULL
DEFAULT value           -- Default value
PRIMARY KEY             -- Primary key
AUTO_INCREMENT          -- Auto-increment (INT types)
UNIQUE                  -- Unique constraint
```

Example:

```sql
CREATE TABLE users (
    id INT PRIMARY KEY AUTO_INCREMENT,
    email VARCHAR(255) NOT NULL UNIQUE,
    name VARCHAR(100) NOT NULL,
    age INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

## Data Manipulation Language (DML)

### INSERT

```sql
-- Single row
INSERT INTO table_name (col1, col2, ...) VALUES (val1, val2, ...);

-- Multiple rows
INSERT INTO table_name VALUES
    (val1, val2, ...),
    (val1, val2, ...);

-- All columns
INSERT INTO table_name VALUES (val1, val2, ...);
```

### SELECT

```sql
SELECT [DISTINCT] columns
FROM table_name
[WHERE condition]
[GROUP BY columns]
[HAVING condition]
[ORDER BY columns [ASC|DESC]]
[LIMIT count [OFFSET offset]];
```

### UPDATE

```sql
UPDATE table_name
SET column1 = value1, column2 = value2, ...
[WHERE condition];
```

### DELETE

```sql
DELETE FROM table_name
[WHERE condition];
```

## Operators

### Comparison

```sql
=, !=, <>, <, >, <=, >=
IS NULL, IS NOT NULL
IN (values...)
BETWEEN value1 AND value2
LIKE 'pattern'
```

### Logical

```sql
AND, OR, NOT
```

### Examples

```sql
SELECT * FROM users WHERE age >= 18 AND is_active = TRUE;
SELECT * FROM products WHERE price BETWEEN 10 AND 100;
SELECT * FROM users WHERE email LIKE '%@example.com';
SELECT * FROM orders WHERE user_id IN (1, 2, 3);
```

## Aggregate Functions

```sql
COUNT(*)          -- Count rows
COUNT(column)     -- Count non-NULL values
SUM(column)       -- Sum of values
AVG(column)       -- Average
MIN(column)       -- Minimum
MAX(column)       -- Maximum
```

Example:

```sql
SELECT 
    department,
    COUNT(*) as employee_count,
    AVG(salary) as avg_salary,
    MAX(salary) as max_salary
FROM employees
GROUP BY department
HAVING COUNT(*) > 5;
```

## Joins

```sql
-- Inner join
SELECT * FROM t1
INNER JOIN t2 ON t1.id = t2.t1_id;

-- Left join
SELECT * FROM t1
LEFT JOIN t2 ON t1.id = t2.t1_id;

-- Right join
SELECT * FROM t1
RIGHT JOIN t2 ON t1.id = t2.t1_id;

-- Cross join
SELECT * FROM t1 CROSS JOIN t2;
```

Example:

```sql
SELECT 
    users.name,
    orders.order_date,
    orders.total
FROM users
LEFT JOIN orders ON users.id = orders.user_id
WHERE orders.order_date > '2024-01-01';
```

## Subqueries

```sql
-- Scalar subquery
SELECT * FROM products
WHERE price > (SELECT AVG(price) FROM products);

-- IN subquery
SELECT * FROM users
WHERE id IN (SELECT user_id FROM orders WHERE total > 1000);

-- EXISTS subquery
SELECT * FROM users u
WHERE EXISTS (
    SELECT 1 FROM orders o WHERE o.user_id = u.id
);
```

## Set Operations

```sql
-- UNION (remove duplicates)
SELECT col FROM table1
UNION
SELECT col FROM table2;

-- UNION ALL (keep duplicates)
SELECT col FROM table1
UNION ALL
SELECT col FROM table2;

-- INTERSECT
SELECT col FROM table1
INTERSECT
SELECT col FROM table2;

-- EXCEPT
SELECT col FROM table1
EXCEPT
SELECT col FROM table2;
```

## Transaction Control

```sql
-- Start transaction
BEGIN;
START TRANSACTION;

-- Commit changes
COMMIT;

-- Rollback changes
ROLLBACK;
```

Example:

```sql
BEGIN;
UPDATE accounts SET balance = balance - 100 WHERE id = 1;
UPDATE accounts SET balance = balance + 100 WHERE id = 2;
COMMIT;
```

## User Management

```sql
-- Create user
CREATE USER 'username'@'host' IDENTIFIED BY 'password';

-- Drop user
DROP USER 'username'@'host';

-- Change password
ALTER USER 'username'@'host' IDENTIFIED BY 'new_password';

-- Grant privileges
GRANT privilege_list ON database.table TO 'username'@'host';

-- Revoke privileges
REVOKE privilege_list ON database.table FROM 'username'@'host';

-- Show grants
SHOW GRANTS FOR 'username'@'host';
```

### Privileges

- `ALL` - All privileges
- `SELECT` - Read data
- `INSERT` - Insert data
- `UPDATE` - Update data
- `DELETE` - Delete data
- `CREATE` - Create databases/tables
- `DROP` - Drop databases/tables
- `ALTER` - Alter tables
- `INDEX` - Create/drop indexes

## Utility Statements

```sql
-- Server information
SHOW VARIABLES;
SHOW STATUS;
SHOW PROCESSLIST;

-- Database inspection
SHOW DATABASES;
SHOW TABLES [FROM database];
SHOW COLUMNS FROM table;
SHOW CREATE TABLE table;
SHOW INDEXES FROM table;
```

## Current Limitations

### Not Supported

1. **Stored Procedures & Functions** - Not implemented
2. **Triggers** - Not implemented
3. **Views** - Planned for future release
4. **Foreign Keys** - Parsed but not enforced
5. **Full-text Search** - Not implemented
6. **Spatial Data Types** - Not implemented
7. **ALTER TABLE** - Limited support

### Partial Support

1. **JSON Functions** - Basic operations only
2. **Window Functions** - Limited support
3. **CTEs (WITH clause)** - May have limitations
4. **String Functions** - Common functions supported
5. **Date Functions** - Basic functions supported

## Best Practices

### Use Indexes

```sql
CREATE TABLE users (
    id INT PRIMARY KEY,
    email VARCHAR(255),
    name VARCHAR(100),
    INDEX idx_email (email)
);
```

### Batch Inserts

```sql
-- Instead of multiple single inserts
INSERT INTO table VALUES (1, 'a'), (2, 'b'), (3, 'c');
```

### Use Transactions

```sql
BEGIN;
-- Multiple related operations
UPDATE ...
INSERT ...
DELETE ...
COMMIT;
```

### Limit Result Sets

```sql
SELECT * FROM large_table LIMIT 100;
```

## Examples

### User Registration System

```sql
CREATE DATABASE userdb;
USE userdb;

CREATE TABLE users (
    id INT PRIMARY KEY AUTO_INCREMENT,
    username VARCHAR(50) NOT NULL UNIQUE,
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_email (email)
);

CREATE TABLE user_profiles (
    user_id INT PRIMARY KEY,
    full_name VARCHAR(100),
    bio TEXT,
    avatar_url VARCHAR(500)
);

-- Insert user
INSERT INTO users (username, email, password_hash)
VALUES ('john_doe', 'john@example.com', 'hashed_password');

-- Get user with profile
SELECT u.*, p.full_name, p.bio
FROM users u
LEFT JOIN user_profiles p ON u.id = p.user_id
WHERE u.username = 'john_doe';
```

### E-commerce Orders

```sql
CREATE TABLE products (
    id INT PRIMARY KEY,
    name VARCHAR(200) NOT NULL,
    price DECIMAL(10,2) NOT NULL,
    stock INT DEFAULT 0
);

CREATE TABLE orders (
    id INT PRIMARY KEY AUTO_INCREMENT,
    user_id INT NOT NULL,
    total DECIMAL(10,2),
    order_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE order_items (
    id INT PRIMARY KEY AUTO_INCREMENT,
    order_id INT NOT NULL,
    product_id INT NOT NULL,
    quantity INT NOT NULL,
    price DECIMAL(10,2) NOT NULL
);

-- Create order transaction
BEGIN;
INSERT INTO orders (user_id, total) VALUES (123, 150.00);
SET @order_id = LAST_INSERT_ID();
INSERT INTO order_items (order_id, product_id, quantity, price)
VALUES (@order_id, 1, 2, 75.00);
UPDATE products SET stock = stock - 2 WHERE id = 1;
COMMIT;
```

## See Also

- [Architecture Documentation](architecture.md)
- [Deployment Guide](deployment.md)
- [Troubleshooting Guide](troubleshooting.md)
