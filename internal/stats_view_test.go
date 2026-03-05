package internal

import (
	"database/sql"
	"testing"
	"time"

	"github.com/rm-hull/fuel-prices-api/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDailyFuelPriceStatsView(t *testing.T) {
	// Re-use setup logic from repository_test.go but since it's not exported,
	// we'll either need to copy it or make it accessible.
	// For now, I'll copy the minimal setup.
	repo := setupTestDB(t)
	sqliteRepo := repo.(*sqliteRepository)
	db := sqliteRepo.db

	now := time.Now().UTC().Truncate(time.Second)

	// Stations in different postcode areas
	// Leeds: LS1 1AA, LS2 1BB
	// Manchester: M1 1AA
	stations := []models.PetrolFillingStation{
		{NodeId: "L1", Location: models.Location{Postcode: "LS1 1AA"}},
		{NodeId: "L2", Location: models.Location{Postcode: "LS2 1BB"}},
		{NodeId: "M1", Location: models.Location{Postcode: "M1 1AA"}},
	}
	_, err := repo.InsertPFS(stations)
	require.NoError(t, err)

	prices := []models.ForecourtPrices{
		{
			NodeId: "L1",
			FuelPrices: []models.FuelPrice{
				{FuelType: "E10", Price: 140.0, PriceLastUpdated: now},
				{FuelType: "B7", Price: 150.0, PriceLastUpdated: now},
			},
		},
		{
			NodeId: "L2",
			FuelPrices: []models.FuelPrice{
				// Multiple prices for today, only latest (144.0) should be used
				{FuelType: "E10", Price: 142.0, PriceLastUpdated: now.Add(-1 * time.Hour)},
				{FuelType: "E10", Price: 144.0, PriceLastUpdated: now},
			},
		},
		{
			NodeId: "M1",
			FuelPrices: []models.FuelPrice{
				{FuelType: "E10", Price: 150.0, PriceLastUpdated: now},
			},
		},
	}
	_, err = repo.InsertPrices(prices)
	require.NoError(t, err)

	// Expected stats:
	// National E10: (140.0 + 144.0 + 150.0) / 3 = 144.666...
	// National B7: 150.0
	// Area LS E10: (140.0 + 144.0) / 2 = 142.0
	// Area M E10: 150.0

	type StatsRow struct {
		Scope        string
		PostcodeArea sql.NullString
		FuelType     string
		MinPrice     float64
		AvgPrice     float64
		MaxPrice     float64
		StddevPrice  float64
		SampleSize   int
	}

	rows, err := db.Query("SELECT scope, postcode_area, fuel_type, min_price, avg_price, max_price, stddev_price, sample_size FROM daily_fuel_price_stats")
	require.NoError(t, err)
	defer rows.Close()

	var results []StatsRow
	for rows.Next() {
		var r StatsRow
		err := rows.Scan(&r.Scope, &r.PostcodeArea, &r.FuelType, &r.MaxPrice, &r.AvgPrice, &r.MaxPrice, &r.StddevPrice, &r.SampleSize)
		require.NoError(t, err)
		results = append(results, r)
	}

	assert.NotEmpty(t, results)

	// Verify National E10
	foundNationalE10 := false
	for _, r := range results {
		if r.Scope == "National" && r.FuelType == "E10" {
			foundNationalE10 = true
			assert.InDelta(t, 0.0, r.MinPrice, 0.0001)
			assert.InDelta(t, 144.7, r.AvgPrice, 0.0001)
			assert.InDelta(t, 150.0, r.MaxPrice, 0.0001)
			assert.Equal(t, 3, r.SampleSize)
			// stddev of [140, 144, 150]
			// mean = 144.6666
			// var = ((140-144.666)^2 + (144-144.666)^2 + (150-144.666)^2) / 3
			// var = (21.777 + 0.444 + 28.444) / 3 = 16.888
			// stddev = sqrt(16.888) = 4.1096
			assert.InDelta(t, 4.1096, r.StddevPrice, 0.001)
		}
	}
	assert.True(t, foundNationalE10)

	// Verify LS E10
	foundLSE10 := false
	for _, r := range results {
		if r.Scope == "Postcode Area" && r.PostcodeArea.String == "LS" && r.FuelType == "E10" {
			foundLSE10 = true
			assert.Equal(t, 0.0, r.MinPrice)
			assert.Equal(t, 142.0, r.AvgPrice)
			assert.Equal(t, 144.0, r.MaxPrice)
			assert.Equal(t, 2, r.SampleSize)
			// stddev of [140, 144]
			// mean = 142
			// var = ((140-142)^2 + (144-142)^2) / 2 = (4 + 4) / 2 = 4
			// stddev = sqrt(4) = 2.0
			assert.Equal(t, 2.0, r.StddevPrice)
		}
	}
	assert.True(t, foundLSE10)
}
