CREATE TABLE IF NOT EXISTS order_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    item_id UUID NOT NULL,
    name VARCHAR(200) NOT NULL,
    quantity INT NOT NULL,
    price NUMERIC(10,2) NOT NULL
);