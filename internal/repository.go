package internal

import (
	"database/sql"
	_ "embed"
	"fmt"
	"log"
	"sync"

	"github.com/rm-hull/fuel-prices-api/internal/models"
	"github.com/tavsec/gin-healthcheck/checks"
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
	InsertPFS(batch []models.PetrolFillingStation) (int, error)
	InsertPrices(batch []models.ForecourtPrices) (int, error)
	Search(boundingBox []float64, perTypeLimit int) ([]models.SearchResult, error)
	Close() error
	Check() checks.Check
}

type sqliteRepository struct {
	db *sql.DB
}

func NewFuelPricesRepository(db *sql.DB) FuelPricesRepository {
	return &sqliteRepository{
		db: db,
	}
}

func (repo *sqliteRepository) Close() error {
	return repo.db.Close()
}

func (repo *sqliteRepository) Check() checks.Check {
	return checks.SqlCheck{Sql: repo.db}
}

func (repo *sqliteRepository) InsertPFS(batch []models.PetrolFillingStation) (int, error) {
	if len(batch) == 0 {
		return 0, nil
	}

	tx, err := repo.db.Begin()
	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction: %w", err)
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
		return 0, fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer func() {
		if err := stmt.Close(); err != nil {
			log.Printf("failed to close statement: %v", err)
		}
	}()

	count := 0
	for _, pfs := range batch {
		_, err = stmt.Exec(pfs.ToTuple()...)
		if err != nil {
			return 0, fmt.Errorf("failed to execute individual insert: %w", err)
		}
		count++
	}

	if err = tx.Commit(); err != nil {
		return 0, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return count, nil
}

func (repo *sqliteRepository) InsertPrices(batch []models.ForecourtPrices) (int, error) {
	if len(batch) == 0 {
		return 0, nil
	}

	tx, err := repo.db.Begin()
	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction: %w", err)
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
		return 0, fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer func() {
		if err := stmt.Close(); err != nil {
			log.Printf("failed to close statement: %v", err)
		}
	}()

	count := 0
	for _, forecourtPrices := range batch {
		for _, fuelPrice := range forecourtPrices.FuelPrices {
			_, err = stmt.Exec(fuelPrice.ToTuple(forecourtPrices.NodeId)...)
			if err != nil {
				return 0, fmt.Errorf("failed to execute individual insert: %w", err)
			}
			count++
		}
	}

	if err = tx.Commit(); err != nil {
		return 0, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return count, nil
}

func (repo *sqliteRepository) Search(boundingBox []float64, perTypeLimit int) ([]models.SearchResult, error) {
	var (
		pfs       []models.SearchResult
		prices    map[string]map[string][]models.PriceInfo
		pfsErr    error
		pricesErr error
		wg        sync.WaitGroup
	)

	wg.Add(2)
	go repo.fetchPfs(boundingBox, &pfs, &pfsErr, wg.Done)
	go repo.fetchPrices(boundingBox, &prices, &pricesErr, perTypeLimit, wg.Done)

	wg.Wait()

	if pfsErr != nil {
		return nil, pfsErr
	}
	if pricesErr != nil {
		return nil, pricesErr
	}

	for i := range pfs {
		nodeId := pfs[i].NodeId
		if fuelPrices, ok := prices[nodeId]; ok {
			pfs[i].FuelPrices = fuelPrices
		}
	}

	return pfs, nil
}

func (repo *sqliteRepository) fetchPfs(boundingBox []float64, results *[]models.SearchResult, err *error, done func()) {
	defer done()

	rows, queryErr := repo.db.Query(searchPfsSQL, boundingBox[1], boundingBox[3], boundingBox[0], boundingBox[2])
	if queryErr != nil {
		*err = fmt.Errorf("failed to execute search query: %w", queryErr)
		return
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			log.Printf("failed to close rows: %v", closeErr)
		}
	}()

	for rows.Next() {
		var result models.SearchResult
		var openingTimesJSON, amenitiesJSON, fuelTypesJSON string
		if scanErr := rows.Scan(
			&result.NodeId, &result.MftOrganisationName, &result.PublicPhoneNumber, &result.TradingName,
			&result.IsSameTradingAndBrandName, &result.BrandName, &result.TemporaryClosure,
			&result.PermanentClosure, &result.PermanentClosureDate, &result.IsMotorwayServiceStation,
			&result.IsSupermarketServiceStation,
			&result.Location.AddressLine1, &result.Location.AddressLine2, &result.Location.City, &result.Location.Country,
			&result.Location.County, &result.Location.Postcode, &result.Location.Latitude, &result.Location.Longitude,
			&openingTimesJSON, &amenitiesJSON, &fuelTypesJSON,
		); scanErr != nil {
			*err = fmt.Errorf("failed to scan row: %w", scanErr)
			return
		}
		if unmarshalErr := json.Unmarshal([]byte(openingTimesJSON), &result.OpeningTimes); unmarshalErr != nil {
			*err = fmt.Errorf("failed to unmarshal opening times: %w", unmarshalErr)
			return
		}
		if unmarshalErr := json.Unmarshal([]byte(amenitiesJSON), &result.Amenities); unmarshalErr != nil {
			*err = fmt.Errorf("failed to unmarshal amenities: %w", unmarshalErr)
			return
		}
		if unmarshalErr := json.Unmarshal([]byte(fuelTypesJSON), &result.FuelTypes); unmarshalErr != nil {
			*err = fmt.Errorf("failed to unmarshal fuel types: %w", unmarshalErr)
			return
		}
		*results = append(*results, result)
	}

	if rowsErr := rows.Err(); rowsErr != nil {
		*err = fmt.Errorf("error iterating over rows: %w", rowsErr)
	}
}

func (repo *sqliteRepository) fetchPrices(boundingBox []float64, results *map[string]map[string][]models.PriceInfo, err *error, perTypeLimit int, done func()) {
	defer done()

	rows, queryErr := repo.db.Query(searchPricesSQL, boundingBox[1], boundingBox[3], boundingBox[0], boundingBox[2], perTypeLimit)
	if queryErr != nil {
		*err = fmt.Errorf("failed to execute search query: %w", queryErr)
		return
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			log.Printf("failed to close rows: %v", closeErr)
		}
	}()

	*results = make(map[string]map[string][]models.PriceInfo)

	for rows.Next() {
		var nodeId string
		var fuelPrice models.FuelPrice
		if scanErr := rows.Scan(
			&nodeId, &fuelPrice.FuelType, &fuelPrice.PriceLastUpdated,
			&fuelPrice.Price, &fuelPrice.PriceChangeEffectiveTimestamp,
		); scanErr != nil {
			*err = fmt.Errorf("failed to scan row: %w", scanErr)
			return
		}
		if _, exists := (*results)[nodeId]; !exists {
			(*results)[nodeId] = make(map[string][]models.PriceInfo)
		}

		(*results)[nodeId][fuelPrice.FuelType] = append((*results)[nodeId][fuelPrice.FuelType], models.PriceInfo{
			Price:         fuelPrice.Price,
			UpdatedOn:     fuelPrice.PriceLastUpdated,
			EffectiveFrom: fuelPrice.PriceChangeEffectiveTimestamp,
		})
	}
}
