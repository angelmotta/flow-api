CREATE TABLE users_state (
    user_state VARCHAR(10) PRIMARY KEY,
    description TEXT
);

insert into users_state(user_state, description)
values ('registered', 'Signup completado pero falta ingresar al menos una cuenta bancaria')
insert into users_state(user_state, description)
values ('active', 'Onboarding completado. Usuario tiene registrado al menos una cuenta bancaria')
insert into users_state(user_state, description)
values ('blocked', 'Usuario bloqueado requiere contactarse con la casa de cambio')
insert into users_state(user_state, description)
values ('deleted', 'Usuario solicitó su baja del servicio')

CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(100) NOT NULL UNIQUE,
    role VARCHAR(20) NOT NULL,
    dni VARCHAR(8) NOT NULL UNIQUE,
    name VARCHAR(100) NOT NULL,
    lastname_main VARCHAR(50) NOT NULL,
    lastname_secondary VARCHAR(50) NOT NULL,
    address VARCHAR(100) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    state VARCHAR(20) NOT NULL REFERENCES users_state ON DELETE RESTRICT ON UPDATE CASCADE,
    deleted_at TIMESTAMP
);

INSERT INTO users (email, role, dni, name, lastname_main, lastname_secondary, address, state)
VALUES ('angelmotta@gmail.com', 'customer', '12345678', 'Angel', 'Motta', 'Paz', 'Manuel Pazos 709', 'registered');

-- Bank Accounts

create table banks(
    bank_name VARCHAR(20) primary key,
    full_name VARCHAR(50) not null
);

insert into banks(bank_name, full_name)
values ('BCP', 'Banco de Crédito del Perú');

insert into banks(bank_name, full_name)
values ('BBVA', 'BBVA Perú');

CREATE TABLE bank_accounts (
    id SERIAL PRIMARY KEY,
    account_number VARCHAR(50) not null,
    currency_type VARCHAR(5) not null references currencies on delete restrict on update cascade,
    bank_name VARCHAR(20) not null references banks on delete restrict on update cascade,
    user_id SERIAL not null references users on delete restrict on update cascade,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    unique (user_id, account_number)
);

delete from bank_accounts;
drop table bank_accounts;

insert into bank_accounts (account_number, currency_type, bank_name, user_id)
values ('123456789001', 'PEN', 'BCP' ,3);

insert into bank_accounts (account_number, currency_type, bank_name, user_id)
values ('123456789002', 'PEN', 'BBVA', 3);

select * from bank_accounts;

-- Exchanges
CREATE TABLE exchange_currency (
    exchange_id VARCHAR(10) primary key,
    currency_main VARCHAR(5) not null references currencies on delete restrict on update cascade,
    currency_secondary VARCHAR(5) not null references currencies on delete restrict on update cascade,
    buy_price_currency_main NUMERIC(6, 3) NOT NULL,
    sale_price_currency_main NUMERIC(6, 3) NOT NULL,
    minimum_valid_time_mins INTEGER NOT NULL,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

insert into exchange_currency (exchange_id, currency_main, currency_secondary, buy_price_currency_main, sale_price_currency_main, minimum_valid_time_mins)
values ('USD-PEN', 'USD', 'PEN', 3.727, 3.732, 3);

select * from exchange_currency;

-- Orders
CREATE TABLE order_type (
    order_type VARCHAR(10) PRIMARY KEY,
    description TEXT
);

insert into order_type (order_type, description)
values ('buy', 'Casa compra la moneda principal');

insert into order_type (order_type, description)
values ('sell', 'Casa vende la moneda principal');

CREATE TABLE order_type (
    order_type VARCHAR(10) PRIMARY KEY,
    description TEXT
);

CREATE TABLE order_state (
     order_state VARCHAR(20) PRIMARY KEY,
     description TEXT
);

insert into order_state (order_state, description)
values
    ('pending', 'Orden registrada'),
    ('inprogress', 'Orden siendo actualmente atendida'),
    ('finished', 'Orden terminada. La casa depositó el dinero solicitado en la orden')

select * from order_state