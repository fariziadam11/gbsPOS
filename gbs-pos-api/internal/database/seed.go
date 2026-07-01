package database

import (
	"gbs-pos-api/internal/model"
	"log"
	"time"

	"gorm.io/gorm"
)

func Seed(db *gorm.DB) {
	// Fuel master data can be seeded independently even on existing deployments.
	seedFuelData(db)

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
			Name:              "Chitato",
			Price:             11500,
			Category:          "Snacks",
			ImageURL:          "https://images.unsplash.com/photo-1621939514649-28b12e81658b",
			StoreType:         "RETAIL",
			StockQuantity:     100,
			LowStockThreshold: 10,
		},
		{
			Name:              "Indomie Goreng",
			Price:             3500,
			Category:          "Snacks",
			ImageURL:          "https://images.unsplash.com/photo-1612929633738-8fe44f7ec841",
			StoreType:         "RETAIL",
			StockQuantity:     200,
			LowStockThreshold: 20,
		},
		{
			Name:              "Teh Botol",
			Price:             5000,
			Category:          "Beverages",
			ImageURL:          "https://images.unsplash.com/photo-1556679343-c7306c1976bc",
			StoreType:         "RETAIL",
			StockQuantity:     150,
			LowStockThreshold: 15,
		},
		{
			Name:              "Sabun Mandi",
			Price:             8000,
			Category:          "Personal Care",
			ImageURL:          "https://images.unsplash.com/photo-1556228578-0d85b1a4d571",
			StoreType:         "RETAIL",
			StockQuantity:     80,
			LowStockThreshold: 8,
		},
		{
			Name:              "Pembersih Lantai",
			Price:             15000,
			Category:          "Household",
			ImageURL:          "https://images.unsplash.com/photo-1585421514284-efb74c2b69ba",
			StoreType:         "RETAIL",
			StockQuantity:     60,
			LowStockThreshold: 6,
		},
		{
			Name:              "Nasi Goreng",
			Price:             25000,
			Category:          "Food",
			ImageURL:          "https://images.unsplash.com/photo-1512058564366-18510be2db19",
			StoreType:         "FNB",
			StockQuantity:     50,
			LowStockThreshold: 5,
		},
		{
			Name:              "Es Teh Manis",
			Price:             8000,
			Category:          "Beverages",
			ImageURL:          "https://images.unsplash.com/photo-1556679343-c7306c1976bc",
			StoreType:         "FNB",
			StockQuantity:     100,
			LowStockThreshold: 10,
		},
		{
			Name:              "Pisang Goreng",
			Price:             12000,
			Category:          "Desserts",
			ImageURL:          "https://images.unsplash.com/photo-1528975604071-b4dc52a2d18c",
			StoreType:         "FNB",
			StockQuantity:     40,
			LowStockThreshold: 5,
		},
		{
			Name:              "Kaos Polos",
			Price:             75000,
			Category:          "Tops",
			ImageURL:          "https://images.unsplash.com/photo-1521572163474-6864f9cf17ab",
			StoreType:         "OUTFIT",
			StockQuantity:     30,
			LowStockThreshold: 5,
		},
		{
			Name:              "Celana Jeans",
			Price:             250000,
			Category:          "Bottoms",
			ImageURL:          "https://images.unsplash.com/photo-1542272604-787c3835535d",
			StoreType:         "OUTFIT",
			StockQuantity:     20,
			LowStockThreshold: 3,
		},
		{
			Name:              "Jaket Hoodie",
			Price:             185000,
			Category:          "Outerwear",
			ImageURL:          "https://images.unsplash.com/photo-1556821840-3a63f95609a7",
			StoreType:         "OUTFIT",
			StockQuantity:     25,
			LowStockThreshold: 3,
		},
	}
	for _, p := range products {
		db.Create(&p)
	}
}

func seedFuelData(db *gorm.DB) {
	var fuelPriceCount int64
	db.Model(&model.FuelPrice{}).Count(&fuelPriceCount)
	if fuelPriceCount == 0 {
		prices := []model.FuelPrice{
			{Code: "PERTALITE", Name: "Pertalite", PricePerLiter: 10000, UpdatedAt: time.Now()},
			{Code: "PERTAMAX", Name: "Pertamax", PricePerLiter: 12500, UpdatedAt: time.Now()},
			{Code: "PERTAMAX_GREEN", Name: "Pertamax Green", PricePerLiter: 13000, UpdatedAt: time.Now()},
			{Code: "PERTAMAX_TURBO", Name: "Pertamax Turbo", PricePerLiter: 14000, UpdatedAt: time.Now()},
			{Code: "PERTAMINA_DEX", Name: "Pertamina Dex", PricePerLiter: 15000, UpdatedAt: time.Now()},
			{Code: "PERTAMINA_DEXLITE", Name: "Pertamina Dexlite", PricePerLiter: 14500, UpdatedAt: time.Now()},
		}
		for _, p := range prices {
			db.Create(&p)
		}
		log.Println("Seeded fuel prices")
	}

	var pumpCount int64
	db.Model(&model.Pump{}).Count(&pumpCount)
	if pumpCount == 0 {
		pumps := []model.Pump{
			{ID: "P01", Name: "Pompa 1", IsActive: true},
			{ID: "P02", Name: "Pompa 2", IsActive: true},
			{ID: "P03", Name: "Pompa 3", IsActive: true},
		}
		nozzles := []model.Nozzle{
			{ID: "P01N01", PumpID: "P01", Name: "Nozzle 1", FuelCode: "PERTALITE", IsActive: true},
			{ID: "P01N02", PumpID: "P01", Name: "Nozzle 2", FuelCode: "PERTAMAX", IsActive: true},
			{ID: "P02N01", PumpID: "P02", Name: "Nozzle 1", FuelCode: "PERTAMAX_GREEN", IsActive: true},
			{ID: "P02N02", PumpID: "P02", Name: "Nozzle 2", FuelCode: "PERTAMAX_TURBO", IsActive: true},
			{ID: "P03N01", PumpID: "P03", Name: "Nozzle 1", FuelCode: "PERTAMINA_DEX", IsActive: true},
			{ID: "P03N02", PumpID: "P03", Name: "Nozzle 2", FuelCode: "PERTAMINA_DEXLITE", IsActive: true},
		}
		for _, p := range pumps {
			db.Create(&p)
		}
		for _, n := range nozzles {
			db.Create(&n)
		}
		log.Println("Seeded pumps and nozzles")
	}
}