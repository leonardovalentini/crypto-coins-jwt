package app

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	"github.com/leonardovalentini/crypto-coins/crypto/docs"
	"github.com/leonardovalentini/crypto-coins/crypto/domain"
	"github.com/leonardovalentini/crypto-coins/crypto/handler"
	"github.com/leonardovalentini/crypto-coins/crypto/service"
	"github.com/leonardovalentini/crypto-coins/lib/helpers"
	"github.com/leonardovalentini/crypto-coins/lib/logger"
	"github.com/leonardovalentini/crypto-coins/lib/middleware"
	"github.com/leonardovalentini/crypto-coins/lib/middleware/authMiddleware"
	httpSwagger "github.com/swaggo/http-swagger"
)

func sanityCheck() {
	envProps := []string{
		"SERVER_ADDRESS",
		"SERVER_PORT",
		"DB_USER",
		"DB_PASSWORD",
		"DB_HOST",
		"DB_PORT",
		"DB_NAME",
		"AUTH_HOST",
		"AUTH_PORT",
		"GECKO_API_KEY",
		"CMC_API_KEY",
	}
	for _, k := range envProps {
		if os.Getenv(k) == "" {
			logger.Fatal(fmt.Sprintf("Environment variable %s not defined. Terminating application...", k))
		}
	}
}

// @title						Crypto API
// @version					1.0
// @description				This API allows you to keep a historical record of your favorite cryptocurrencies.
// @securityDefinitions.apikey	CookieAuth
// @in							cookie
// @name						session_token
// @host						localhost:8181
func Start() {
	sanityCheck()

	dbClient := getDbClient()
	defer dbClient.Close()

	client := &http.Client{Timeout: 10 * time.Second}

	var log logger.LoggerI = logger.NewLogger(context.Background())
	ur := domain.NewUserCoinRepositoryDb(dbClient)
	ucs := service.NewUserCoinService(ur)
	uch := handler.NewUserCoinHandler(ucs)

	pr := domain.NewPriceRepositoryDb(dbClient)
	cps1 := domain.NewBrokerSite1(client, "https://api.coingecko.com")
	cps2 := domain.NewBrokerSite2(client, "https://pro-api.coinmarketcap.com")
	cps := service.NewCoinsPriceService(ur, pr, []domain.BrokerSite{cps1, cps2}, log, time.Now)
	cph := handler.NewCoinsPriceHandler(cps)

	router := mux.NewRouter()
	router.Use(helpers.RecoveryMiddleware)
	router.Use(middleware.RequestIDMiddleware)
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	router.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)
	router.Use(logger.LoggingMiddleware())

	jobsRouter := router.PathPrefix("/jobs").Subrouter()
	cph.RegisterPublicRoutes(jobsRouter)

	privateRouter := router.PathPrefix("/api").Subrouter()
	privateRouter.Use(authMiddleware.AuthMiddleware())

	uch.RegisterPrivateRoutes(privateRouter)
	cph.RegisterPrivateRoutes(privateRouter)

	address := os.Getenv("SERVER_ADDRESS")
	port := os.Getenv("SERVER_PORT")

	docs.SwaggerInfo.Host = fmt.Sprintf("%s:%s", address, port)

	logger.Info(fmt.Sprintf("Starting server on %s:%s ...", address, port))
	logger.Fatal(http.ListenAndServe(fmt.Sprintf("%s:%s", address, port), router).Error())

}

func getDbClient() *sqlx.DB {
	dbUser := os.Getenv("DB_USER")
	dbPasswd := os.Getenv("DB_PASSWORD")
	dbAddr := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")

	dataSource := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", dbUser, dbPasswd, dbAddr, dbPort, dbName)
	client, err := sqlx.Open("mysql", dataSource)
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}
	// See "Important settings" section.
	client.SetConnMaxLifetime(time.Minute * 3)
	client.SetMaxOpenConns(10)
	client.SetMaxIdleConns(10)

	if err = client.Ping(); err != nil {
		log.Fatalf("Database unreachable: %v", err)
	}
	logger.Info("Connected to MySQL successfully!")
	return client
}
