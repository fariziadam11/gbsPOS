package model

type FuelSalesReport struct {
	Summary    []FuelSalesReportItem
	PumpTotals []FuelSalesPumpReportItem
}

type FuelSalesReportItem struct {
	FuelCode    string  `gorm:"column:fuel_code"`
	Liters      float64 `gorm:"column:liters"`
	TotalAmount float64 `gorm:"column:total_amount"`
}

type FuelSalesPumpReportItem struct {
	PumpID      string  `gorm:"column:pump_id"`
	Liters      float64 `gorm:"column:liters"`
	TotalAmount float64 `gorm:"column:total_amount"`
}
