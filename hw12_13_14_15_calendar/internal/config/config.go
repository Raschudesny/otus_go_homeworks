package config

import (
	"errors"
	"fmt"
	"os"
	"path"

	"github.com/spf13/viper"
	"go.uber.org/zap"
)

const configErrorCausedFallthroughToDefaultsMsg = "configuration file error detected, default value will be used"

var (
	ErrLoggerLevelIsEmpty   = errors.New("logger level is empty")
	ErrLoggerFileIsEmpty    = errors.New("logger output file path is empty")
	ErrDBHostIsEmpty        = errors.New("db host is empty")
	ErrDBPortIsInvalid      = errors.New("db port is invalid")
	ErrDBUsernameIsEmpty    = errors.New("db username is empty")
	ErrDBPassIsEmpty        = errors.New("db pass is empty")
	ErrDBDBIsEmpty          = errors.New("database name is empty")
	ErrHTTPPortIsInvalid    = errors.New("http port is invalid")
	ErrHTTPTimeoutIsInvalid = errors.New("http connection timeout is invalid")
	ErrGRPCPortIsInvalid    = errors.New("grpc port is invalid")
	ErrGRPCTimeoutIsInvalid = errors.New("grpc connection timeout is invalid")
)

type Config struct {
	Logger  LoggerConfig
	Storage StorageConfig
	API     APIConfig
}

type LoggerConfig struct {
	Level string
	File  string
}

type StorageConfig struct {
	UseMemoryStorage bool `mapstructure:"inmemorystorage"`
	DB               DBConfig
}

type DBConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	DB       string
}

type APIConfig struct {
	GRPC GRPCApiConfig `mapstructure:"grpc"`
	HTTP HTTPApiConfig `mapstructure:"http"`
}

type GRPCApiConfig struct {
	Port              int
	ConnectionTimeout int `mapstructure:"connectionTimeout"`
}

type HTTPApiConfig struct {
	Port              int
	ConnectionTimeout int `mapstructure:"connectionTimeout"`
}

func (db *DBConfig) fallthroughToDefaults() {
	if db.Host == "" {
		db.Host = "localhost"
		zap.L().Error(configErrorCausedFallthroughToDefaultsMsg, zap.Error(ErrDBHostIsEmpty), zap.String("default", db.Host))
	}
	if db.Port == 0 {
		db.Port = 5432
		zap.L().Error(configErrorCausedFallthroughToDefaultsMsg, zap.Error(ErrDBPortIsInvalid), zap.Int("default", db.Port))
	}
	if db.Username == "" {
		db.Username = "postgres"
		zap.L().Error(configErrorCausedFallthroughToDefaultsMsg, zap.Error(ErrDBUsernameIsEmpty), zap.String("default", db.Username))
	}
	if db.Password == "" {
		db.Password = "postgres"
		zap.L().Error(configErrorCausedFallthroughToDefaultsMsg, zap.Error(ErrDBPassIsEmpty), zap.String("default", db.Password))
	}
	if db.DB == "" {
		db.DB = "calendar"
		zap.L().Error(configErrorCausedFallthroughToDefaultsMsg, zap.Error(ErrDBDBIsEmpty), zap.String("default", db.DB))
	}
}

func (conf *LoggerConfig) fallthroughToDefaults() {
	if conf.File == "" {
		defaultLogPath := "./log_output"
		dir, err := os.Getwd()
		if err == nil {
			defaultLogPath = path.Join(dir, defaultLogPath)
		}
		conf.File = defaultLogPath
		zap.L().Error(configErrorCausedFallthroughToDefaultsMsg, zap.Error(ErrLoggerFileIsEmpty), zap.String("default", conf.File))
	}
	if conf.Level == "" {
		conf.Level = "info"
		zap.L().Error(configErrorCausedFallthroughToDefaultsMsg, zap.Error(ErrLoggerLevelIsEmpty), zap.String("default", conf.Level))
	}
}

func (conf *StorageConfig) fallthroughToDefaults() {
	if !conf.UseMemoryStorage {
		conf.DB.fallthroughToDefaults()
	}
}

func (conf *APIConfig) fallthroughToDefaults() {
	if conf.HTTP.Port == 0 {
		conf.HTTP.Port = 80
		zap.L().Error(configErrorCausedFallthroughToDefaultsMsg, zap.Error(ErrHTTPPortIsInvalid), zap.Int("default", conf.HTTP.Port))
	}
	if conf.HTTP.ConnectionTimeout == 0 {
		conf.HTTP.ConnectionTimeout = 10
		zap.L().Error(configErrorCausedFallthroughToDefaultsMsg, zap.Error(ErrHTTPTimeoutIsInvalid), zap.Int("default", conf.HTTP.ConnectionTimeout))
	}
	if conf.GRPC.Port == 0 {
		conf.GRPC.Port = 50051
		zap.L().Error(configErrorCausedFallthroughToDefaultsMsg, zap.Error(ErrGRPCPortIsInvalid), zap.Int("default", conf.GRPC.Port))
	}
	if conf.GRPC.ConnectionTimeout == 0 {
		conf.GRPC.ConnectionTimeout = 10
		zap.L().Error(configErrorCausedFallthroughToDefaultsMsg, zap.Error(ErrGRPCTimeoutIsInvalid), zap.Int("default", conf.GRPC.ConnectionTimeout))
	}
}

func (conf *Config) fallthroughToDefaults() {
	conf.Storage.fallthroughToDefaults()
	conf.Logger.fallthroughToDefaults()
	conf.API.fallthroughToDefaults()
}

func NewConfig(configFilePath string) (cfg *Config, err error) {
	viper.SetConfigFile(configFilePath)
	viper.SetConfigType("yaml")
	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("error during reading config file: %w", err)
	}
	if err = viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}
	cfg.fallthroughToDefaults()
	zap.L().Info("result config", zap.String("config", fmt.Sprintf("%v", cfg)))
	return cfg, nil
}
