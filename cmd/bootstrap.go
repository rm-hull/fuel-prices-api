package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/rm-hull/godx"

	"github.com/rm-hull/fuel-prices-api/internal"
)

// bootstrap initialises shared resources used by both the API server and import
// commands. It returns the authenticated client, a repository, and an error
// if something failed during startup.
func bootstrap(dbPath string) (internal.FuelPricesClient, internal.FuelPricesRepository, error) {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	godx.GitVersion()
	godx.EnvironmentVars()
	godx.UserInfo()

	clientId := os.Getenv("CLIENT_ID")
	clientSecret := os.Getenv("CLIENT_SECRET")

	client, err := internal.NewFuelPricesClient(clientId, clientSecret)
	if err != nil {
		return nil, nil, fmt.Errorf("GOV.UK authentication failed: %w", err)
	}

	db, err := internal.Connect(dbPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	if err := internal.Migrate("migrations", dbPath); err != nil {
		_ = db.Close()
		return nil, nil, fmt.Errorf("failed to migrate SQL: %w", err)
	}

	repo := internal.NewFuelPricesRepository(db)

	return client, repo, nil
}
