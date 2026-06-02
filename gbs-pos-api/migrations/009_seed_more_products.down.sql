-- Revert: remove all products seeded by 009 migration
-- Original 11 products (from 008_seed_data) are preserved

DELETE FROM products
WHERE name IN (
    'Chitato', 'Indomie Goreng', 'Tango Wafer', 'Oreo Cookies', 'Pop Mie', 'Keripik Singkong',
    'Teh Botol Sosro', 'Aqua Mineral Water 600ml', 'Coca-Cola 390ml', 'Ultra Milk 1L', 'Nescafe Original',
    'Sabun Mandi Lifebuoy', 'Shampoo Clear 170ml', 'Pasta Gigi Pepsodent',
    'Pembersih Lantai', 'Deterjen Rinso 1kg', 'Tissue Paseo 250s',
    'Baterai AA Panasonic (4pcs)', 'Lampu LED Philips 9W',
    'Nasi Goreng Special', 'Nasi Padang Rendang', 'Mie Ayam Bakso', 'Sate Ayam (10 tusuk)', 'Gado-Gado',
    'Es Teh Manis', 'Es Jeruk', 'Kopi Susu', 'Jus Alpukat',
    'Pisang Goreng (3pcs)', 'Es Krim Cone', 'Pancake Coklat',
    'Kaos Polos Basic', 'Kemeja Flanel', 'Polo Shirt Lacoste', 'T-Shirt Graphic',
    'Celana Jeans Slim Fit', 'Celana Chino', 'Celana Training',
    'Jaket Hoodie', 'Bomber Jacket', 'Cardigan Rajut',
    'Topi Baseball', 'Tote Bag Canvas', 'Sneakers Casual'
);
