package metrics

import (
	"log"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/rm-hull/fuel-prices-api/internal/models"
)

type fuelPricesGaugeCollector struct {
	avgDesc    *prometheus.Desc
	minDesc    *prometheus.Desc
	maxDesc    *prometheus.Desc
	stddevDesc *prometheus.Desc
	sampleDesc *prometheus.Desc
	valFunc    func() ([]models.SnapshotStatistics, error)
}

func (c *fuelPricesGaugeCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.avgDesc
	ch <- c.minDesc
	ch <- c.maxDesc
	ch <- c.stddevDesc
	ch <- c.sampleDesc
}

func (c *fuelPricesGaugeCollector) Collect(ch chan<- prometheus.Metric) {

	stats, err := c.valFunc()
	if err != nil {
		log.Printf("failed to collect fuel price stats: %v", err)
		return
	}

	for _, s := range stats {
		postcodeArea := ""
		if s.PostcodeArea != nil {
			postcodeArea = *s.PostcodeArea
		}

		ch <- prometheus.MustNewConstMetric(c.avgDesc, prometheus.GaugeValue, s.AveragePrice, postcodeArea, s.FuelType)
		ch <- prometheus.MustNewConstMetric(c.minDesc, prometheus.GaugeValue, s.LowestPrice, postcodeArea, s.FuelType)
		ch <- prometheus.MustNewConstMetric(c.maxDesc, prometheus.GaugeValue, s.HighestPrice, postcodeArea, s.FuelType)
		ch <- prometheus.MustNewConstMetric(c.stddevDesc, prometheus.GaugeValue, s.StandardDeviation, postcodeArea, s.FuelType)
		ch <- prometheus.MustNewConstMetric(c.sampleDesc, prometheus.GaugeValue, float64(s.SampleSize), postcodeArea, s.FuelType)
	}
}

func RegisterFuelStatsCollector(reg prometheus.Registerer, fn func() ([]models.SnapshotStatistics, error)) {

	labels := []string{"postcode_area", "fuel_type"}
	collector := fuelPricesGaugeCollector{
		avgDesc:    prometheus.NewDesc("fuel_prices_govuk_api_price_avg", "Average price at national and postcode area by fuel_type", labels, nil),
		minDesc:    prometheus.NewDesc("fuel_prices_govuk_api_price_min", "Minimum price at national and postcode area by fuel_type", labels, nil),
		maxDesc:    prometheus.NewDesc("fuel_prices_govuk_api_price_max", "Maximum price at national and postcode area by fuel_type", labels, nil),
		stddevDesc: prometheus.NewDesc("fuel_prices_govuk_api_price_standard_deviation", "StdDev price at national and postcode area by fuel_type", labels, nil),
		sampleDesc: prometheus.NewDesc("fuel_prices_govuk_api_price_sample", "Price sample size at national and postcode area by fuel_type", labels, nil),
		valFunc:    fn,
	}

	if reg != nil {
		reg.MustRegister(&collector)
	}
}
