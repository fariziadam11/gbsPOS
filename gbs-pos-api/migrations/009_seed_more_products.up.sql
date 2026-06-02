-- Seed expanded product catalog (~40 products)
-- All products have Unsplash placeholder images

INSERT INTO products (name, price, category, image_url, store_type, created_at, updated_at) VALUES
-- RETAIL: Snacks (6)
('Chitato', 11500.00, 'Snacks', 'https://images.unsplash.com/photo-1621939514649-28b12e81658b', 'RETAIL', NOW(), NOW()),
('Indomie Goreng', 3500.00, 'Snacks', 'https://images.unsplash.com/photo-1612929633738-8fe44f7ec841', 'RETAIL', NOW(), NOW()),
('Tango Wafer', 8500.00, 'Snacks', 'https://images.unsplash.com/photo-1543508168-1f7db3e5a125', 'RETAIL', NOW(), NOW()),
('Oreo Cookies', 12000.00, 'Snacks', 'https://images.unsplash.com/photo-1558961363-fa8fdf82db35', 'RETAIL', NOW(), NOW()),
('Pop Mie', 6500.00, 'Snacks', 'https://images.unsplash.com/photo-1596627689632-d9162125b34f', 'RETAIL', NOW(), NOW()),
('Keripik Singkong', 9000.00, 'Snacks', 'https://images.unsplash.com/photo-1599494779852-139e1315cf3c', 'RETAIL', NOW(), NOW()),

-- RETAIL: Beverages (5)
('Teh Botol Sosro', 5000.00, 'Beverages', 'https://images.unsplash.com/photo-1556679343-c7306c1976bc', 'RETAIL', NOW(), NOW()),
('Aqua Mineral Water 600ml', 4500.00, 'Beverages', 'https://images.unsplash.com/photo-1548839140-29a749e1cf4d', 'RETAIL', NOW(), NOW()),
('Coca-Cola 390ml', 7500.00, 'Beverages', 'https://images.unsplash.com/photo-1622483762418-e391c6a17436', 'RETAIL', NOW(), NOW()),
('Ultra Milk 1L', 18500.00, 'Beverages', 'https://images.unsplash.com/photo-1563636619-e9143da7973b', 'RETAIL', NOW(), NOW()),
('Nescafe Original', 25000.00, 'Beverages', 'https://images.unsplash.com/photo-1514432324607-a09d9b4aefdd', 'RETAIL', NOW(), NOW()),

-- RETAIL: Personal Care (3)
('Sabun Mandi Lifebuoy', 8000.00, 'Personal Care', 'https://images.unsplash.com/photo-1556228578-0d85b1a4d571', 'RETAIL', NOW(), NOW()),
('Shampoo Clear 170ml', 22000.00, 'Personal Care', 'https://images.unsplash.com/photo-1585751119414-ef2636f8aede', 'RETAIL', NOW(), NOW()),
('Pasta Gigi Pepsodent', 12500.00, 'Personal Care', 'https://images.unsplash.com/photo-1559599101-f09722fb4948', 'RETAIL', NOW(), NOW()),

-- RETAIL: Household (3)
('Pembersih Lantai', 15000.00, 'Household', 'https://images.unsplash.com/photo-1585421514284-efb74c2b69ba', 'RETAIL', NOW(), NOW()),
('Deterjen Rinso 1kg', 32000.00, 'Household', 'https://images.unsplash.com/photo-1582735689369-4fe89db71144', 'RETAIL', NOW(), NOW()),
('Tissue Paseo 250s', 18000.00, 'Household', 'https://images.unsplash.com/photo-1583947215259-38e31be8751f', 'RETAIL', NOW(), NOW()),

-- RETAIL: Electronics (2)
('Baterai AA Panasonic (4pcs)', 28000.00, 'Electronics', 'https://images.unsplash.com/photo-1598391998735-55ebe996d3a1', 'RETAIL', NOW(), NOW()),
('Lampu LED Philips 9W', 35000.00, 'Electronics', 'https://images.unsplash.com/photo-1565814329459-e7e779f0f399', 'RETAIL', NOW(), NOW()),

