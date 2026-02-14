package internal

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	neturl "net/url"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/rm-hull/fuel-prices-api/internal/models"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

// HTTPStatusError is returned when the remote server responds with a non-2xx status.
type HTTPStatusError struct {
	URL        string
	Status     string
	StatusCode int
}

func (e *HTTPStatusError) Error() string {
	return fmt.Sprintf("http status response from %s: %s", e.URL, e.Status)
}

type BatchCallback[T any] func([]T) (int, error)

type FuelPricesClient interface {
	GetFuelPrices(BatchCallback[models.ForecourtPrices]) (int, error)
	GetFillingStations(BatchCallback[models.PetrolFillingStation]) (int, error)
}

type timeTracker struct {
	started         time.Time
	lastAuth        time.Time
	lastPfsFetch    time.Time
	lastPricesFetch time.Time
}

type fuelPricesManager struct {
	baseUrl     string
	authReq     models.AuthRequest
	tokenData   models.TokenData
	timeTracker timeTracker
	startTime   time.Time
	client      *http.Client
}

func NewFuelPricesClient(clientId, clientSecret string) (FuelPricesClient, error) {
	mgr := &fuelPricesManager{
		baseUrl:   "https://www.fuel-finder.service.gov.uk/api/v1",
		startTime: time.Now(),
		timeTracker: timeTracker{
			started: time.Now(),
		},
		client: &http.Client{},
		authReq: models.AuthRequest{
			ClientId:     clientId,
			ClientSecret: clientSecret,
		},
	}

	err := mgr.authenticate()
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate: %v", err)
	}

	return mgr, nil
}

func (mgr *fuelPricesManager) GetFuelPrices(callback BatchCallback[models.ForecourtPrices]) (int, error) {
	decode := func(body io.ReadCloser, batchNo int) ([]models.ForecourtPrices, int, error) {
		bodyBytes, err := io.ReadAll(body)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to read response body: %w", err)
		}

		var resp models.ForecourtPricesResponse
		if bodyBytes[0] == '[' {
			var wtf []models.ForecourtPrices
			if err := json.Unmarshal(bodyBytes, &wtf); err != nil {
				return nil, 0, fmt.Errorf("failed to unmarshal response: %w", err)
			}
			resp.Success = true
			resp.Data = wtf
			resp.MetaData = models.MetaData{
				TotalBatches: batchNo + 2,
				BatchNumber:  batchNo,
				BatchSize:    len(wtf),
			}
			log.Printf("WARNING: API returned an array instead of the expected object, treating as a single batch with %d records", len(wtf))
		} else {
			if err := json.Unmarshal(bodyBytes, &resp); err != nil {
				return nil, 0, fmt.Errorf("failed to unmarshal response: %w", err)
			}
		}

		if !resp.Success {
			return nil, 0, fmt.Errorf("API error: %s", resp.Message)
		}

		return resp.Data, resp.MetaData.TotalBatches, nil
	}

	return fetchBatched(mgr, "pfs/fuel-prices", &mgr.timeTracker.lastPricesFetch, decode, callback)
}

func (mgr *fuelPricesManager) GetFillingStations(callback BatchCallback[models.PetrolFillingStation]) (int, error) {
	decode := func(body io.ReadCloser, batchNo int) ([]models.PetrolFillingStation, int, error) {
		var resp []models.PetrolFillingStation
		decoder := json.NewDecoder(body)
		if err := decoder.Decode(&resp); err != nil {
			return nil, 0, fmt.Errorf("failed to unmarshal response: %w", err)
		}
		return resp, 0, nil // No total batches info available for PFS
	}

	return fetchBatched(mgr, "pfs", &mgr.timeTracker.lastPfsFetch, decode, callback)
}

func (mgr *fuelPricesManager) authenticate() error {
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
	mgr.timeTracker.lastAuth = time.Now()
	log.Printf("Authenticated successfully, token expires in %d seconds", resp.Data.ExpiresIn)

	return nil
}

func (mgr *fuelPricesManager) tokenRefresh() error {

	tokenReq := models.TokenRefreshRequest{
		ClientId:     mgr.authReq.ClientId,
		RefreshToken: mgr.tokenData.RefreshToken,
	}
	url := fmt.Sprintf("%s/oauth/regenerate_access_token", mgr.baseUrl)
	body, err := mgr.post(url, "application/json", tokenReq)
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

	mgr.tokenData.AccessToken = resp.Data.AccessToken
	mgr.tokenData.ExpiresIn = resp.Data.ExpiresIn
	mgr.timeTracker.lastAuth = time.Now()
	log.Printf("Token refresh completed successfully, token expires in %d seconds", mgr.tokenData.ExpiresIn)

	return nil
}

func (mgr *fuelPricesManager) checkTokenExpiry() error {
	expiryTime := mgr.timeTracker.lastAuth.Add(time.Duration(mgr.tokenData.ExpiresIn) * time.Second)
	expiresSoon := time.Until(expiryTime) < 5*time.Minute

	if expiresSoon {
		log.Printf("Access token has either expired or is expiring soon, refreshing...")
		if err := mgr.tokenRefresh(); err != nil {
			return fmt.Errorf("failed to refresh token: %w", err)
		}
	}
	return nil
}

func fetchBatched[T any](
	mgr *fuelPricesManager,
	path string,
	lastFetch *time.Time,
	decode func(io.ReadCloser, int) ([]T, int, error),
	callback BatchCallback[T],
) (int, error) {
	if err := mgr.checkTokenExpiry(); err != nil {
		return 0, fmt.Errorf("failed to refresh token: %w", err)
	}

	batchNo := 1
	count := 0

	startTime := time.Now()
	effectiveStartTimestamp := ""
	if !lastFetch.IsZero() {
		log.Printf("Time since last fetch for %s: %s", path, time.Since(*lastFetch))
		effectiveStartTimestamp = lastFetch.Format("2006-01-02 15:04:05") // Not quite RFC3339 ...
	}

	for {
		url := fmt.Sprintf("%s/%s?batch-number=%d", mgr.baseUrl, path, batchNo)
		if effectiveStartTimestamp != "" {
			url += "&effective-start-timestamp=" + neturl.QueryEscape(effectiveStartTimestamp)
		}
		body, err := mgr.get(url)
		if err != nil {
			var stErr *HTTPStatusError
			if errors.As(err, &stErr) && stErr.StatusCode == 400 {
				log.Printf("No more batches available for %s, stopping at batch %d", path, batchNo-1)
				break
			}
			return 0, err
		}

		data, totalBatches, err := decode(body, batchNo)
		if err != nil {
			_ = body.Close()
			return 0, err
		}
		_ = body.Close()

		numRecords, err := callback(data)
		if err != nil {
			return 0, fmt.Errorf("callback error: %w", err)
		}
		count += numRecords
		batchNo++

		if numRecords == 0 || (totalBatches > 0 && batchNo > totalBatches) {
			break
		}
	}

	*lastFetch = startTime
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
