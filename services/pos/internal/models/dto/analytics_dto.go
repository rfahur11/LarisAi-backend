package dto

// PaymentMethodBreakdown merepresentasikan total pendapatan dan persentase untuk suatu metode pembayaran.
type PaymentMethodBreakdown struct {
	PaymentType string  `json:"payment_type"`
	TotalAmount int64   `json:"total_amount"`
	Count       int     `json:"count"`
	Percentage  float64 `json:"percentage"`
}

// TopSellingProduct merepresentasikan produk dengan penjualan tertinggi.
type TopSellingProduct struct {
	ProductID   string `json:"product_id"`
	Barcode     string `json:"barcode"`
	Name        string `json:"name"`
	Category    string `json:"category"`
	TotalSold   int    `json:"total_sold"`
	TotalAmount int64  `json:"total_amount"`
}

// DailySale merepresentasikan agregasi pendapatan harian untuk grafik tren.
type DailySale struct {
	Date        string `json:"date"` // Format: YYYY-MM-DD
	TotalAmount int64  `json:"total_amount"`
	OrderCount  int    `json:"order_count"`
}

// AnalyticsSummaryResponse adalah kontrak output ke client untuk data analitik dasbor.
type AnalyticsSummaryResponse struct {
	TotalRevenue      int64                    `json:"total_revenue"`
	TotalOrders       int                      `json:"total_orders"`
	AverageOrderValue int64                    `json:"average_order_value"`
	PaymentMethods    []PaymentMethodBreakdown `json:"payment_methods"`
	TopProducts       []TopSellingProduct      `json:"top_products"`
	DailySales        []DailySale              `json:"daily_sales"`
}
