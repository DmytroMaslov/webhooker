package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Postgress PgCredentials
}

type PgCredentials struct {
	User     string
	Password string
	Host     string
	DbName   string
}

func GetConfig() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	user := os.Getenv("PG_USER")
	pass := os.Getenv("PG_PASS")
	host := os.Getenv("PG_HOST")
	dbName := os.Getenv("PG_DB_NAME")

	if user == "" && pass == "" && host == "" && dbName == "" {
		return nil, fmt.Errorf("config fields is empty")
	}

	return &Config{
		Postgress: PgCredentials{
			User:     user,
			Password: pass,
			Host:     host,
			DbName:   dbName,
		},
	}, nil
}
