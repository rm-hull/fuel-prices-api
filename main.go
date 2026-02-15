package main

import (
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/rm-hull/fuel-prices-api/internal"

	_ "github.com/mattn/go-sqlite3"
)

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	db, err := internal.Connect("data/fuel_prices.db")
	if err != nil {
		log.Fatalf("Database connection failed: %v", err)
		return
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Fatalf("Failed to close database: %v", err)
		}
	}()
	repo := internal.NewFuelPricesRepository(db)

	clientId := os.Getenv("CLIENT_ID")
	clientSecret := os.Getenv("CLIENT_SECRET")

	client, err := internal.NewFuelPricesClient(clientId, clientSecret)
	if err != nil {
		log.Fatalf("Authentication failed: %v\n", err)
	}

	count := 0
	for {
		if count%4 == 0 {
			numPFS, err := client.GetFillingStations(repo.InsertPFS)
			if err != nil {
				log.Fatalf("Error fetching PFS: %v\n", err)
			}
			log.Printf("Inserted %d PFS", numPFS)
		}

		numPrices, err := client.GetFuelPrices(repo.InsertPrices)
		if err != nil {
			log.Fatalf("Error fetching fuel prices: %v\n", err)
		}
		log.Printf("Inserted %d price records", numPrices)

		time.Sleep(15 * time.Minute)
		count++
	}
	// results, err := repo.Search([]float64{-1.6237449645996096, 53.945882632598945, -1.4258193969726562, 54.03288059902232})
	// if err != nil {
	// 	log.Fatalf("Error searching: %v", err)
	// }
	// log.Printf("Found %d results", len(results))
	// jsonBytes, err := json.MarshalIndent(results, "", "  ")
	// if err != nil {
	// 	log.Fatalf("Error marshalling results: %v", err)
	// }
	// log.Printf("Results:\n%s", string(jsonBytes))
}
