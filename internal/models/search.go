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
}

type SearchResponse struct {
	Results     []SearchResult `json:"results"`
	Attribution []string       `json:"attribution"`
	LastUpdated *time.Time     `json:"last_updated,omitempty"`
}
