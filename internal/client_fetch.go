package internal

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/rm-hull/fuel-prices-api/internal/models"
	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

type FuelPricesClient interface {
	Authenticate() error
	GetAllFuelPrices(func([]models.ForecourtPrices) error) (int, error)
	GetFillingStations(func([]models.PetrolFillingStation) error) (int, error)
}

type fuelPricesManager struct {
	baseUrl   string
	authReq   models.AuthRequest
	tokenData models.TokenData
	startTime time.Time
	client    *http.Client
}

func NewFuelPricesClient(clientId, clientSecret string) *fuelPricesManager {
	return &fuelPricesManager{
		baseUrl:   "https://www.fuel-finder.service.gov.uk/api/v1",
		startTime: time.Now(),
		client:    &http.Client{},
		authReq: models.AuthRequest{
			ClientId:     clientId,
			ClientSecret: clientSecret,
		},
	}
}

func (mgr *fuelPricesManager) Authenticate() error {
	// http call to https://www.fuel-finder.service.gov.uk/api/v1/oauth/generate_access_token

	url := fmt.Sprintf("%s/oauth/generate_access_token", mgr.baseUrl)
	body, err := mgr.post(url, "application/json", mgr.authReq)
	if err != nil {
		return err
	}
	defer func() {
		if err := body.Close(); err != nil {
			log.Printf("failed to close body: %v", err)
		}
	}()

	var resp models.AuthResponse
	decoder := json.NewDecoder(body)
	if err := decoder.Decode(&resp); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}
	if !resp.Success {
		return fmt.Errorf("authentication failed: %s", resp.Message)
	}

	mgr.tokenData = resp.Data

	return nil
}

func (mgr *fuelPricesManager) GetAllFuelPrices(callback func([]models.ForecourtPrices) error) (int, error) {
	// http call to https://www.fuel-finder.service.gov.uk/api/v1/fuelprices
	batchNo := 1
	count := 0
	for {
		url := fmt.Sprintf("%s/pfs/fuel-prices?batch-number=%d", mgr.baseUrl, batchNo)
		body, err := mgr.get(url)
		if err != nil {
			return 0, err
		}
		defer func() {
			if err := body.Close(); err != nil {
				log.Printf("failed to close body: %v", err)
			}
		}()

		bodyBytes, err := io.ReadAll(body)
		if err != nil {
			return 0, fmt.Errorf("failed to read response body: %w", err)
		}

		var resp models.ForecourtPricesResponse
		if err := json.Unmarshal(bodyBytes, &resp); err != nil {
			log.Printf("Response body: %s", string(bodyBytes))
			return 0, fmt.Errorf("failed to unmarshal response: %w", err)
		}

		if !resp.Success {
			return 0, fmt.Errorf("API error: %s", resp.Message)
		}

		if err := callback(resp.Data); err != nil {
			return 0, fmt.Errorf("callback error: %w", err)
		}
		batchNo++
		count += len(resp.Data)

		if batchNo >= resp.MetaData.TotalBatches {
			break
		}
	}
	return count, nil
}

func (mgr *fuelPricesManager) GetFillingStations(callback func([]models.PetrolFillingStation) error) (int, error) {
	// http call to https://www.fuel-finder.service.gov.uk/api/v1/pfs?batch-number=1
	batchNo := 1
	count := 0
	for {
		url := fmt.Sprintf("%s/pfs?batch-number=%d", mgr.baseUrl, batchNo)
		body, err := mgr.get(url)
		if err != nil {
			var stErr *HTTPStatusError
			if errors.As(err, &stErr) && stErr.StatusCode == 400 {
				log.Printf("No more batches available, stopping at batch %d", batchNo-1)
				break
			}
			return 0, err
		}
		defer func() {
			if err := body.Close(); err != nil {
				log.Printf("failed to close body: %v", err)
			}
		}()

		var resp []models.PetrolFillingStation
		decoder := json.NewDecoder(body)
		if err := decoder.Decode(&resp); err != nil {
			return 0, fmt.Errorf("failed to unmarshal response: %w", err)
		}

		if err := callback(resp); err != nil {
			return 0, fmt.Errorf("callback error: %w", err)
		}

		count += len(resp)
		batchNo++

		if len(resp) == 0 {
			break
		}
	}

	return count, nil
}

func (mgr *fuelPricesManager) get(url string) (io.ReadCloser, error) {
	log.Printf("GET %s", url)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+mgr.tokenData.AccessToken)
	req.Header.Set("Accept", "application/json")

	resp, err := mgr.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch from %s: %w", url, err)
	}

	if resp.StatusCode > 299 {
		_ = resp.Body.Close()
		return nil, &HTTPStatusError{URL: url, Status: resp.Status, StatusCode: resp.StatusCode}
	}
	return resp.Body, nil
}

// HTTPStatusError is returned when the remote server responds with a non-2xx status.
type HTTPStatusError struct {
	URL        string
	Status     string
	StatusCode int
}

func (e *HTTPStatusError) Error() string {
	return fmt.Sprintf("http status response from %s: %s", e.URL, e.Status)
}

func (mgr *fuelPricesManager) post(url, contentType string, data any) (io.ReadCloser, error) {
	log.Printf("POST %s", url)
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", contentType)

	resp, err := mgr.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to perform request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return resp.Body, nil
}
