package models

import "time"

type PriceInfo struct {
	Price         float64    `json:"price"`
	UpdatedOn     time.Time  `json:"updated_on"`
	EffectiveFrom *time.Time `json:"effective_from,omitempty"`
}

type SearchResult struct {
	PetrolFillingStation
	FuelPrices map[string][]PriceInfo `json:"fuel_prices,omitempty"`
	Retailer   *Retailer              `json:"retailer,omitempty"`
}

type SearchResponse struct {
	Results     []SearchResult    `json:"results"`
	Attribution []string          `json:"attribution"`
	Statistics  *SearchStatistics `json:"statistics,omitempty"`
	LastUpdated *time.Time        `json:"last_updated,omitempty"`
}

type SearchStatistics struct {
	CheapestPfs       map[string][]string       `json:"cheapest_stations,omitempty"`
	LowestPrice       map[string]float64        `json:"lowest_price,omitempty"`
	AveragePrice      map[string]float64        `json:"average_price,omitempty"`
	HighestPrice      map[string]float64        `json:"highest_price,omitempty"`
	StandardDeviation map[string]float64        `json:"standard_deviation,omitempty"`
	PriceDistribution map[string]map[string]int `json:"price_distribution,omitempty"`
	BrandDistribution map[string]int            `json:"brand_distribution,omitempty"`
}

type Snapshot struct {
	Scope             string  `json:"scope"`
	PostcodeArea      *string `json:"postcode_area,omitempty"`
	FuelType          string  `json:"fuel_type"`
	LowestPrice       float64 `json:"lowest_price"`
	AveragePrice      float64 `json:"average_price"`
	HighestPrice      float64 `json:"highest_price"`
	StandardDeviation float64 `json:"standard_deviation"`
	SampleSize        int     `json:"sample_size"`
}

type SnapshotStatistics struct {
	Snapshot    []Snapshot `json:"snapshot,omitempty"`
	LastUpdated *time.Time `json:"last_updated,omitempty"`
}
