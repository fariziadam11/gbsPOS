package dto

type DashboardSummary struct {
	TotalOrders   int     `json:"totalOrders"`
	TotalRevenue  float64 `json:"totalRevenue"`
	AvgOrderValue float64 `json:"avgOrderValue"`
	CashTotal     float64 `json:"cashTotal"`
	CardTotal     float64 `json:"cardTotal"`
	QrisTotal     float64 `json:"qrisTotal"`
	VoidedCount   int     `json:"voidedCount"`
}

type RevenuePoint struct {
	Date    string  `json:"date"`
	Revenue float64 `json:"revenue"`
	Orders  int     `json:"orders"`
}

type TopProduct struct {
	ProductID   int     `json:"productId"`
	ProductName string  `json:"productName"`
	TotalSold   int     `json:"totalSold"`
	Revenue     float64 `json:"revenue"`
}

type ImportResult struct {
	Success int      `json:"success"`
	Failed  int      `json:"failed"`
	Errors  []string `json:"errors,omitempty"`
}

