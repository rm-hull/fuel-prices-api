package metrics

import (
	"log"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/rm-hull/fuel-prices-api/internal/models"
)

type fuelPricesGaugeCollector struct {
	avgDesc      *prometheus.Desc
	minDesc      *prometheus.Desc
	maxDesc      *prometheus.Desc
	stddevDesc   *prometheus.Desc
	sampleDesc   *prometheus.Desc
	distDesc     *prometheus.Desc
	snapshotFunc func() (*models.SnapshotStatistics, error)
	distFunc     func() (*models.DistributionStatistics, error)
}

func (c *fuelPricesGaugeCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.avgDesc
	ch <- c.minDesc
	ch <- c.maxDesc
	ch <- c.stddevDesc
	ch <- c.sampleDesc
	ch <- c.distDesc
}

func (c *fuelPricesGaugeCollector) Collect(ch chan<- prometheus.Metric) {

	snapshotStats, err := c.snapshotFunc()
	if err != nil {
		log.Printf("failed to collect fuel price snapshot stats: %v", err)
	} else {
		for _, s := range snapshotStats.Snapshot {
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

	distStats, err := c.distFunc()
	if err != nil {
		log.Printf("failed to collect fuel price distribution stats: %v", err)
	} else {
		for _, d := range distStats.Distribution {
			postcodeArea := ""
			if d.PostcodeArea != nil {
				postcodeArea = *d.PostcodeArea
			}

			for bucket, count := range d.Buckets {
				ch <- prometheus.MustNewConstMetric(c.distDesc, prometheus.GaugeValue, float64(count), postcodeArea, d.FuelType, strconv.Itoa(bucket))
			}
		}
	}
}

func RegisterFuelStatsCollector(reg prometheus.Registerer, snapshotFn func() (*models.SnapshotStatistics, error), distFn func() (*models.DistributionStatistics, error)) {

	labels := []string{"postcode_area", "fuel_type"}
	distLabels := []string{"postcode_area", "fuel_type", "price_bucket"}

	collector := fuelPricesGaugeCollector{
		avgDesc:      prometheus.NewDesc("fuel_prices_govuk_api_price_avg", "Average price at national and postcode area by fuel_type", labels, nil),
		minDesc:      prometheus.NewDesc("fuel_prices_govuk_api_price_min", "Minimum price at national and postcode area by fuel_type", labels, nil),
		maxDesc:      prometheus.NewDesc("fuel_prices_govuk_api_price_max", "Maximum price at national and postcode area by fuel_type", labels, nil),
		stddevDesc:   prometheus.NewDesc("fuel_prices_govuk_api_price_standard_deviation", "StdDev price at national and postcode area by fuel_type", labels, nil),
		sampleDesc:   prometheus.NewDesc("fuel_prices_govuk_api_price_sample_size", "Price sample size at national and postcode area by fuel_type", labels, nil),
		distDesc:     prometheus.NewDesc("fuel_prices_govuk_api_price_distribution", "Price distribution sample size at national and postcode area by fuel_type and price bucket", distLabels, nil),
		snapshotFunc: snapshotFn,
		distFunc:     distFn,
	}

	if reg != nil {
		reg.MustRegister(&collector)
	}
}
