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
		Run: func(_ *cobra.Command, _ []string) {
			cmd.ApiServer(dbPath, port, debug)
		},
	}
	apiServerCmd.Flags().IntVar(&port, "port", 8080, "Port to run HTTP server on")
	apiServerCmd.Flags().BoolVar(&debug, "debug", false, "Enable debugging (pprof) - WARING: do not enable in production")

	rootCmd.AddCommand(apiServerCmd)
	rootCmd.PersistentFlags().StringVar(&dbPath, "db", "./data/fuel_prices.db", "Path to fuel-prices SQLite database")

	if err = rootCmd.Execute(); err != nil {
		panic(err)
	}
}