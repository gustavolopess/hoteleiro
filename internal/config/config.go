package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

type Config struct {
	DynamoDbUri string
	AwsRegion   string
}

const (
	dynamoDbUriConfigKey = "dynamodb_uri"
	awsRegionConfigKey   = "aws_region"
)

func LoadConfig() *Config {
	appEnv := os.Getenv("APP_ENV")
	if appEnv == "" {
		appEnv = "local"
	}

	viper.SetConfigName("config." + appEnv)
	viper.SetConfigType("yml")
	viper.AddConfigPath("./config/")
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("fatal error config file: %w", err))
	}

	return &Config{
		DynamoDbUri: viper.GetString(dynamoDbUriConfigKey),
		AwsRegion:   viper.GetString(awsRegionConfigKey),
	}
}
