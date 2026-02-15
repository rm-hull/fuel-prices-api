package models

import "time"

type PriceInfo struct {
	Price     float64   `json:"price"`
	UpdatedOn time.Time `json:"updated_on"`
}

type SearchResult struct {
	PetrolFillingStation
	FuelPrices map[string][]PriceInfo `json:"fuel_prices,omitempty"`
}
