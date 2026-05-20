CREATE TABLE menu_items (
    id SERIAL PRIMARY KEY,
    restaurant_id INT REFERENCES restaurants(id),
    name VARCHAR(255),
    description TEXT,
    price NUMERIC(10,2),
    stock INT DEFAULT 0
);