package app

import (
	"crypto/rand"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	"github.com/leonardovalentini/crypto-coins/auth/docs"
	"github.com/leonardovalentini/crypto-coins/auth/domain"
	"github.com/leonardovalentini/crypto-coins/auth/handler"
	"github.com/leonardovalentini/crypto-coins/auth/service"
	"github.com/leonardovalentini/crypto-coins/lib/helpers"
	"github.com/leonardovalentini/crypto-coins/lib/logger"
	"github.com/leonardovalentini/crypto-coins/lib/middleware"
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
	}
	for _, k := range envProps {
		if os.Getenv(k) == "" {
			logger.Fatal(fmt.Sprintf("Environment variable %s not defined. Terminating application...", k))
		}
	}
}

// @title						Auth API
// @version					1.0
// @description				This API allows for user registration and validation.
// @securityDefinitions.apikey	CookieAuth
// @in							cookie
// @name						session_token
// @host						localhost:8080
func Start() {
	sanityCheck()

	dbClient := getDbClient()
	defer dbClient.Close()

	UserRepositoryDb := domain.NewUserRepositoryDb(dbClient)
	ah := handler.NewAuthHandler(service.NewAuthService(UserRepositoryDb, rand.Read, time.Now))

	router := mux.NewRouter()
	router.Use(helpers.RecoveryMiddleware)
	router.Use(middleware.RequestIDMiddleware)
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	router.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)
	router.Use(logger.LoggingMiddleware())

	ah.RegisterRoutes(router)

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

	dataSource := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", dbUser, dbPasswd, dbAddr, dbPort, dbName)
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
