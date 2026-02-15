package routes

import (
	"fmt"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/rm-hull/fuel-prices-api/internal"
	"github.com/rm-hull/fuel-prices-api/internal/models"

	"github.com/gin-gonic/gin"
)

const MAX_BOUNDS = 5000 // Maximum bounds in meters (5 KM)

func Search(repo internal.FuelPricesRepository, client internal.FuelPricesClient) func(c *gin.Context) {
	return func(c *gin.Context) {
		bbox, err := parseBBox(c.Query("bbox"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		results, err := repo.Search(bbox)

		if err != nil {
			log.Printf("error while fetching fuel prices: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "An internal server error occurred"})
			return
		}

		c.JSON(http.StatusOK, models.SearchResponse{
			Results:     results,
			Attribution: internal.ATTRIBUTION,
			LastUpdated: client.LastUpdated(),
		})
	}
}

func parseBBox(bboxStr string) ([]float64, error) {
	bboxParts := strings.Split(bboxStr, ",")
	if len(bboxParts) != 4 {
		return nil, fmt.Errorf("bbox must have 4 comma-separated values")
	}

	bbox := make([]float64, 4)
	for i, part := range bboxParts {
		val, err := strconv.ParseFloat(strings.TrimSpace(part), 64)
		if err != nil {
			return nil, fmt.Errorf("invalid bbox value '%s': not a valid float", part)
		}
		bbox[i] = val
	}

	if math.Abs(bbox[2]-bbox[0]) > MAX_BOUNDS || math.Abs(bbox[3]-bbox[1]) > MAX_BOUNDS {
		return nil, fmt.Errorf("bbox must define a valid area (no more than %d KM in either dimension)", MAX_BOUNDS/1000)
	}

	return bbox, nil
}
