package cmd

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/Depado/ginprom"
	"github.com/aurowora/compress"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"

	"github.com/rm-hull/fuel-prices-api/internal"
	"github.com/rm-hull/fuel-prices-api/internal/routes"
	"github.com/rm-hull/godx"
	healthcheck "github.com/tavsec/gin-healthcheck"
	"github.com/tavsec/gin-healthcheck/checks"
	hc_config "github.com/tavsec/gin-healthcheck/config"
)

func ApiServer(dbPath string, port int, debug bool) {

	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	godx.GitVersion()
	godx.EnvironmentVars()
	godx.UserInfo()

	clientId := os.Getenv("CLIENT_ID")
	clientSecret := os.Getenv("CLIENT_SECRET")

	client, err := internal.NewFuelPricesClient(clientId, clientSecret)
	if err != nil {
		log.Fatalf("GOV.UK authentication failed: %v\n", err)
	}

	db, err := internal.Connect(dbPath)
	if err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("error closing database: %v", err)
		}
	}()

	err = internal.Migrate("migrations", dbPath)
	if err != nil {
		log.Fatalf("failed to migrate SQL: %v", err)
	}

	repo := internal.NewFuelPricesRepository(db)
	if _, err := internal.StartCron(client, repo); err != nil {
		log.Fatalf("failed to start CRON jobs: %v", err)
	}

	r := gin.New()

	prometheus := ginprom.New(
		ginprom.Engine(r),
		ginprom.Path("/metrics"),
		ginprom.Ignore("/healthz"),
	)

	r.Use(
		gin.Recovery(),
		gin.LoggerWithWriter(gin.DefaultWriter, "/healthz", "/metrics"),
		prometheus.Instrument(),
		compress.Compress(),
		cors.Default(),
	)

	if debug {
		log.Println("WARNING: pprof endpoints are enabled and exposed. Do not run with this flag in production.")
		pprof.Register(r)
	}

	err = healthcheck.New(r, hc_config.DefaultConfig(), []checks.Check{
		checks.SqlCheck{Sql: db},
	})
	if err != nil {
		log.Fatalf("failed to initialize healthcheck: %v", err)
	}

	v1 := r.Group("/v1/fuel-prices")
	v1.GET("/search", routes.Search(repo, client))

	addr := fmt.Sprintf(":%d", port)
	log.Printf("Starting HTTP API Server on port %d...", port)
	if err := r.Run(addr); err != nil && err != http.ErrServerClosed {
		log.Fatalf("HTTP API Server failed to start on port %d: %v", port, err)
	}
}
