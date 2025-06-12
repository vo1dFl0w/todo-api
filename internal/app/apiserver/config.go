package apiserver

import (
	"log"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env 	 	string `yaml:"env" env-default:"local" env-required:"true"`
	HTTPAddr 	string `yaml:"httpaddr" env-default:"localhost:8080" env-required:"true"`
	DatabaseURL string `yaml:"databaseurl" env-required:"true"`
	JWTSecret   string `yaml:"jwt_secret"`
}

func NewConfig() *Config {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		log.Fatal("config path is not set")
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("config is not exist: %s", err)
	}

	return &cfg
}