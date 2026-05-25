package config

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/go-viper/mapstructure/v2"
	"github.com/spf13/viper"

	"jcourse/internal/application"
	"jcourse/internal/domain/account"
	"jcourse/internal/domain/auth"
	"jcourse/internal/domain/course"
	"jcourse/internal/domain/point"
	"jcourse/internal/domain/review"
	"jcourse/internal/domain/review/policy"
	"jcourse/internal/domain/stat"
	"jcourse/internal/infrastructure/email"
	"jcourse/internal/infrastructure/persistence"
	infratask "jcourse/internal/infrastructure/task"
	"jcourse/internal/interface/web/middleware"
)

type AppConfig struct {
	Server   ServerConfig                       `mapstructure:"server"`
	Postgres persistence.PostgresConfig         `mapstructure:"postgres"`
	Redis    persistence.RedisConfig            `mapstructure:"redis"`
	Session  middleware.SessionConfig           `mapstructure:"session"`
	Auth     AuthConfig                         `mapstructure:"auth"`
	Review   ReviewConfig                       `mapstructure:"review"`
	APIKey   auth.ApiKeyConfig                  `mapstructure:"api_key"`
	Admin    application.AdminUserCommandConfig `mapstructure:"admin"`
	Point    point.TransferFeeConfig            `mapstructure:"point"`
	Asynq    infratask.Config                   `mapstructure:"asynq"`
	Stats    stat.Config                        `mapstructure:"stats"`
	SMTP     email.SMTPConfig                   `mapstructure:"smtp"`
	JAccount JAccountConfig                     `mapstructure:"jaccount"`
}

type AuthConfig struct {
	Registration    account.RegistrationConfig
	PasswordReset   account.PasswordResetConfig
	Login           account.LoginConfig
	Access          auth.AccessConfig
	PasswordHash    account.PasswordHashConfig
	UsernameDeriver account.UsernameDeriverConfig
}

type ReviewConfig struct {
	Command         application.ReviewCommandConfig
	FrequencyPolicy policy.FrequencyPolicyConfig
}

type ServerConfig struct {
	Addr string `mapstructure:"addr"`
}

type JAccountConfig struct {
	ClientID            string   `mapstructure:"client_id"`
	ClientSecret        string   `mapstructure:"client_secret"`
	AuthorizeURL        string   `mapstructure:"authorize_url"`
	TokenURL            string   `mapstructure:"token_url"`
	APIBaseURL          string   `mapstructure:"api_base_url"`
	RedirectURL         string   `mapstructure:"redirect_url"`
	FrontendCallbackURL string   `mapstructure:"frontend_callback_url"`
	Scopes              []string `mapstructure:"scopes"`
	CourseSyncEnabled   bool     `mapstructure:"course_sync_enabled"`
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
	setDefaults(v)

	if err := v.ReadInConfig(); err != nil {
		return AppConfig{}, fmt.Errorf("read config: %w", err)
	}

	var conf AppConfig
	if err := v.Unmarshal(&conf,
		viper.DecodeHook(mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToTimeDurationHookFunc(),
			mapstructure.StringToSliceHookFunc(","),
		)),
		func(c *mapstructure.DecoderConfig) {
			c.MapFieldName = toSnakeCase
		},
	); err != nil {
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
	if strings.TrimSpace(conf.Auth.UsernameDeriver.Salt) == "" {
		return AppConfig{}, fmt.Errorf("auth.username_deriver.salt is required")
	}
	return conf, nil
}

func setDefaults(v *viper.Viper) {
	setSectionDefaults(v, "session", map[string]any{
		"max_age": middleware.DefaultSessionConfig.MaxAge,
		"secure":  middleware.DefaultSessionConfig.Secure,
	})
	setSectionDefaults(v, "stats", map[string]any{
		"daily_cron":        stat.DefaultConfig.DailyCron,
		"scheduler_enabled": stat.DefaultConfig.SchedulerEnabled,
		"timezone":          stat.DefaultConfig.Timezone,
	})
	setSectionDefaults(v, "auth.registration", map[string]any{
		"email_whitelist": account.DefaultRegistrationConfig.EmailWhitelist,
		"code_interval":   account.DefaultRegistrationConfig.CodeInterval.String(),
		"code_ttl":        account.DefaultRegistrationConfig.CodeTTL.String(),
		"code_length":     account.DefaultRegistrationConfig.CodeLength,
	})
	setSectionDefaults(v, "auth.password_reset", map[string]any{
		"code_interval": account.DefaultPasswordResetConfig.CodeInterval.String(),
		"code_ttl":      account.DefaultPasswordResetConfig.CodeTTL.String(),
		"code_length":   account.DefaultPasswordResetConfig.CodeLength,
	})
	setSectionDefaults(v, "auth.login", map[string]any{
		"max_attempts": account.DefaultLoginConfig.MaxAttempts,
		"lockout":      account.DefaultLoginConfig.Lockout.String(),
	})
	setSectionDefaults(v, "auth.access", map[string]any{
		"flush_cron":        auth.DefaultAccessConfig.FlushCron,
		"scheduler_enabled": auth.DefaultAccessConfig.SchedulerEnabled,
		"flush_batch_size":  auth.DefaultAccessConfig.FlushBatchSize,
	})
	setSectionDefaults(v, "auth.password_hash", map[string]any{
		"iterations":  account.DefaultPasswordHashConfig.Iterations,
		"salt_length": account.DefaultPasswordHashConfig.SaltLength,
	})
	setSectionDefaults(v, "point", map[string]any{
		"rate_bps": point.DefaultTransferFeeConfig.RateBps,
		"min_fee":  point.DefaultTransferFeeConfig.MinFee,
	})
	setSectionDefaults(v, "review.frequency_policy", map[string]any{
		"window":           policy.DefaultFrequencyPolicyConfig.Window.String(),
		"max_reviews":      policy.DefaultFrequencyPolicyConfig.MaxReviews,
		"similarity_ratio": policy.DefaultFrequencyPolicyConfig.SimilarityRatio,
	})
	setSectionDefaults(v, "review.command.hot_scores", map[string]any{
		"review_create_score": course.DefaultHotScoreConfig.ReviewCreateScore,
		"review_update_score": course.DefaultHotScoreConfig.ReviewUpdateScore,
		"review_vote_score":   course.DefaultHotScoreConfig.ReviewVoteScore,
	})
	setSectionDefaults(v, "review.command.vote", map[string]any{
		"max_daily_votes": review.DefaultVoteConfig.MaxDailyVotes,
	})
	setSectionDefaults(v, "api_key", map[string]any{
		"max_user_keys": auth.DefaultApiKeyConfig.MaxUserKeys,
	})
	setSectionDefaults(v, "admin", map[string]any{
		"default_suspend_days": application.DefaultAdminUserCommandConfig.DefaultSuspendDays,
	})
	setSectionDefaults(v, "jaccount", map[string]any{
		"authorize_url":       "https://jaccount.sjtu.edu.cn/oauth2/authorize",
		"token_url":           "https://jaccount.sjtu.edu.cn/oauth2/token",
		"api_base_url":        "https://api.sjtu.edu.cn/",
		"scopes":              []string{},
		"course_sync_enabled": false,
	})
}

func setSectionDefaults(v *viper.Viper, section string, values map[string]any) {
	for key, value := range values {
		v.SetDefault(section+"."+key, value)
	}
}

var snakeCaseBoundary = regexp.MustCompile(`([a-z0-9])([A-Z])`)

func toSnakeCase(s string) string {
	return strings.ToLower(snakeCaseBoundary.ReplaceAllString(s, `${1}_${2}`))
}
