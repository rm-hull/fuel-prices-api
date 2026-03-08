package routes

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rm-hull/fuel-prices-api/internal"
)

func SnapshotStats(repo internal.FuelPricesRepository) func(c *gin.Context) {
	return func(c *gin.Context) {
		stats, err := repo.SnapshotStats()

		if err != nil {
			log.Printf("error while fetching stats: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "An internal server error occurred"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"snapshot":     stats.Snapshot,
			"last_updated": stats.LastUpdated,
			"attribution":  internal.ATTRIBUTION,
		})
	}
}
