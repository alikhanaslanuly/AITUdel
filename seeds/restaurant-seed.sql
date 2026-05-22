
INSERT INTO categories (id, name) VALUES
  (1, 'Fast Food'),
  (2, 'Asian'),
  (3, 'Pizza'),
  (4, 'Healthy'),
  (5, 'Burgers')
ON CONFLICT (id) DO NOTHING;

INSERT INTO restaurants (id, category_id, name, description, rating, open_time, close_time) VALUES
  (1, 4, 'AITU Cafeteria',    'The main campus canteen — salads, soups and daily specials.', 4.5, '08:00', '20:00'),
  (2, 5, 'Burger House',      'Juicy smash-burgers and crispy fries made to order.',          4.2, '10:00', '23:00'),
  (3, 2, 'Dragon Wok',        'Authentic wok-fried noodles and dim-sum from scratch.',        4.7, '11:00', '22:00'),
  (4, 3, 'Pizza Planet',      'Stone-baked Neapolitan-style pizzas with 20 toppings.',       4.0, '11:00', '23:00'),
  (5, 1, 'Astana Street Eats','Street-food classics: shawarma, hotdogs, and more.',           4.3, '09:00', '02:00')
ON CONFLICT (id) DO NOTHING;

INSERT INTO menu_items (restaurant_id, name, description, price, stock) VALUES
  (1, 'Caesar Salad',      'Romaine, parmesan, croutons, house dressing',          950,  30),
  (1, 'Chicken Soup',      'Home-style broth with vegetables and noodles',         700,  25),
  (1, 'Grilled Salmon',    'Atlantic salmon fillet with roasted vegetables',       1800, 15),
  (1, 'Quinoa Bowl',       'Quinoa, avocado, chickpeas, tahini drizzle',           1200, 20),
  (1, 'Green Smoothie',    'Spinach, banana, apple juice, ginger',                 550,  40);

INSERT INTO menu_items (restaurant_id, name, description, price, stock) VALUES
  (2, 'Classic Smash',     'Double smash patty, cheddar, pickles, special sauce', 1400, 50),
  (2, 'BBQ Bacon Burger',  'Crispy bacon, BBQ sauce, caramelised onions',          1600, 40),
  (2, 'Veggie Burger',     'Black-bean patty, guacamole, jalapeños',               1300, 30),
  (2, 'Cheeseburger Kids', 'Single patty, mild ketchup, for little ones',          900,  25),
  (2, 'Loaded Fries',      'Crinkle fries, cheese sauce, bacon bits',              750,  60);

INSERT INTO menu_items (restaurant_id, name, description, price, stock) VALUES
  (3, 'Pad Thai',          'Rice noodles, shrimp, peanuts, tamarind sauce',        1350, 35),
  (3, 'Beef Pho',          'Vietnamese broth, rice noodles, fresh herbs',           1200, 30),
  (3, 'Dumplings (8pc)',   'Steamed pork and cabbage dumplings with dip',           1100, 40),
  (3, 'Fried Rice',        'Wok-fried jasmine rice with egg and vegetables',        950,  45),
  (3, 'Mango Sticky Rice', 'Thai dessert with fresh mango and coconut cream',       700,  20);

INSERT INTO menu_items (restaurant_id, name, description, price, stock) VALUES
  (4, 'Margherita',        'San Marzano tomato, fresh mozzarella, basil',          1500, 30),
  (4, 'Pepperoni Classic', 'Double pepperoni, mozzarella, tomato base',             1700, 35),
  (4, 'BBQ Chicken',       'Pulled chicken, red onion, BBQ drizzle',               1800, 25),
  (4, 'Quattro Formaggi',  'Four cheese blend on white cream base',                1900, 20),
  (4, 'Garlic Bread',      'Toasted ciabatta with herb butter',                    500,  50);

INSERT INTO menu_items (restaurant_id, name, description, price, stock) VALUES
  (5, 'Chicken Shawarma',  'Marinated chicken, pickles, garlic sauce in lavash',   900,  60),
  (5, 'Beef Shawarma',     'Slow-roasted beef with hummus and fresh veggies',       1050, 50),
  (5, 'Hot Dog NYC-style', 'Beef frank, mustard, sauerkraut, relish',              650,  70),
  (5, 'Samsa (3pc)',       'Flaky pastry stuffed with spiced lamb and onion',       800,  40),
  (5, 'Fresh Lemonade',    'Hand-squeezed lemon, mint, sparkling water',            400,  80);

INSERT INTO reviews (restaurant_id, user_id, rating, comment) VALUES
  (1, 1, 5, 'Best campus lunch spot — the quinoa bowl is amazing!'),
  (1, 2, 4, 'Affordable and fresh. Soup could be a bit hotter.'),
  (2, 1, 4, 'Smash burger was incredibly juicy. Will be back!'),
  (2, 3, 5, 'Loaded fries are life-changing. 10/10.'),
  (3, 2, 5, 'Dragon Wok has the most authentic Pad Thai in Astana.'),
  (3, 1, 5, 'Dumplings are hand-made, you can taste the difference.'),
  (4, 3, 4, 'Great pizza, dough was a bit thick on the edges.'),
  (4, 2, 3, 'Delivery was slow but quality was good.'),
  (5, 1, 4, 'Shawarma is perfect for a quick bite between lectures.'),
  (5, 3, 5, 'Best value in Astana. Samsa is outstanding.');

SELECT setval('categories_id_seq', (SELECT MAX(id) FROM categories));
SELECT setval('restaurants_id_seq', (SELECT MAX(id) FROM restaurants));
