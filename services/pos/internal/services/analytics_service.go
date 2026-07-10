package services

import (
	"context"
	"sort"
	"time"

	"github.com/larisai/pos-service/internal/models/dto"
	"github.com/larisai/pos-service/internal/models/entity"
	"github.com/larisai/pos-service/internal/repositories"
)

type AnalyticsService interface {
	GetSummary(ctx context.Context) (*dto.AnalyticsSummaryResponse, error)
}

type analyticsService struct {
	analyticsRepo repositories.AnalyticsRepository
}

func NewAnalyticsService(analyticsRepo repositories.AnalyticsRepository) AnalyticsService {
	return &analyticsService{
		analyticsRepo: analyticsRepo,
	}
}

func (s *analyticsService) GetSummary(ctx context.Context) (*dto.AnalyticsSummaryResponse, error) {
	txs, err := s.analyticsRepo.GetValidTransactions(ctx)
	if err != nil {
		return nil, err
	}

	prods, err := s.analyticsRepo.GetActiveProducts(ctx)
	if err != nil {
		return nil, err
	}

	// Map produk untuk lookup kategori dan barcode
	prodMap := make(map[string]entity.Product)
	for _, p := range prods {
		prodMap[p.ID.Hex()] = p
	}

	var totalRevenue int64
	totalOrders := len(txs)
	paymentMap := make(map[string]struct {
		Amount int64
		Count  int
	})
	productSalesMap := make(map[string]*dto.TopSellingProduct)
	dailyMap := make(map[string]*dto.DailySale)

	for _, tx := range txs {
		totalRevenue += tx.TotalAmount

		// Agregasi metode pembayaran
		pm := tx.PaymentType
		if pm == "" {
			pm = "TUNAI"
		}
		currPm := paymentMap[pm]
		currPm.Amount += tx.TotalAmount
		currPm.Count++
		paymentMap[pm] = currPm

		// Agregasi daily sales
		dateStr := tx.CreatedAt.Format("2006-01-02")
		if ds, exists := dailyMap[dateStr]; exists {
			ds.TotalAmount += tx.TotalAmount
			ds.OrderCount++
		} else {
			dailyMap[dateStr] = &dto.DailySale{
				Date:        dateStr,
				TotalAmount: tx.TotalAmount,
				OrderCount:  1,
			}
		}

		// Agregasi top products
		for _, item := range tx.Items {
			pid := item.ProductID.Hex()
			if tp, exists := productSalesMap[pid]; exists {
				tp.TotalSold += item.Quantity
				tp.TotalAmount += item.Subtotal
			} else {
				category := "Umum"
				barcode := item.Barcode
				name := item.Name
				if p, ok := prodMap[pid]; ok {
					category = p.Category
					if barcode == "" {
						barcode = p.Barcode
					}
					if name == "" {
						name = p.Name
					}
				}
				productSalesMap[pid] = &dto.TopSellingProduct{
					ProductID:   pid,
					Barcode:     barcode,
					Name:        name,
					Category:    category,
					TotalSold:   item.Quantity,
					TotalAmount: item.Subtotal,
				}
			}
		}
	}

	var avgOrderValue int64
	if totalOrders > 0 {
		avgOrderValue = totalRevenue / int64(totalOrders)
	}

	// Format payment methods
	var paymentMethods []dto.PaymentMethodBreakdown
	for pm, data := range paymentMap {
		pct := 0.0
		if totalRevenue > 0 {
			pct = (float64(data.Amount) / float64(totalRevenue)) * 100.0
		}
		paymentMethods = append(paymentMethods, dto.PaymentMethodBreakdown{
			PaymentType: pm,
			TotalAmount: data.Amount,
			Count:       data.Count,
			Percentage:  pct,
		})
	}
	sort.Slice(paymentMethods, func(i, j int) bool {
		return paymentMethods[i].TotalAmount > paymentMethods[j].TotalAmount
	})

	// Format top products
	var topProducts []dto.TopSellingProduct
	for _, tp := range productSalesMap {
		topProducts = append(topProducts, *tp)
	}
	sort.Slice(topProducts, func(i, j int) bool {
		if topProducts[i].TotalSold == topProducts[j].TotalSold {
			return topProducts[i].TotalAmount > topProducts[j].TotalAmount
		}
		return topProducts[i].TotalSold > topProducts[j].TotalSold
	})
	if len(topProducts) > 5 {
		topProducts = topProducts[:5]
	}

	// Format daily sales
	var dailySales []dto.DailySale
	for _, ds := range dailyMap {
		dailySales = append(dailySales, *ds)
	}
	sort.Slice(dailySales, func(i, j int) bool {
		return dailySales[i].Date < dailySales[j].Date
	})

	// Pastikan minimal ada 7 hari terakhir dalam dailySales jika data kurang
	if len(dailySales) < 7 {
		today := time.Now()
		dateExists := make(map[string]bool)
		for _, ds := range dailySales {
			dateExists[ds.Date] = true
		}
		for i := 6; i >= 0; i-- {
			d := today.AddDate(0, 0, -i).Format("2006-01-02")
			if !dateExists[d] {
				dailySales = append(dailySales, dto.DailySale{
					Date:        d,
					TotalAmount: 0,
					OrderCount:  0,
				})
			}
		}
		sort.Slice(dailySales, func(i, j int) bool {
			return dailySales[i].Date < dailySales[j].Date
		})
	}

	return &dto.AnalyticsSummaryResponse{
		TotalRevenue:      totalRevenue,
		TotalOrders:       totalOrders,
		AverageOrderValue: avgOrderValue,
		PaymentMethods:    paymentMethods,
		TopProducts:       topProducts,
		DailySales:        dailySales,
	}, nil
}
