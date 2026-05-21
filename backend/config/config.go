package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type AppConfig struct {
	Server   ServerConfig   `mapstructure:"server"`
	Postgres PostgresConfig `mapstructure:"postgres"`
	Redis    RedisConfig    `mapstructure:"redis"`
	Session  SessionConfig  `mapstructure:"session"`
	Auth     AuthConfig     `mapstructure:"auth"`
	Asynq    AsynqConfig    `mapstructure:"asynq"`
}

type AsynqConfig struct {
	Concurrency int `mapstructure:"concurrency"`
}

type SessionConfig struct {
	Secret string `mapstructure:"secret"`
	MaxAge int    `mapstructure:"max_age"` // seconds
	Secure bool   `mapstructure:"secure"`  // HTTPS only
}

type AuthConfig struct {
	EmailWhitelist           []string `mapstructure:"email_whitelist"`
	VerificationCodeInterval int      `mapstructure:"verification_code_interval"` // seconds
	VerificationCodeTTL      int      `mapstructure:"verification_code_ttl"`      // seconds
}

type ServerConfig struct {
	Addr string `mapstructure:"addr"`
}

type PostgresConfig struct {
	DSN string `mapstructure:"dsn"`
}

type RedisConfig struct {
	Addr     string `mapstructure:"addr"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

func Load(configPath string) (AppConfig, error) {
	v := viper.New()

	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		v.AddConfigPath(".")
		v.SetConfigName("config")
		v.SetConfigType("yaml")
	}

	v.SetEnvPrefix("JCOURSE")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		return AppConfig{}, fmt.Errorf("read config: %w", err)
	}

	var conf AppConfig
	if err := v.Unmarshal(&conf); err != nil {
		return AppConfig{}, fmt.Errorf("unmarshal config: %w", err)
	}
	return conf, nil
}
