package config

import (
	"errors"
	"fmt"

	"github.com/spf13/viper"
	"go.uber.org/zap"
)

const configErrorCausedFallthroughToDefaultsMsg = "configuration file error detected, default value will be used"

var (
	ErrLoggerLevelIsEmpty = errors.New("logger level is empty")
	ErrLoggerFileIsEmpty  = errors.New("logger output file path is empty")
	ErrDBHostIsEmpty      = errors.New("db host is empty")
	ErrDBPortIsInvalid    = errors.New("db port is invalid")
	ErrDBUsernameIsEmpty  = errors.New("db username is empty")
	ErrDBPassIsEmpty      = errors.New("db pass is empty")
	ErrDBDBIsEmpty        = errors.New("database name is empty")
	ErrAPIPortIsInvalid   = errors.New("api port is invalid")
)

type Config struct {
	Logger  LoggerConfig
	Storage StorageConfig
	API     APIConfig `mapstructure:"api"`
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
	Port int
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
		conf.File = "./log_output"
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

func (conf APIConfig) fallthroughToDefaults() {
	if conf.Port == 0 {
		conf.Port = 9933
		zap.L().Error(configErrorCausedFallthroughToDefaultsMsg, zap.Error(ErrAPIPortIsInvalid), zap.Int("default", conf.Port))
	}
}

func (conf Config) fallthroughToDefaults() {
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
