package models

import (
	// "encoding/json"
	"log"
	"time"

	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

type Location struct {
	AddressLine1 string  `json:"address_line_1"`
	AddressLine2 string  `json:"address_line_2,omitempty"`
	City         string  `json:"city"`
	Country      string  `json:"country"`
	County       string  `json:"county,omitempty"`
	Postcode     string  `json:"postcode"`
	Latitude     float64 `json:"latitude,string"`
	Longitude    float64 `json:"longitude,string"`
}

type DailyOpeningTimes struct {
	Open      string `json:"open"`
	Close     string `json:"close"`
	Is24Hours bool   `json:"is_24_hours"`
}

type PetrolFillingStation struct {
	NodeId                      string     `json:"node_id"`
	MftOrganisationName         string     `json:"mft_organisation_name"`
	PublicPhoneNumber           string     `json:"public_phone_number"`
	TradingName                 string     `json:"trading_name"`
	IsSameTradingAndBrandName   bool       `json:"is_same_trading_and_brand_name"`
	BrandName                   string     `json:"brand_name"`
	TemporaryClosure            bool       `json:"temporary_closure"`
	PermanentClosure            bool       `json:"permanent_closure"`
	PermanentClosureDate        *time.Time `json:"permanent_closure_date"`
	IsMotorwayServiceStation    bool       `json:"is_motorway_service_station"`
	IsSupermarketServiceStation bool       `json:"is_supermarket_service_station"`
	Location                    Location   `json:"location"`
	Amenities                   []string   `json:"amenities"`
	OpeningTimes                struct {
		UsualDays   map[string]DailyOpeningTimes `json:"usual_days"`
		BankHoliday struct {
			Type      string `json:"type"`
			OpenTime  string `json:"open_time"`
			CloseTime string `json:"close_time"`
			Is24Hours bool   `json:"is_24_hours"`
		} `json:"bank_holiday"`
	} `json:"opening_times"`
	FuelTypes []string `json:"fuel_types"`
}

type FuelPrice struct {
	FuelType         string    `json:"fuel_type"`
	Price            float64   `json:"price"`
	PriceLastUpdated time.Time `json:"price_last_updated"`
}

type ForecourtPrices struct {
	NodeId              string      `json:"node_id"`
	MftOrganisationName string      `json:"mft_organisation_name"`
	PublicPhoneNumber   string      `json:"public_phone_number"`
	TradingName         string      `json:"trading_name"`
	FuelPrices          []FuelPrice `json:"fuel_prices"`
}

type MetaData struct {
	BatchNumber  int  `json:"batch_number"`
	BatchSize    int  `json:"batch_size"`
	TotalBatches int  `json:"total_batches"`
	Cached       bool `json:"cached"`
}

type ForecourtPricesResponse struct {
	Success  bool              `json:"success"`
	Data     []ForecourtPrices `json:"data"`
	Message  string            `json:"message,omitempty"`
	MetaData MetaData          `json:"metadata"`
}

func (pfs *PetrolFillingStation) ToTuple() []any {

	return []any{
		pfs.NodeId,
		pfs.MftOrganisationName,
		pfs.PublicPhoneNumber,
		pfs.TradingName,
		pfs.IsSameTradingAndBrandName,
		pfs.BrandName,
		pfs.TemporaryClosure,
		pfs.PermanentClosure,
		pfs.PermanentClosureDate,
		pfs.IsMotorwayServiceStation,
		pfs.IsSupermarketServiceStation,
		pfs.Location.AddressLine1,
		pfs.Location.AddressLine2,
		pfs.Location.City,
		pfs.Location.Country,
		pfs.Location.County,
		pfs.Location.Postcode,
		pfs.Location.Latitude,
		pfs.Location.Longitude,
		toJSON(pfs.OpeningTimes),
		toJSON(pfs.Amenities),
		toJSON(pfs.FuelTypes),
	}
}

func (fp *FuelPrice) ToTuple(nodeId string) []any {

	return []any{
		nodeId,
		fp.FuelType,
		fp.PriceLastUpdated,
		fp.Price,
	}
}

func toJSON(v any) string {
	jsonBytes, err := json.Marshal(v)
	if err != nil {
		log.Fatalf("Error marshaling to JSON: %v", err)
	}
	return string(jsonBytes)
}
