package moderation

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	green20220302 "github.com/alibabacloud-go/green-20220302/v3/client"

	"jcourse/pkg/logx"
)

type AliyunGreenConfig struct {
	Enabled         bool   `mapstructure:"enabled"`
	AccessKeyID     string `mapstructure:"access_key_id"`
	AccessKeySecret string `mapstructure:"access_key_secret"`
	RegionID        string `mapstructure:"region_id"`
	Endpoint        string `mapstructure:"endpoint"`
	Service         string `mapstructure:"service"`
	SensitiveLevel  string `mapstructure:"sensitive_level"`
	ConnectTimeout  int    `mapstructure:"connect_timeout"`
	ReadTimeout     int    `mapstructure:"read_timeout"`
}

var DefaultAliyunGreenConfig = AliyunGreenConfig{
	Enabled:        false,
	RegionID:       "cn-shanghai",
	Endpoint:       "green-cip.cn-shanghai.aliyuncs.com",
	Service:        "comment_detection_pro",
	SensitiveLevel: "medium",
	ConnectTimeout: 3000,
	ReadTimeout:    10000,
}

type AliyunGreenModerator struct {
	client *green20220302.Client
	config AliyunGreenConfig
}

func NewAliyunGreenModerator(config AliyunGreenConfig) (*AliyunGreenModerator, error) {
	if !config.Enabled {
		return nil, nil
	}

	if strings.TrimSpace(config.AccessKeyID) == "" {
		return nil, fmt.Errorf("aliyun green access key id is required")
	}
	if strings.TrimSpace(config.AccessKeySecret) == "" {
		return nil, fmt.Errorf("aliyun green access key secret is required")
	}

	client, err := green20220302.NewClient(&openapi.Config{
		AccessKeyId:     new(config.AccessKeyID),
		AccessKeySecret: new(config.AccessKeySecret),
		RegionId:        new(config.RegionID),
		Endpoint:        new(config.Endpoint),
		ConnectTimeout:  new(config.ConnectTimeout),
		ReadTimeout:     new(config.ReadTimeout),
	})
	if err != nil {
		return nil, fmt.Errorf("create aliyun green client: %w", err)
	}

	return &AliyunGreenModerator{
		client: client,
		config: config,
	}, nil
}

func (m *AliyunGreenModerator) IsSensitive(ctx context.Context, accountID string, content string) (bool, error) {
	select {
	case <-ctx.Done():
		return false, ctx.Err()
	default:
	}
	if len(content) >= 600 {
		logx.Warn(ctx, "aliyun green text moderation content length is too long", "content", content)
		return false, nil
	}
	params, err := json.Marshal(map[string]any{
		"accountId": accountID,
		"content":   content,
	})
	if err != nil {
		return false, fmt.Errorf("marshal aliyun green service parameters: %w", err)
	}

	request := &green20220302.TextModerationPlusRequest{
		Service:           new(m.config.Service),
		ServiceParameters: new(string(params)),
	}
	response, err := m.client.TextModerationPlus(request)
	if err != nil {
		return false, fmt.Errorf("aliyun green text moderation: %w", err)
	}

	if response == nil || response.StatusCode == nil || *response.StatusCode != int32(http.StatusOK) {
		return false, fmt.Errorf("aliyun green text moderation http status: %d", int32Value(responseStatusCode(response)))
	}
	body := response.Body
	if body == nil || body.Code == nil || *body.Code != http.StatusOK {
		return false, fmt.Errorf("aliyun green text moderation code: %d, message: %s", int32Value(bodyCode(body)), stringValue(bodyMessage(body)))
	}
	if body.Data == nil {
		return false, fmt.Errorf("aliyun green text moderation data is empty")
	}

	return riskLevelIsSensitive(stringValue(body.Data.RiskLevel), m.config.SensitiveLevel), nil
}

func riskLevelIsSensitive(riskLevel string, threshold string) bool {
	riskRank := riskLevelRank(riskLevel)
	thresholdRank := riskLevelRank(threshold)
	return riskRank > 0 && thresholdRank > 0 && riskRank >= thresholdRank
}

func riskLevelRank(level string) int {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "high":
		return 3
	case "medium":
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
