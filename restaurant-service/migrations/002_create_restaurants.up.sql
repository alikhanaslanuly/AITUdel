CREATE TABLE restaurants (
    id SERIAL PRIMARY KEY,
    category_id INT REFERENCES categories(id),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    rating FLOAT DEFAULT 0,
    open_time TIME,
    close_time TIME
);