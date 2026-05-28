package database

import (
	"gbs-pos-api/internal/model"
	"log"

	"gorm.io/gorm"
)

func Seed(db *gorm.DB) {
	var count int64
	db.Model(&model.User{}).Count(&count)
	if count > 0 {
		return
	}
	log.Println("Seeding data...")
	users := []model.User{
		{
			Username:     "admin",
			PasswordHash: "$2a$10$uIjrPVsZtsoK01VHa6VC8e0t3O62BpTnF/YomtOLAN0BF087eAah2",
			Name:         "Admin User",
			Role:         "ADMIN",
		},
		{
			Username:     "cashier",
			PasswordHash: "$2a$10$7OgCWELW2gl7lL/dAmzFkeJVf540NN4ZboNCJYawE6to/b.Z5s/G2",
			Name:         "Cashier User",
			Role:         "CASHIER",
		},
	}
	for _, u := range users {
		db.Create(&u)
	}
	products := []model.Product{
		{
			Name:      "Chitato",
			Price:     11500,
			Category:  "Snacks",
			ImageURL:  "https://images.unsplash.com/photo-1621939514649-28b12e81658b",
			StoreType: "RETAIL",
		},
		{
			Name:      "Indomie Goreng",
			Price:     3500,
			Category:  "Snacks",
			ImageURL:  "https://images.unsplash.com/photo-1612929633738-8fe44f7ec841",
			StoreType: "RETAIL",
		},
		{
			Name:      "Teh Botol",
			Price:     5000,
			Category:  "Beverages",
			ImageURL:  "https://images.unsplash.com/photo-1556679343-c7306c1976bc",
			StoreType: "RETAIL",
		},
		{
			Name:      "Sabun Mandi",
			Price:     8000,
			Category:  "Personal Care",
			ImageURL:  "https://images.unsplash.com/photo-1556228578-0d85b1a4d571",
			StoreType: "RETAIL",
		},
		{
			Name:      "Pembersih Lantai",
			Price:     15000,
			Category:  "Household",
			ImageURL:  "https://images.unsplash.com/photo-1585421514284-efb74c2b69ba",
			StoreType: "RETAIL",
		},
		{
			Name:      "Nasi Goreng",
			Price:     25000,
			Category:  "Food",
			ImageURL:  "https://images.unsplash.com/photo-1512058564366-18510be2db19",
			StoreType: "FNB",
		},
		{
			Name:      "Es Teh Manis",
			Price:     8000,
			Category:  "Beverages",
			ImageURL:  "https://images.unsplash.com/photo-1556679343-c7306c1976bc",
			StoreType: "FNB",
		},
		{
			Name:      "Pisang Goreng",
			Price:     12000,
			Category:  "Desserts",
			ImageURL:  "https://images.unsplash.com/photo-1528975604071-b4dc52a2d18c",
			StoreType: "FNB",
		},
		{
			Name:      "Kaos Polos",
			Price:     75000,
			Category:  "Tops",
			ImageURL:  "https://images.unsplash.com/photo-1521572163474-6864f9cf17ab",
			StoreType: "OUTFIT",
		},
		{
			Name:      "Celana Jeans",
			Price:     250000,
			Category:  "Bottoms",
			ImageURL:  "https://images.unsplash.com/photo-1542272604-787c3835535d",
			StoreType: "OUTFIT",
		},
		{
			Name:      "Jaket Hoodie",
			Price:     185000,
			Category:  "Outerwear",
			ImageURL:  "https://images.unsplash.com/photo-1556821840-3a63f95609a7",
			StoreType: "OUTFIT",
		},
	}
	for _, p := range products {
		db.Create(&p)
	}
}