package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"

	"jcourse/internal/domain/account"
	"jcourse/internal/domain/auth"
	"jcourse/internal/domain/course"
	"jcourse/internal/domain/point"
	"jcourse/internal/domain/review"
	"jcourse/internal/domain/review/policy"
	"jcourse/internal/domain/stat"
)

type AppConfig struct {
	Server    ServerConfig    `mapstructure:"server"`
	Postgres  PostgresConfig  `mapstructure:"postgres"`
	Redis     RedisConfig     `mapstructure:"redis"`
	Session   SessionConfig   `mapstructure:"session"`
	Auth      AuthConfig      `mapstructure:"auth"`
	Review    ReviewConfig    `mapstructure:"review"`
	APIKey    APIKeyConfig    `mapstructure:"api_key"`
	Admin     AdminConfig     `mapstructure:"admin"`
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
	Timezone         string `mapstructure:"timezone"`
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
	VerificationCodeLength   int      `mapstructure:"verification_code_length"`
	MaxLoginAttempts         int      `mapstructure:"max_login_attempts"`
	LoginLockoutDuration     int      `mapstructure:"login_lockout_duration"` // seconds
	PasswordHashIterations   int      `mapstructure:"password_hash_iterations"`
	PasswordSaltLength       int      `mapstructure:"password_salt_length"`
}

type ReviewConfig struct {
	FrequencyWindowSeconds int     `mapstructure:"frequency_window_seconds"`
	FrequencyMaxReviews    int     `mapstructure:"frequency_max_reviews"`
	SimilarityRatio        float64 `mapstructure:"similarity_ratio"`
	MaxDailyVotes          int     `mapstructure:"max_daily_votes"`
}

type APIKeyConfig struct {
	MaxUserKeys int `mapstructure:"max_user_keys"`
}

type AdminConfig struct {
	DefaultSuspendDays int `mapstructure:"default_suspend_days"`
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
	statsDefaults := stat.DefaultConfig()
	registrationDefaults := account.DefaultRegistrationConfig()
	loginDefaults := account.DefaultLoginConfig()
	passwordHashDefaults := account.DefaultPasswordHashConfig()
	pointDefaults := point.DefaultTransferFeeConfig()
	reviewFrequencyDefaults := policy.DefaultFrequencyPolicyConfig()
	voteDefaults := review.DefaultVoteConfig()
	apiKeyDefaults := auth.DefaultApiKeyConfig()
	adminDefaults := auth.DefaultAdminConfig()

	v.SetDefault("stats.daily_cron", statsDefaults.DailyCron)
	v.SetDefault("stats.scheduler_enabled", statsDefaults.SchedulerEnabled)
	v.SetDefault("stats.timezone", statsDefaults.Timezone)
	v.SetDefault("auth.verification_code_interval", int(registrationDefaults.CodeInterval/time.Second))
	v.SetDefault("auth.verification_code_ttl", int(registrationDefaults.CodeTTL/time.Second))
	v.SetDefault("auth.verification_code_length", registrationDefaults.CodeLength)
	v.SetDefault("auth.max_login_attempts", loginDefaults.MaxAttempts)
	v.SetDefault("auth.login_lockout_duration", int(loginDefaults.Lockout/time.Second))
	v.SetDefault("auth.password_hash_iterations", passwordHashDefaults.Iterations)
	v.SetDefault("auth.password_salt_length", passwordHashDefaults.SaltLength)
	v.SetDefault("point.transfer_fee_rate_bps", pointDefaults.RateBps)
	v.SetDefault("point.transfer_min_fee", pointDefaults.MinFee)
	v.SetDefault("review.frequency_window_seconds", int(reviewFrequencyDefaults.Window/time.Second))
	v.SetDefault("review.frequency_max_reviews", reviewFrequencyDefaults.MaxReviews)
	v.SetDefault("review.similarity_ratio", reviewFrequencyDefaults.SimilarityRatio)
	v.SetDefault("review.max_daily_votes", voteDefaults.MaxDailyVotes)
	v.SetDefault("api_key.max_user_keys", apiKeyDefaults.MaxUserKeys)
	v.SetDefault("admin.default_suspend_days", adminDefaults.DefaultSuspendDays)
	v.SetDefault("course_hot.review_create_score", course.DefaultHotScoreConfig.ReviewCreateScore)
	v.SetDefault("course_hot.review_update_score", course.DefaultHotScoreConfig.ReviewUpdateScore)
	v.SetDefault("course_hot.review_vote_score", course.DefaultHotScoreConfig.ReviewVoteScore)

	if err := v.ReadInConfig(); err != nil {
		return AppConfig{}, fmt.Errorf("read config: %w", err)
	}

	var conf AppConfig
	if err := v.Unmarshal(&conf); err != nil {
		return AppConfig{}, fmt.Errorf("unmarshal config: %w", err)
	}
	if strings.TrimSpace(conf.Session.Secret) == "" || conf.Session.Secret == "change-me-in-production" {
		return AppConfig{}, fmt.Errorf("session.secret must be set to a strong random value")
	}
	if len(conf.Session.Secret) < 32 {
		return AppConfig{}, fmt.Errorf("session.secret must be at least 32 characters")
	}
	if conf.Session.MaxAge <= 0 {
		return AppConfig{}, fmt.Errorf("session.max_age must be positive")
	}
	if strings.TrimSpace(conf.Auth.UsernameSalt) == "" {
		return AppConfig{}, fmt.Errorf("auth.username_salt is required")
	}
	return conf, nil
}
