package config

import (
	"flag"
	"github.com/caarlos0/env/v6"
	"log"
	"strings"
)

type Config struct {
	RunAddr        string `env:"RUN_ADDRESS"`
	DBCfg          string `env:"DATABASE_URI"`
	AccrualAddress string `env:"ACCRUAL_SYSTEM_ADDRESS"`
}

func NewServerConfig() *Config {
	var db string
	var cfg Config
	var accrAddr string
	var addr string

	flag.StringVar(&addr, "a", "localhost:8080", "address and port to run server")
	flag.StringVar(&accrAddr, "r", "http://localhost:8081", "address and port to connect accrual server")
	flag.StringVar(&db, "d", "postgres://db_user:db_pass@localhost:5432/mydb?sslmode=disable", "db credential")

	flag.Parse()

	cfg.RunAddr = addr
	cfg.AccrualAddress = accrAddr
	cfg.DBCfg = db

	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("Error parsing env vars: %v", err)
	}

	cfg.RunAddr = strings.TrimPrefix(cfg.RunAddr, "http://")
	cfg.RunAddr = strings.TrimPrefix(cfg.RunAddr, "https://")

	return &cfg
}

func RecoveryServer() {
	if r := recover(); r != nil {
		log.Printf("CRITICAL error %v", r)

	}
}
