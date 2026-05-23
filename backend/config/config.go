package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type AppConfig struct {
	Server    ServerConfig    `mapstructure:"server"`
	Postgres  PostgresConfig  `mapstructure:"postgres"`
	Redis     RedisConfig     `mapstructure:"redis"`
	Session   SessionConfig   `mapstructure:"session"`
	Auth      AuthConfig      `mapstructure:"auth"`
	Point     PointConfig     `mapstructure:"point"`
	CourseHot CourseHotConfig `mapstructure:"course_hot"`
	Asynq     AsynqConfig     `mapstructure:"asynq"`
	Stats     StatsConfig     `mapstructure:"stats"`
	SMTP      SMTPConfig      `mapstructure:"smtp"`
}

type SMTPConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	From     string `mapstructure:"from"`
}

type AsynqConfig struct {
	Concurrency int `mapstructure:"concurrency"`
}

type StatsConfig struct {
	DailyCron        string `mapstructure:"daily_cron"`
	SchedulerEnabled bool   `mapstructure:"scheduler_enabled"`
}

type SessionConfig struct {
	Secret string `mapstructure:"secret"`
	MaxAge int    `mapstructure:"max_age"` // seconds
	Secure bool   `mapstructure:"secure"`  // HTTPS only
}

type AuthConfig struct {
	EmailWhitelist           []string `mapstructure:"email_whitelist"`
	UsernameSalt             string   `mapstructure:"username_salt"`
	VerificationCodeInterval int      `mapstructure:"verification_code_interval"` // seconds
	VerificationCodeTTL      int      `mapstructure:"verification_code_ttl"`      // seconds
	MaxLoginAttempts         int      `mapstructure:"max_login_attempts"`
	LoginLockoutDuration     int      `mapstructure:"login_lockout_duration"` // seconds
}

type PointConfig struct {
	TransferFeeRateBps int `mapstructure:"transfer_fee_rate_bps"`
	TransferMinFee     int `mapstructure:"transfer_min_fee"`
}

type CourseHotConfig struct {
	ReviewCreateScore int64 `mapstructure:"review_create_score"`
	ReviewUpdateScore int64 `mapstructure:"review_update_score"`
	ReviewVoteScore   int64 `mapstructure:"review_vote_score"`
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
	v.SetDefault("stats.daily_cron", "10 0 * * *")
	v.SetDefault("stats.scheduler_enabled", true)
	v.SetDefault("course_hot.review_create_score", 3)
	v.SetDefault("course_hot.review_update_score", 1)
	v.SetDefault("course_hot.review_vote_score", 1)

	if err := v.ReadInConfig(); err != nil {
		return AppConfig{}, fmt.Errorf("read config: %w", err)
	}

	var conf AppConfig
	if err := v.Unmarshal(&conf); err != nil {
		return AppConfig{}, fmt.Errorf("unmarshal config: %w", err)
	}
	if strings.TrimSpace(conf.Auth.UsernameSalt) == "" {
		return AppConfig{}, fmt.Errorf("auth.username_salt is required")
	}
	return conf, nil
}
