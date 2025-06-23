package logger

import (
	"github.com/stretchr/testify/assert/yaml"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
)

var Log = zap.NewNop()

type Config struct {
	Level            string   `yaml:"level"`
	Encoding         string   `yaml:"encoding"`
	OutputPaths      []string `yaml:"outputPaths"`
	ErrorOutputPaths []string `yaml:"errorOutputPaths"`
	EncoderConfig    struct {
		MessageKey string `yaml:"messageKey"`
		LevelKey   string `yaml:"levelKey"`
		LevelEnc   string `yaml:"levelEncoder"`
		TimeKey    string `yaml:"timeKey"`
		TimeEnc    string `yaml:"encodeTime"`
	} `yaml:"encoderConfig"`
}

type AppConfig struct {
	Logger Config `yaml:"logger"`
}

func Init(level string) error {
	var appCfg AppConfig

	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return err
	}

	var cfg zap.Config

	configFile, err := os.ReadFile("config.yaml")

	if err != nil {
		cfg = zap.NewDevelopmentConfig()
	} else {
		err := yaml.Unmarshal(configFile, &appCfg)
		if err != nil {
			return err
		}

		cfg = ConvertToZapConfig(appCfg.Logger)
	}
	cfg.Level = lvl

	zapLogger, err := cfg.Build()
	if err != nil {
		return err
	}
	Log = zapLogger
	return nil
}

func ConvertToZapConfig(c Config) zap.Config {
	zc := zap.Config{
		Level:            zap.NewAtomicLevelAt(parseLevel(c.Level)),
		Development:      false,
		Encoding:         c.Encoding,
		OutputPaths:      c.OutputPaths,
		ErrorOutputPaths: c.ErrorOutputPaths,
		EncoderConfig: zapcore.EncoderConfig{
			MessageKey:  c.EncoderConfig.MessageKey,
			LevelKey:    c.EncoderConfig.LevelKey,
			EncodeLevel: parseLevelEncoder(c.EncoderConfig.LevelEnc),
			TimeKey:     c.EncoderConfig.TimeKey,
			EncodeTime:  parseTimeEncoder(c.EncoderConfig.TimeEnc),
		},
	}
	return zc
}

func parseLevel(level string) zapcore.Level {
	l := new(zapcore.Level)
	err := l.UnmarshalText([]byte(level))
	if err != nil {
		return zapcore.InfoLevel // дефолт
	}
	return *l
}

func parseLevelEncoder(enc string) zapcore.LevelEncoder {
	switch enc {
	case "capital":
		return zapcore.CapitalLevelEncoder
	case "color":
		return zapcore.CapitalColorLevelEncoder
	default:
		return zapcore.LowercaseLevelEncoder
	}
}

func parseTimeEncoder(enc string) zapcore.TimeEncoder {
	switch enc {
	case "iso8601":
		return zapcore.ISO8601TimeEncoder
	case "epoch":
		return zapcore.EpochTimeEncoder
	default:
		return zapcore.RFC3339TimeEncoder
	}
}

func Sync() {
	if Log != nil {
		_ = Log.Sync()
	}
}
