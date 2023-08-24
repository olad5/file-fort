package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Configurations struct {
	DatabaseUrl  string
	Port         string
	JwtSecretKey string
	CacheAddress string
}

func GetConfig(filepath string) *Configurations {
	err := godotenv.Load(filepath)
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	configurations := Configurations{
		DatabaseUrl:  os.Getenv("DATABASE_URL"),
		Port:         os.Getenv("PORT"),
		JwtSecretKey: os.Getenv("SECRET_KEY"),
		CacheAddress: os.Getenv("REDIS_URL"),
	}

	return &configurations
}
