package config

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/go-viper/mapstructure/v2"
	"github.com/spf13/viper"

	"jcourse/internal/application"
	"jcourse/internal/domain/account"
	"jcourse/internal/domain/account/credential"
	"jcourse/internal/domain/account/identity"
	"jcourse/internal/domain/account/verification"
	"jcourse/internal/domain/auth"
	"jcourse/internal/domain/course"
	"jcourse/internal/domain/point"
	"jcourse/internal/domain/review"
	"jcourse/internal/domain/review/policy"
	"jcourse/internal/domain/stat"
	"jcourse/internal/infrastructure/moderation"
	"jcourse/internal/infrastructure/persistence"
	"jcourse/internal/infrastructure/smtp"
	"jcourse/internal/infrastructure/task"
	"jcourse/internal/interface/web/middleware"
)

type AppConfig struct {
	Server   ServerConfig                       `mapstructure:"server"`
	Postgres persistence.PostgresConfig         `mapstructure:"postgres"`
	Redis    persistence.RedisConfig            `mapstructure:"redis"`
	Session  middleware.SessionConfig           `mapstructure:"session"`
	Auth     AuthConfig                         `mapstructure:"auth"`
	Course   CourseConfig                       `mapstructure:"course"`
	Review   ReviewConfig                       `mapstructure:"review"`
	APIKey   auth.ApiKeyConfig                  `mapstructure:"api_key"`
	Admin    application.AdminUserCommandConfig `mapstructure:"admin"`
	Point    point.TransferFeeConfig            `mapstructure:"point"`
	Asynq    task.Config                        `mapstructure:"asynq"`
	Stats    stat.Config                        `mapstructure:"stats"`
	SMTP     smtp.SMTPConfig                    `mapstructure:"smtp"`
	JAccount JAccountConfig                     `mapstructure:"jaccount"`
}

type AuthConfig struct {
	Registration    account.RegistrationConfig
	Verification    verification.Config
	Login           account.LoginConfig
	Access          auth.AccessConfig
	PasswordHash    credential.PasswordHashConfig
	UsernameDeriver identity.UsernameDeriverConfig
}

type CourseConfig struct {
	RatingScore course.RatingScoreConfig `mapstructure:"rating_score"`
}

type ReviewConfig struct {
	Command         application.ReviewCommandConfig `mapstructure:"command"`
	FrequencyPolicy policy.FrequencyPolicyConfig    `mapstructure:"frequency_policy"`
	SafetyPolicy    moderation.AliyunGreenConfig    `mapstructure:"safety_policy"`
}

type ServerConfig struct {
	Addr  string     `mapstructure:"addr"`
	Debug bool       `mapstructure:"debug"`
	Cors  CorsConfig `mapstructure:"cors"`
}

type CorsConfig struct {
	AllowedOrigins []string `mapstructure:"allowed_origins"`
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
	setSectionDefaults(v, "server", map[string]any{
		"addr":  ":8080",
		"debug": false,
	})
	setSectionDefaults(v, "server.cors", map[string]any{
		"allowed_origins": []string{
			"http://localhost:5173",
			"http://127.0.0.1:5173",
		},
	})
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
	})
	setSectionDefaults(v, "auth.verification", map[string]any{
		"code_interval": verification.DefaultConfig.CodeInterval.String(),
		"code_ttl":      verification.DefaultConfig.CodeTTL.String(),
		"code_length":   verification.DefaultConfig.CodeLength,
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
		"iterations":  credential.DefaultPasswordHashConfig.Iterations,
		"salt_length": credential.DefaultPasswordHashConfig.SaltLength,
	})
	setSectionDefaults(v, "point", map[string]any{
		"rate_bps": point.DefaultTransferFeeConfig.RateBps,
		"min_fee":  point.DefaultTransferFeeConfig.MinFee,
	})
	setSectionDefaults(v, "course.rating_score", map[string]any{
		"prior_count":       course.DefaultRatingScoreConfig.PriorCount,
		"refresh_cron":      course.DefaultRatingScoreConfig.RefreshCron,
		"scheduler_enabled": course.DefaultRatingScoreConfig.SchedulerEnabled,
	})
	setSectionDefaults(v, "review.frequency_policy", map[string]any{
		"window":           policy.DefaultFrequencyPolicyConfig.Window.String(),
		"max_reviews":      policy.DefaultFrequencyPolicyConfig.MaxReviews,
		"similarity_ratio": policy.DefaultFrequencyPolicyConfig.SimilarityRatio,
		"suspend_duration": policy.DefaultFrequencyPolicyConfig.SuspendDuration.String(),
	})
	setSectionDefaults(v, "review.safety_policy", map[string]any{
		"enabled":         moderation.DefaultAliyunGreenConfig.Enabled,
		"region_id":       moderation.DefaultAliyunGreenConfig.RegionID,
		"endpoint":        moderation.DefaultAliyunGreenConfig.Endpoint,
		"service":         moderation.DefaultAliyunGreenConfig.Service,
		"sensitive_level": moderation.DefaultAliyunGreenConfig.SensitiveLevel,
		"connect_timeout": moderation.DefaultAliyunGreenConfig.ConnectTimeout,
		"read_timeout":    moderation.DefaultAliyunGreenConfig.ReadTimeout,
	})
	setSectionDefaults(v, "review.command.hot_scores", map[string]any{
		"review_create_score": course.DefaultHotScoreConfig.ReviewCreateScore,
		"review_update_score": course.DefaultHotScoreConfig.ReviewUpdateScore,
		"review_vote_score":   course.DefaultHotScoreConfig.ReviewVoteScore,
	})
	setSectionDefaults(v, "review.command.vote", map[string]any{
		"max_daily_votes": review.DefaultVoteConfig.MaxDailyVotes,
	})
	v.SetDefault("review.command.frequency_violation_admin_emails", []string{})
	setSectionDefaults(v, "api_key", map[string]any{
		"max_user_keys":     auth.DefaultApiKeyConfig.MaxUserKeys,
		"snowflake_node_id": auth.DefaultApiKeyConfig.SnowflakeNodeID,
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
