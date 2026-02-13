package internal

import (
	"database/sql"
	_ "embed"
	"encoding/json"
	"fmt"
	"log"

	"github.com/rm-hull/fuel-prices-api/internal/models"
)

//go:embed sql/insert_pfs.sql
var insertPfsSQL string

//go:embed sql/insert_prices.sql
var insertPricesSQL string

//go:embed sql/search_pfs.sql
var searchPfsSQL string

//go:embed sql/search_prices.sql
var searchPricesSQL string

type FuelPricesRepository interface {
	InsertPFS(batch []models.PetrolFillingStation) error
	InsertPrices(batch []models.ForecourtPrices) error
	Search(boundingBox []float64) ([]models.SearchResult, error)
}

type sqliteRepository struct {
	db *sql.DB
}

func NewFuelPricesRepository(db *sql.DB) FuelPricesRepository {
	return &sqliteRepository{
		db: db,
	}
}

func (repo *sqliteRepository) InsertPFS(batch []models.PetrolFillingStation) error {
	if len(batch) == 0 {
		return nil
	}

	tx, err := repo.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				log.Printf("error rolling back transaction: %v", rbErr)
			}
		}
	}()

	stmt, err := tx.Prepare(insertPfsSQL)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer func() {
		if err := stmt.Close(); err != nil {
			log.Printf("failed to close statement: %v", err)
		}
	}()

	for _, pfs := range batch {
		_, err = stmt.Exec(pfs.ToTuple()...)
		if err != nil {
			return fmt.Errorf("failed to execute individual insert: %w", err)
		}
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (repo *sqliteRepository) InsertPrices(batch []models.ForecourtPrices) error {
	if len(batch) == 0 {
		return nil
	}

	tx, err := repo.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				log.Printf("error rolling back transaction: %v", rbErr)
			}
		}
	}()

	stmt, err := tx.Prepare(insertPricesSQL)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer func() {
		if err := stmt.Close(); err != nil {
			log.Printf("failed to close statement: %v", err)
		}
	}()

	for _, forecourtPrices := range batch {
		for _, fuelPrice := range forecourtPrices.FuelPrices {
			_, err = stmt.Exec(fuelPrice.ToTuple(forecourtPrices.NodeId)...)
			if err != nil {
				return fmt.Errorf("failed to execute individual insert: %w", err)
			}
		}
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (repo *sqliteRepository) Search(boundingBox []float64) ([]models.SearchResult, error) {

	rows, err := repo.db.Query(searchPfsSQL, boundingBox[1], boundingBox[3], boundingBox[0], boundingBox[2])
	if err != nil {
		return nil, fmt.Errorf("failed to execute search query: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("failed to close rows: %v", err)
		}
	}()

	var results []models.SearchResult
	for rows.Next() {
		var result models.SearchResult
		var openingTimesJSON, amenitiesJSON, fuelTypesJSON string
		if err := rows.Scan(
			&result.NodeId, &result.MftOrganisationName, &result.PublicPhoneNumber, &result.TradingName,
			&result.IsSameTradingAndBrandName, &result.BrandName, &result.TemporaryClosure,
			&result.PermanentClosure, &result.PermanentClosureDate, &result.IsMotorwayServiceStation,
			&result.IsSupermaketServiceStation,
			&result.Location.AddressLine1, &result.Location.AddressLine2, &result.Location.City, &result.Location.Country,
			&result.Location.County, &result.Location.Postcode, &result.Location.Latitude, &result.Location.Longitude,
			&openingTimesJSON, &amenitiesJSON, &fuelTypesJSON,
		); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		if err := json.Unmarshal([]byte(openingTimesJSON), &result.OpeningTimes); err != nil {
			return nil, fmt.Errorf("failed to unmarshal opening times: %w", err)
		}
		if err := json.Unmarshal([]byte(amenitiesJSON), &result.Amenities); err != nil {
			return nil, fmt.Errorf("failed to unmarshal amenities: %w", err)
		}
		if err := json.Unmarshal([]byte(fuelTypesJSON), &result.FuelTypes); err != nil {
			return nil, fmt.Errorf("failed to unmarshal fuel types: %w", err)
		}
		results = append(results, result)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over rows: %w", err)
	}

	return results, nil
}
