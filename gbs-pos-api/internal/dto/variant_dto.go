package dto

type CreateVariantRequest struct {
	SKU                string                 `json:"sku"`
	Name               string                 `json:"name" binding:"required"`
	Attributes         map[string]interface{} `json:"attributes"`
	Price              *float64               `json:"price"`
	StockQuantity      int                    `json:"stockQuantity"`
	LowStockThreshold  *int                   `json:"lowStockThreshold"`
	IsActive           *bool                  `json:"isActive"`
	SortOrder          int                    `json:"sortOrder"`
}

type UpdateVariantRequest struct {
	SKU                *string                `json:"sku"`
	Name                *string                `json:"name"`
	Attributes          map[string]interface{} `json:"attributes"`
	Price               *float64              `json:"price"`
	StockQuantity       *int                  `json:"stockQuantity"`
	LowStockThreshold   *int                  `json:"lowStockThreshold"`
	IsActive            *bool                 `json:"isActive"`
	SortOrder           *int                  `json:"sortOrder"`
}

type VariantResponse struct {
	ID                int                    `json:"id"`
	ProductID         int                    `json:"productId"`
	SKU               string                 `json:"sku"`
	Name              string                 `json:"name"`
	Attributes        map[string]interface{} `json:"attributes"`
	Price             *float64              `json:"price"`
	StockQuantity     int                    `json:"stockQuantity"`
	LowStockThreshold *int                   `json:"lowStockThreshold"`
	IsActive          bool                   `json:"isActive"`
	SortOrder         int                    `json:"sortOrder"`
	CreatedAt         string                 `json:"createdAt"`
	UpdatedAt         string                 `json:"updatedAt"`
}
