package config

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/go-viper/mapstructure/v2"
	"github.com/spf13/viper"

	"jcourse/internal/domain/auth"
	"jcourse/internal/domain/canteen"
	"jcourse/internal/domain/canteen/policy"
	"jcourse/internal/domain/course"
	"jcourse/internal/domain/point"
	"jcourse/internal/domain/stat"
	"jcourse/internal/infrastructure/bkjw"
	"jcourse/internal/infrastructure/moderation"
	"jcourse/internal/infrastructure/persistence"
	"jcourse/internal/infrastructure/smtp"
	"jcourse/internal/infrastructure/task"
	"jcourse/internal/interface/web/middleware"
)

type AppConfig struct {
	CourseSource bkjw.Config                `mapstructure:"course_source"`
	Server       ServerConfig               `mapstructure:"server"`
	Postgres     persistence.PostgresConfig `mapstructure:"postgres"`
	Redis        persistence.RedisConfig    `mapstructure:"redis"`
	Session      middleware.SessionConfig   `mapstructure:"session"`
	Auth         AuthConfig                 `mapstructure:"auth"`
	Course       CourseConfig               `mapstructure:"course"`
	Canteen      CanteenConfig              `mapstructure:"canteen"`
	Review       ReviewConfig               `mapstructure:"review"`
	APIKey       auth.ApiKeyConfig          `mapstructure:"api_key"`
	Point        point.TransferFeeConfig    `mapstructure:"point"`
	Asynq        task.Config                `mapstructure:"asynq"`
	Stats        stat.Config                `mapstructure:"stats"`
	SMTP         smtp.SMTPConfig            `mapstructure:"smtp"`
	ResHub       ResHubConfig               `mapstructure:"reshub"`
}

type AuthConfig struct {
	Access auth.AccessConfig
}

type CourseConfig struct {
	RatingScore course.RatingScoreConfig `mapstructure:"rating_score"`
}

type ReviewConfig struct {
	SafetyPolicy moderation.AliyunGreenConfig `mapstructure:"safety_policy"`
}

// CanteenConfig keeps structural knobs in configuration; operator toggles live
// in system settings.
type CanteenConfig struct {
	Rating canteen.RatingConfig `mapstructure:"rating"`
	Review CanteenReviewConfig  `mapstructure:"review"`
	Media  CanteenMediaConfig   `mapstructure:"media"`
}

type CanteenReviewConfig struct {
	Frequency       policy.Config `mapstructure:"frequency"`
	MaxDailyReviews int           `mapstructure:"max_daily_reviews"`
}

type CanteenMediaConfig struct {
	Driver           string `mapstructure:"driver"`
	Dir              string `mapstructure:"dir"`
	MaxBytes         int64  `mapstructure:"max_bytes"`
	MaxPerReview     int    `mapstructure:"max_per_review"`
	MaxUploadsPerDay int    `mapstructure:"max_uploads_per_day"`
	MaxPixels        int    `mapstructure:"max_pixels"`
}

var DefaultCanteenMediaConfig = CanteenMediaConfig{
	Driver:           "local",
	Dir:              ".local/canteen-media",
	MaxBytes:         5 * 1024 * 1024,
	MaxPerReview:     3,
	MaxUploadsPerDay: 20,
	MaxPixels:        8000000,
}

// Normalized fills zero values so a partial config cannot disable media limits.
func (c CanteenMediaConfig) Normalized() CanteenMediaConfig {
	defaults := DefaultCanteenMediaConfig
	if c.Driver == "" {
		c.Driver = defaults.Driver
	}
	if c.Dir == "" {
		c.Dir = defaults.Dir
	}
	if c.MaxBytes <= 0 {
		c.MaxBytes = defaults.MaxBytes
	}
	if c.MaxPerReview <= 0 {
		c.MaxPerReview = defaults.MaxPerReview
	}
	if c.MaxUploadsPerDay <= 0 {
		c.MaxUploadsPerDay = defaults.MaxUploadsPerDay
	}
	if c.MaxPixels <= 0 {
		c.MaxPixels = defaults.MaxPixels
	}
	return c
}

type ServerConfig struct {
	Addr  string     `mapstructure:"addr"`
	Debug bool       `mapstructure:"debug"`
	Cors  CorsConfig `mapstructure:"cors"`
}

type CorsConfig struct {
	AllowedOrigins []string `mapstructure:"allowed_origins"`
}

