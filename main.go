package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/rm-hull/fuel-prices-api/internal"

	_ "github.com/mattn/go-sqlite3"
)


func main() {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	clientId := os.Getenv("CLIENT_ID")
	clientSecret := os.Getenv("CLIENT_SECRET")

	client := internal.NewFuelPricesClient(clientId, clientSecret)

	if err := client.Authenticate(); err != nil {
		fmt.Printf("Authentication failed: %v\n", err)
		return
	}

	// Example usage of GetAllFuelPrices
	// _, err := client.GetAllFuelPrices(func(forecourts []models.ForecourtPrices) {
	// 	for _, fc := range forecourts {
	// 		for _, fuelPrice := range fc.FuelPrices {
	// 			fmt.Printf("Node: %s, Fuel Type: %s, Price: %.2f, Last Updated: %s\n",
	// 				fc.NodeId, fuelPrice.FuelType, fuelPrice.Price, fuelPrice.PriceLastUpdated.Format("2006-01-02 15:04:05"))
	// 		}
	// 	}
	// })

	db, err := internal.Connect("data/fuel_prices.db")
	if err != nil {
		fmt.Printf("Database connection failed: %v\n", err)
		return
	}
	defer func() {
		if err := db.Close(); err != nil {
			fmt.Printf("Failed to close database: %v\n", err)
		}
	}()

	err = internal.CreateDB(db)
	if err != nil {
		fmt.Printf("Database creation failed: %v\n", err)
		return
	}

	repo := internal.NewFuelPricesRepository(db)

	// numPFS, err := client.GetFillingStations(func(stations []models.PetrolFillingStation) error {
	// 	return repo.InsertPFS(stations)
	// })
	// if err != nil {
	// 	fmt.Printf("Error fetching fuel prices: %v\n", err)
	// }
	// log.Printf("Inserted %d PFS", numPFS)

	// numPrices, err := client.GetAllFuelPrices(func(prices []models.ForecourtPrices) error {
	// 	return repo.InsertPrices(prices)
	// })
	// if err != nil {
	// 	fmt.Printf("Error fetching fuel prices: %v\n", err)
	// }
	// log.Printf("Inserted %d price records", numPrices)

	results, err := repo.Search([]float64{-1.6237449645996096, 53.945882632598945, -1.4258193969726562, 54.03288059902232})
	if err != nil {
		fmt.Printf("Error searching: %v\n", err)
		return
	}
	log.Printf("Found %d results", len(results))
}