-- FNB: Food (5)
('Nasi Goreng Special', 25000.00, 'Food', 'https://images.unsplash.com/photo-1512058564366-18510be2db19', 'FNB', NOW(), NOW()),
('Nasi Padang Rendang', 35000.00, 'Food', 'https://images.unsplash.com/photo-1603089645312-4f2c3f79c4b5', 'FNB', NOW(), NOW()),
('Mie Ayam Bakso', 22000.00, 'Food', 'https://images.unsplash.com/photo-1552611052-33e04de081de', 'FNB', NOW(), NOW()),
('Sate Ayam (10 tusuk)', 30000.00, 'Food', 'https://images.unsplash.com/photo-1555939594-58d7cb561ad1', 'FNB', NOW(), NOW()),
('Gado-Gado', 18000.00, 'Food', 'https://images.unsplash.com/photo-1551529834-525807d6b4f3', 'FNB', NOW(), NOW()),

-- FNB: Beverages (4)
('Es Teh Manis', 8000.00, 'Beverages', 'https://images.unsplash.com/photo-1556679343-c7306c1976bc', 'FNB', NOW(), NOW()),
('Es Jeruk', 10000.00, 'Beverages', 'https://images.unsplash.com/photo-1603569283847-aa295f0d016a', 'FNB', NOW(), NOW()),
('Kopi Susu', 12000.00, 'Beverages', 'https://images.unsplash.com/photo-1517701550927-30cf4ba1dba5', 'FNB', NOW(), NOW()),
('Jus Alpukat', 18000.00, 'Beverages', 'https://images.unsplash.com/photo-1594481532377-84c7b90e4657', 'FNB', NOW(), NOW()),

-- FNB: Desserts (3)
('Pisang Goreng (3pcs)', 12000.00, 'Desserts', 'https://images.unsplash.com/photo-1528975604071-b4dc52a2d18c', 'FNB', NOW(), NOW()),
('Es Krim Cone', 8000.00, 'Desserts', 'https://images.unsplash.com/photo-1497034825429-c343d7c6a68f', 'FNB', NOW(), NOW()),
('Pancake Coklat', 20000.00, 'Desserts', 'https://images.unsplash.com/photo-1567620905732-2d1ec7ab7445', 'FNB', NOW(), NOW()),

-- OUTFIT: Tops (4)
('Kaos Polos Basic', 75000.00, 'Tops', 'https://images.unsplash.com/photo-1521572163474-6864f9cf17ab', 'OUTFIT', NOW(), NOW()),
('Kemeja Flanel', 150000.00, 'Tops', 'https://images.unsplash.com/photo-1596755094514-f87e34085b2c', 'OUTFIT', NOW(), NOW()),
('Polo Shirt Lacoste', 285000.00, 'Tops', 'https://images.unsplash.com/photo-1586790170083-2f9ceadc732d', 'OUTFIT', NOW(), NOW()),
('T-Shirt Graphic', 95000.00, 'Tops', 'https://images.unsplash.com/photo-1576566588028-4147f3842f27', 'OUTFIT', NOW(), NOW()),

-- OUTFIT: Bottoms (3)
('Celana Jeans Slim Fit', 250000.00, 'Bottoms', 'https://images.unsplash.com/photo-1542272604-787c3835535d', 'OUTFIT', NOW(), NOW()),
('Celana Chino', 175000.00, 'Bottoms', 'https://images.unsplash.com/photo-1473966968600-fa801b869a1a', 'OUTFIT', NOW(), NOW()),
('Celana Training', 120000.00, 'Bottoms', 'https://images.unsplash.com/photo-1517438476312-64d2e5d50284', 'OUTFIT', NOW(), NOW()),

-- OUTFIT: Outerwear (3)
('Jaket Hoodie', 185000.00, 'Outerwear', 'https://images.unsplash.com/photo-1556821840-3a63f95609a7', 'OUTFIT', NOW(), NOW()),
('Bomber Jacket', 320000.00, 'Outerwear', 'https://images.unsplash.com/photo-1591047139829-d91daecb44c3', 'OUTFIT', NOW(), NOW()),
('Cardigan Rajut', 210000.00, 'Outerwear', 'https://images.unsplash.com/photo-1434389677669-e08b4a93e3c0', 'OUTFIT', NOW(), NOW()),

-- OUTFIT: Accessories & Footwear (3)
('Topi Baseball', 55000.00, 'Accessories', 'https://images.unsplash.com/photo-1588850561407-ed78c282e89b', 'OUTFIT', NOW(), NOW()),
('Tote Bag Canvas', 45000.00, 'Accessories', 'https://images.unsplash.com/photo-1553062407-98eeb64c6a62', 'OUTFIT', NOW(), NOW()),
('Sneakers Casual', 350000.00, 'Footwear', 'https://images.unsplash.com/photo-1542291026-7eec264c27ff', 'OUTFIT', NOW(), NOW());
