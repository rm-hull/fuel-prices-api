package cmd

import (
	"fmt"
	"log"
)

func Import(dbPath string) error {

	client, repo, err := bootstrap(dbPath)
	if err != nil {
		return err
	}
	defer func() {
		if err := repo.Close(); err != nil {
			log.Printf("failed to close repository: %v", err)
		}
	}()

	numPFS, err := client.GetFillingStations(repo.InsertPFS)
	if err != nil {
		return fmt.Errorf("failed to fetch filling stations: %w", err)
	}
	log.Printf("imported %d filling stations", numPFS)

	numPrices, err := client.GetFuelPrices(repo.InsertPrices)
	if err != nil {
		return fmt.Errorf("failed to fetch fuel prices: %w", err)
	}
	log.Printf("imported %d fuel prices", numPrices)

	return nil
}
