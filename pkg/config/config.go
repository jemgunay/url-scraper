package config

import (
	"errors"
	"fmt"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/yaml.v3"
)

// Config contains the service configuration.
type Config struct {
	Port   int  `yaml:"port"`
	Debug  bool `yaml:"debug"`
	Logger `yaml:"-"`
}

// New initialises a Config from a yaml file on disk. It also initialises a
// service Logger.
func New(filePath string) (Config, error) {
	conf := Config{}

	f, err := os.Open(filePath)
	if err != nil {
		return conf, fmt.Errorf("failed to open yaml config file: %s", err)
	}
	defer f.Close()

	if err := yaml.NewDecoder(f).Decode(&conf); err != nil {
		return conf, fmt.Errorf("failed to decode yaml config: %w", err)
	}

	logLevel := zapcore.InfoLevel
	if conf.Debug {
		logLevel = zapcore.DebugLevel
	}
	if conf.Logger, err = newLogger(logLevel); err != nil {
		return conf, fmt.Errorf("failed to initialise logger: %w", err)
	}

	if err := conf.Validate(); err != nil {
		return conf, fmt.Errorf("failed to validate yaml config: %w", err)
	}

	return conf, nil
}

// Validate validates that the config has been properly initialised.
func (c Config) Validate() error {
	switch {
	case c.Logger == nil:
		return errors.New("logger is uninitialised")
	case c.Port == 0:
		return errors.New("invalid port config provided")
	}
	return nil
}

// Logger defines the required logger functionality.
type Logger interface {
	Debug(msg string, fields ...zapcore.Field)
	Info(msg string, fields ...zapcore.Field)
	Warn(msg string, fields ...zapcore.Field)
	Error(msg string, fields ...zapcore.Field)
	Fatal(msg string, fields ...zapcore.Field)
	With(fields ...zap.Field) *zap.Logger
}

func newLogger(level zapcore.Level) (Logger, error) {
	loggerConfig := zap.NewProductionConfig()
	loggerConfig.EncoderConfig.EncodeLevel = zapcore.LowercaseLevelEncoder
	loggerConfig.EncoderConfig.TimeKey = "ts"
	loggerConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	loggerConfig.Level.SetLevel(level)

	logger, err := loggerConfig.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build logger: %s", err)
	}

	return logger, nil
}
