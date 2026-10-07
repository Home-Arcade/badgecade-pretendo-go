package main

import (
	"os"

	"github.com/PretendoNetwork/plogger-go"
	"github.com/PretendoNetwork/nintendo-badge-arcade-secure/database"
	"github.com/joho/godotenv"
)

var logger = plogger.NewLogger()

func init() {
	err := godotenv.Load()

	if err != nil {
		logger.Warning("Error loading .env file")
	}

	if os.Getenv("KERBEROS_PASSWORD") == "" || os.Getenv("MONGO_URI") == "" || os.Getenv("DATABASE_URI") == "" {
		logger.Critical("KERBEROS_PASSWORD, MONGO_URI, and DATABASE_URI are required")
		os.Exit(1)
	}
	if err := validateFileStoreConfig(); err != nil {
		logger.Critical(err.Error())
		os.Exit(1)
	}
	database.ConnectAll()
}
