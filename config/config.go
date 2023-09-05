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
	AwsEndpoint  string
	AwsRegion    string
	AwsS3Bucket  string
	AwsSecretKey string
	AwsAccessKey string
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
		AwsEndpoint:  os.Getenv("AWS_ENDPOINT"),
		AwsS3Bucket:  os.Getenv("AWS_S3_BUCKET"),
		AwsRegion:    os.Getenv("AWS_REGION"),
		AwsSecretKey: os.Getenv("AWS_SECRET_ACCESS_KEY"),
		AwsAccessKey: os.Getenv("AWS_ACCESS_KEY_ID"),
	}

	return &configurations
}
