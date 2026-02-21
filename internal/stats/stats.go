package stats

import (
	"fmt"
	"math"

	"github.com/rm-hull/fuel-prices-api/internal/models"
)

func Derive(results []models.SearchResult, bucketSize int) *models.SearchStatistics {
	if bucketSize <= 0 {
		bucketSize = 3
	}
	stats := &models.SearchStatistics{
		CheapestPfs:       make(map[string][]string),
		LowestPrice:       make(map[string]float64),
		AveragePrice:      make(map[string]float64),
		HighestPrice:      make(map[string]float64),
		PriceDistribution: make(map[string]map[string]int),
		StandardDeviation: make(map[string]float64),
		BrandDistribution: make(map[string]int),
	}

	// Group prices by fuel type
	fuelTypePrices := make(map[string][]float64)
	fuelTypeStations := make(map[string]map[float64][]string) // price -> station names

	for _, result := range results {
		for fuelType, priceInfos := range result.FuelPrices {
			if len(priceInfos) == 0 {
				continue
			}

			// Use the most recent price (first in slice)
			price := priceInfos[0].Price
			fuelTypePrices[fuelType] = append(fuelTypePrices[fuelType], price)

			if fuelTypeStations[fuelType] == nil {
				fuelTypeStations[fuelType] = make(map[float64][]string)
			}
			fuelTypeStations[fuelType][price] = append(fuelTypeStations[fuelType][price], result.NodeId)
		}
	}

	for fuelType, prices := range fuelTypePrices {
		if len(prices) == 0 {
			continue
		}

		// Lowest/avg/hightest price and cheapest stations
		lowestPrice := prices[0]
		highestPrice := prices[0]
		sum := 0.0

		for _, p := range prices {
			if p < lowestPrice {
				lowestPrice = p
			}
			if p > highestPrice {
				highestPrice = p
			}
			sum += p
		}
		stats.LowestPrice[fuelType] = lowestPrice
		stats.HighestPrice[fuelType] = highestPrice
		stats.CheapestPfs[fuelType] = fuelTypeStations[fuelType][lowestPrice]

		avgPrice := sum / float64(len(prices))
		stats.AveragePrice[fuelType] = math.Round(avgPrice*10) / 10

		// Standard deviation
		if len(prices) > 1 {
			variance := 0.0
			for _, p := range prices {
				variance += math.Pow(p-avgPrice, 2)
			}
			variance /= float64(len(prices))
			stats.StandardDeviation[fuelType] = math.Sqrt(variance)
		}

		stats.PriceDistribution[fuelType] = make(map[string]int)
		for _, p := range prices {
			price := int(p)
			bucketStart := (price / bucketSize) * bucketSize
			bucketEnd := bucketStart + bucketSize - 1
			bucketKey := fmt.Sprintf("%d-%d", bucketStart, bucketEnd)
			stats.PriceDistribution[fuelType][bucketKey]++
		}
	}

	// Brand distribution - count results by retailer
	for _, result := range results {
		if result.Retailer != nil {
			stats.BrandDistribution[result.Retailer.Name]++
		}
	}

	return stats
}
