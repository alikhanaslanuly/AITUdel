CREATE TABLE reviews (
    id SERIAL PRIMARY KEY,
    restaurant_id INT REFERENCES restaurants(id),
    user_id VARCHAR(100),
    rating INT CHECK (rating >=1 AND rating <=5),
    comment TEXT
);