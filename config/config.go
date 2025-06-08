package config

import (
	"flag"
	"github.com/caarlos0/env/v6"
	"log"
	"reflect"
	"time"
)

type Config struct {
	RunAddr         string `env:"RUN_ADDRESS"`
	DBCfg           string `env:"DATABASE_URI"`
	Accrual_address string `env:"ACCRUAL_SYSTEM_ADDRESS"`
}

func NewServerConfig() *Config {
	var db string
	var cfg Config
	var accr_addr string
	var addr string

	flag.StringVar(&addr, "a", "localhost:8080", "address and port to run server")
	flag.StringVar(&accr_addr, "r", "localhost:8081", "address and port to run server")
	flag.StringVar(&db, "d", "", "db credential")

	flag.Parse()

	cfg.RunAddr = addr
	cfg.Accrual_address = accr_addr
	cfg.DBCfg = db

	err := env.ParseWithFuncs(&cfg, map[reflect.Type]env.ParserFunc{
		reflect.TypeOf(time.Duration(0)): func(value string) (interface{}, error) {
			num, err := time.ParseDuration(value)
			if err == nil {
				return num, nil
			}
			seconds, err := time.ParseDuration(value + "s")
			if err != nil {
				return nil, err
			}
			return seconds, nil
		},
	},
	)

	if err != nil {
		return nil
	}

	return &cfg
}

func RecoveryServer() {
	if r := recover(); r != nil {
		log.Printf("CRITICAL error %v", r)

	}
}