// ResHubConfig contains only CHDXK-side client configuration. The res-hub
// service remains a separate deployment and owns the protocol implementation.
type ResHubConfig struct {
	Issuer          string        `mapstructure:"issuer"`
	ClientID        string        `mapstructure:"client_id"`
	ClientSecret    string        `mapstructure:"client_secret"`
	RedirectURI     string        `mapstructure:"redirect_uri"`
	LogoutReturnURI string        `mapstructure:"logout_return_uri"`
	Timeout         time.Duration `mapstructure:"timeout"`
	PointsBasePath  string        `mapstructure:"points_base_path"`
	ClientCertFile  string        `mapstructure:"client_cert_file"`
	ClientKeyFile   string        `mapstructure:"client_key_file"`
	CAFile          string        `mapstructure:"ca_file"`
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
	if strings.TrimSpace(conf.ResHub.Issuer) == "" {
		return AppConfig{}, fmt.Errorf("reshub.issuer is required")
	}
	if strings.TrimSpace(conf.ResHub.ClientID) == "" {
		return AppConfig{}, fmt.Errorf("reshub.client_id is required")
	}
	if strings.TrimSpace(conf.ResHub.ClientSecret) == "" {
		return AppConfig{}, fmt.Errorf("reshub.client_secret is required")
	}
	if strings.TrimSpace(conf.ResHub.RedirectURI) == "" {
		return AppConfig{}, fmt.Errorf("reshub.redirect_uri is required")
	}
	if (strings.TrimSpace(conf.ResHub.ClientCertFile) == "") != (strings.TrimSpace(conf.ResHub.ClientKeyFile) == "") {
		return AppConfig{}, fmt.Errorf("reshub.client_cert_file and client_key_file must be configured together")
	}
	if (conf.ResHub.ClientCertFile != "" || conf.ResHub.CAFile != "") && !strings.HasPrefix(conf.ResHub.Issuer, "https://") {
		return AppConfig{}, fmt.Errorf("reshub TLS configuration requires an HTTPS issuer")
	}
	return conf, nil
}

func setDefaults(v *viper.Viper) {
	setSectionDefaults(v, "course_source", map[string]any{"enabled": true, "cache_dir": ".local/course-cache"})
	setSectionDefaults(v, "server", map[string]any{
		"addr":  ":8080",
		"debug": false,
	})
	setSectionDefaults(v, "server.cors", map[string]any{
		"allowed_origins": []string{"*"},
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
	setSectionDefaults(v, "auth.access", map[string]any{
		"flush_cron":        auth.DefaultAccessConfig.FlushCron,
		"scheduler_enabled": auth.DefaultAccessConfig.SchedulerEnabled,
		"flush_batch_size":  auth.DefaultAccessConfig.FlushBatchSize,
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
	setSectionDefaults(v, "canteen.rating", map[string]any{
		"prior_count":       canteen.DefaultRatingConfig.PriorCount,
		"refresh_cron":      canteen.DefaultRatingConfig.RefreshCron,
		"scheduler_enabled": canteen.DefaultRatingConfig.SchedulerEnabled,
	})
	setSectionDefaults(v, "canteen.review", map[string]any{
		"max_daily_reviews": 5,
	})
	setSectionDefaults(v, "canteen.review.frequency", map[string]any{
		"window":           policy.DefaultConfig.Window.String(),
		"max_reviews":      policy.DefaultConfig.MaxReviews,
		"similarity_ratio": policy.DefaultConfig.SimilarityRatio,
	})
	setSectionDefaults(v, "canteen.media", map[string]any{
		"driver":              DefaultCanteenMediaConfig.Driver,
		"dir":                 DefaultCanteenMediaConfig.Dir,
		"max_bytes":           DefaultCanteenMediaConfig.MaxBytes,
		"max_per_review":      DefaultCanteenMediaConfig.MaxPerReview,
		"max_uploads_per_day": DefaultCanteenMediaConfig.MaxUploadsPerDay,
		"max_pixels":          DefaultCanteenMediaConfig.MaxPixels,
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
	setSectionDefaults(v, "api_key", map[string]any{
		"snowflake_node_id": auth.DefaultApiKeyConfig.SnowflakeNodeID,
	})
	setSectionDefaults(v, "reshub", map[string]any{
		"timeout":          3 * time.Second,
		"points_base_path": "/api/integrations/v1",
		"client_cert_file": "",
		"client_key_file":  "",
		"ca_file":          "",
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
