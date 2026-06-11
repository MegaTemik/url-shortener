package config

import (
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	Env         string `yaml:"env" env:"ENV" env-Default:"local"`
	DatabaseURL string `yaml:"database_url" env:"DATABASE_URL"`
	HTTPServer  `yaml:"http_server"`
	Auth
}

type Auth struct {
	User     string
	Password string
}

type HTTPServer struct {
	Address     string        `yaml:"address" env-default:"localhost:8080"`
	TimeOut     time.Duration `yaml:"timeout" env-default:"4s"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env-default:"60s"`
}

func MustLoad() *Config {
	err := godotenv.Load()
	// if err != nil {
	// 	log.Fatal("Error loading .env file")
	// }
	if err != nil {
		log.Println("No .env file found, relying on environment variables")
	}

	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		log.Fatal("CONFIG_PATH is not set")
	}
	//check if file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("config file does not exist: %s", configPath)
	}

	var cfg Config
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("cannot read config: %s", err)
	}

	if dbURL := os.Getenv("DATABASE_URL"); dbURL != "" {
		cfg.DatabaseURL = dbURL
		log.Println("Using DATABASE_URL from environment")
	}

	if authPass := os.Getenv("AUTH_PASSWORD"); authPass != "" {
		cfg.Auth.Password = authPass
		log.Println("Using AUTH_PASSWORD from environment")
	}

	if authUser := os.Getenv("AUTH_USER"); authUser != "" {
		cfg.Auth.User = authUser
		log.Println("Using AUTH_USER from environment")
	}

	if cfg.Auth.User == "" || cfg.Auth.Password == "" {
		log.Fatal("Authentication credentials are not set in config or environment variables")
	}

	return &cfg
}
