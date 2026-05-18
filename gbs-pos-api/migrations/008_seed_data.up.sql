INSERT INTO users (username, password_hash, name, role) VALUES
('admin', '$2a$10$uIjrPVsZtsoK01VHa6VC8e0t3O62BpTnF/YomtOLAN0BF087eAah2', 'Admin User', 'ADMIN'),
('cashier', '$2a$10$7OgCWELW2gl7lL/dAmzFkeJVf540NN4ZboNCJYawE6to/b.Z5s/G2', 'Cashier User', 'CASHIER');

INSERT INTO products (name, price, category, image_url, store_type) VALUES
('Chitato', 11500, 'Snacks', 'https://images.unsplash.com/photo-1621939514649-28b12e81658b', 'RETAIL'),
('Indomie Goreng', 3500, 'Snacks', 'https://images.unsplash.com/photo-1612929633738-8fe44f7ec841', 'RETAIL'),
('Teh Botol', 5000, 'Beverages', 'https://images.unsplash.com/photo-1556679343-c7306c1976bc', 'RETAIL'),
('Sabun Mandi', 8000, 'Personal Care', 'https://images.unsplash.com/photo-1556228578-0d85b1a4d571', 'RETAIL'),
('Pembersih Lantai', 15000, 'Household', 'https://images.unsplash.com/photo-1585421514284-efb74c2b69ba', 'RETAIL'),
('Nasi Goreng', 25000, 'Food', 'https://images.unsplash.com/photo-1512058564366-18510be2db19', 'FNB'),
('Es Teh Manis', 8000, 'Beverages', 'https://images.unsplash.com/photo-1556679343-c7306c1976bc', 'FNB'),
('Pisang Goreng', 12000, 'Desserts', 'https://images.unsplash.com/photo-1528975604071-b4dc52a2d18c', 'FNB'),
('Kaos Polos', 75000, 'Tops', 'https://images.unsplash.com/photo-1521572163474-6864f9cf17ab', 'OUTFIT'),
('Celana Jeans', 250000, 'Bottoms', 'https://images.unsplash.com/photo-1542272604-787c3835535d', 'OUTFIT'),
('Jaket Hoodie', 185000, 'Outerwear', 'https://images.unsplash.com/photo-1556821840-3a63f95609a7', 'OUTFIT');
