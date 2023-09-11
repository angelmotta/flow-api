CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(100) NOT NULL UNIQUE,
    role VARCHAR(20) NOT NULL,
    dni VARCHAR(8) NOT NULL UNIQUE,
    name VARCHAR(100) NOT NULL,
    lastname_main VARCHAR(50) NOT NULL,
    lastname_secondary VARCHAR(50) NOT NULL,
    address VARCHAR(100) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE exchanges (
    id SERIAL PRIMARY KEY,
    pair VARCHAR(10) NOT NULL,
    buy_price DECIMAL(10, 2) NOT NULL,
    sell_price DECIMAL(10, 2) NOT NULL,
    valid_duration INTEGER NOT NULL,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

--- insert into users one record
INSERT INTO users (email, role, dni, name, lastname_main, lastname_secondary, address)
VALUES ('angelmotta@gmail.com', 'customer', '12345678', 'Angel', 'Motta', 'Paz', 'Manuel Pazos 709');
