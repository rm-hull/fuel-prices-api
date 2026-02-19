package main

import (
	"github.com/rm-hull/fuel-prices-api/cmd"

	"github.com/spf13/cobra"
)

func main() {
	var err error
	var dbPath string
	var port int
	var debug bool

	rootCmd := &cobra.Command{
		Use:  "fuel-prices",
		Long: `Fuel Prices API`,
	}

	apiServerCmd := &cobra.Command{
		Use:   "api-server [--db <path>] [--port <port>] [--debug]",
		Short: "Start HTTP API server",
		RunE: func(_ *cobra.Command, _ []string) error {
			return cmd.ApiServer(dbPath, port, debug)
		},
	}

	importCmd := &cobra.Command{
		Use:   "import [--db <path>]",
		Short: "Perform one-off import of fuel prices and filling stations from the GOV.UK API",
		RunE: func(_ *cobra.Command, _ []string) error {
			return cmd.Import(dbPath)
		},
	}
	apiServerCmd.Flags().IntVar(&port, "port", 8080, "Port to run HTTP server on")
	apiServerCmd.Flags().BoolVar(&debug, "debug", false, "Enable debugging (pprof) - WARING: do not enable in production")

	rootCmd.AddCommand(apiServerCmd)
	rootCmd.AddCommand(importCmd)
	rootCmd.PersistentFlags().StringVar(&dbPath, "db", "./data/fuel_prices.db", "Path to fuel-prices SQLite database")

	if err = rootCmd.Execute(); err != nil {
		panic(err)
	}
}
