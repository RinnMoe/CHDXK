package moderation

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	green20220302 "github.com/alibabacloud-go/green-20220302/v3/client"
	"github.com/alibabacloud-go/tea/dara"
	"github.com/alibabacloud-go/tea/tea"
)

const (
	DefaultAliyunGreenRegionID       = "cn-shanghai"
	DefaultAliyunGreenEndpoint       = "green-cip.cn-shanghai.aliyuncs.com"
	DefaultAliyunGreenService        = "comment_detection_pro"
	DefaultConnectTimeout            = 3 * time.Second
	DefaultReadTimeout               = 10 * time.Second
	DefaultSensitiveRiskLevel        = "medium"
	riskLevelHigh                    = "high"
	riskLevelMedium                  = "medium"
	aliyunGreenSuccessCode     int32 = http.StatusOK
)

type AliyunGreenConfig struct {
	Enabled         bool          `mapstructure:"enabled"`
	AccessKeyID     string        `mapstructure:"access_key_id"`
	AccessKeySecret string        `mapstructure:"access_key_secret"`
	RegionID        string        `mapstructure:"region_id"`
	Endpoint        string        `mapstructure:"endpoint"`
	Service         string        `mapstructure:"service"`
	SensitiveLevel  string        `mapstructure:"sensitive_level"`
	ConnectTimeout  time.Duration `mapstructure:"connect_timeout"`
	ReadTimeout     time.Duration `mapstructure:"read_timeout"`
}

var DefaultAliyunGreenConfig = AliyunGreenConfig{
	Enabled:        false,
	RegionID:       DefaultAliyunGreenRegionID,
	Endpoint:       DefaultAliyunGreenEndpoint,
	Service:        DefaultAliyunGreenService,
	SensitiveLevel: DefaultSensitiveRiskLevel,
	ConnectTimeout: DefaultConnectTimeout,
	ReadTimeout:    DefaultReadTimeout,
}

type AliyunGreenModerator struct {
	client  *green20220302.Client
	runtime *dara.RuntimeOptions
	config  AliyunGreenConfig
}

func NewAliyunGreenModerator(config AliyunGreenConfig) (*AliyunGreenModerator, error) {
	config = normalizeAliyunGreenConfig(config)
	if strings.TrimSpace(config.AccessKeyID) == "" {
		return nil, fmt.Errorf("aliyun green access key id is required")
	}
	if strings.TrimSpace(config.AccessKeySecret) == "" {
		return nil, fmt.Errorf("aliyun green access key secret is required")
	}

	client, err := green20220302.NewClient(&openapi.Config{
		AccessKeyId:     tea.String(config.AccessKeyID),
		AccessKeySecret: tea.String(config.AccessKeySecret),
		RegionId:        tea.String(config.RegionID),
		Endpoint:        tea.String(config.Endpoint),
		ConnectTimeout:  tea.Int(durationMillis(config.ConnectTimeout)),
		ReadTimeout:     tea.Int(durationMillis(config.ReadTimeout)),
	})
	if err != nil {
		return nil, fmt.Errorf("create aliyun green client: %w", err)
	}

	return &AliyunGreenModerator{
		client: client,
		runtime: &dara.RuntimeOptions{
			ConnectTimeout: tea.Int(durationMillis(config.ConnectTimeout)),
			ReadTimeout:    tea.Int(durationMillis(config.ReadTimeout)),
		},
		config: config,
	}, nil
}

func (m *AliyunGreenModerator) IsSensitive(ctx context.Context, accountID string, content string) (bool, error) {
	select {
	case <-ctx.Done():
		return false, ctx.Err()
	default:
	}

	params, err := json.Marshal(map[string]any{
		"accountId": accountID,
		"content":   content,
	})
	if err != nil {
		return false, fmt.Errorf("marshal aliyun green service parameters: %w", err)
	}

	request := &green20220302.TextModerationPlusRequest{
		Service:           tea.String(m.config.Service),
		ServiceParameters: tea.String(string(params)),
	}
	response, err := m.client.TextModerationPlusWithOptions(request, m.runtime)
	if err != nil {
		return false, fmt.Errorf("aliyun green text moderation: %w", err)
	}

	if response == nil || response.StatusCode == nil || *response.StatusCode != int32(http.StatusOK) {
		return false, fmt.Errorf("aliyun green text moderation http status: %d", int32Value(responseStatusCode(response)))
	}
	body := response.Body
	if body == nil || body.Code == nil || *body.Code != aliyunGreenSuccessCode {
		return false, fmt.Errorf("aliyun green text moderation code: %d, message: %s", int32Value(bodyCode(body)), stringValue(bodyMessage(body)))
	}
	if body.Data == nil {
		return false, fmt.Errorf("aliyun green text moderation data is empty")
	}

	return riskLevelIsSensitive(stringValue(body.Data.RiskLevel), m.config.SensitiveLevel), nil
}

func normalizeAliyunGreenConfig(config AliyunGreenConfig) AliyunGreenConfig {
	defaults := DefaultAliyunGreenConfig
	if strings.TrimSpace(config.RegionID) == "" {
		config.RegionID = defaults.RegionID
	}
	if strings.TrimSpace(config.Endpoint) == "" {
		config.Endpoint = defaults.Endpoint
	}
	if strings.TrimSpace(config.Service) == "" {
		config.Service = defaults.Service
	}
	if strings.TrimSpace(config.SensitiveLevel) == "" {
		config.SensitiveLevel = defaults.SensitiveLevel
	}
	if config.ConnectTimeout <= 0 {
		config.ConnectTimeout = defaults.ConnectTimeout
	}
	if config.ReadTimeout <= 0 {
		config.ReadTimeout = defaults.ReadTimeout
	}
	return config
}

func durationMillis(d time.Duration) int {
	return int(d / time.Millisecond)
}

func riskLevelIsSensitive(riskLevel string, threshold string) bool {
	riskRank := riskLevelRank(riskLevel)
	thresholdRank := riskLevelRank(threshold)
	return riskRank > 0 && thresholdRank > 0 && riskRank >= thresholdRank
}

func riskLevelRank(level string) int {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case riskLevelHigh:
		return 3
	case riskLevelMedium:
		return 2
	case "low":
		return 1
	case "none", "safe", "pass":
		return 0
	default:
		return 0
	}
}

func responseStatusCode(response *green20220302.TextModerationPlusResponse) *int32 {
	if response == nil {
		return nil
	}
	return response.StatusCode
}

func bodyCode(body *green20220302.TextModerationPlusResponseBody) *int32 {
	if body == nil {
		return nil
	}
	return body.Code
}

func bodyMessage(body *green20220302.TextModerationPlusResponseBody) *string {
	if body == nil {
		return nil
	}
	return body.Message
}

func int32Value(v *int32) int32 {
	if v == nil {
		return 0
	}
	return *v
}

func stringValue(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}
