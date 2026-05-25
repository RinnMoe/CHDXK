package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/snowflake"
)

const (
	ApiKeyRoleSystem = "system"
	ApiKeyRoleUser   = "user"

	apiKeyPrefix       = "jc"
	apiKeySecretLength = 16
	apiKeyIDLength     = 8
)

type ApiKeyConfig struct {
	MaxUserKeys     int
	SnowflakeNodeID int64
}

var DefaultApiKeyConfig = ApiKeyConfig{MaxUserKeys: 20, SnowflakeNodeID: 1}

type ApiKeyCredential struct {
	KeyID      int64
	Secret     []byte
	SecretHash string
}

func NewApiKeyCredential(keyID int64) (ApiKeyCredential, error) {
	if keyID <= 0 {
		return ApiKeyCredential{}, ErrInvalidApiKey
	}
	secret := make([]byte, apiKeySecretLength)
	if _, err := rand.Read(secret); err != nil {
		return ApiKeyCredential{}, err
	}
	credential := ApiKeyCredential{KeyID: keyID, Secret: secret}
	credential.SecretHash = credential.HashSecret()
	return credential, nil
}

func ParseApiKeyCredential(key string) (*ApiKeyCredential, error) {
	parts := strings.Split(strings.TrimSpace(key), "_")
	if len(parts) != 3 || parts[0] != apiKeyPrefix {
		return nil, ErrInvalidApiKey
	}

	keyIDBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil || len(keyIDBytes) != apiKeyIDLength {
		return nil, ErrInvalidApiKey
	}
	secret, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil || len(secret) != apiKeySecretLength {
		return nil, ErrInvalidApiKey
	}

	credential := ApiKeyCredential{
		KeyID:  int64(binary.BigEndian.Uint64(keyIDBytes)),
		Secret: secret,
	}
	if credential.KeyID <= 0 {
		return nil, ErrInvalidApiKey
	}
	credential.SecretHash = credential.HashSecret()
	return &credential, nil
}

func (c ApiKeyCredential) Key() string {
	return fmt.Sprintf("%s_%s_%s", apiKeyPrefix, EncodeApiKeyID(c.KeyID), base64.RawURLEncoding.EncodeToString(c.Secret))
}

func (c ApiKeyCredential) HashSecret() string {
	sum := sha256.Sum256(c.Secret)
	return hex.EncodeToString(sum[:])
}

func EncodeApiKeyID(id int64) string {
	var b [apiKeyIDLength]byte
	binary.BigEndian.PutUint64(b[:], uint64(id))
	return base64.RawURLEncoding.EncodeToString(b[:])
}

type ApiKey struct {
	ID         int64
	Name       string
	SecretHash string
	Role       string
	UserID     int
	CreatedAt  time.Time
	LastUsedAt *time.Time
}

func (k ApiKey) MaskedKey() string {
	return fmt.Sprintf("%s_%s_********", apiKeyPrefix, EncodeApiKeyID(k.ID))
}

type ApiKeyRepository interface {
	Create(ctx context.Context, apiKey *ApiKey) error
	Delete(ctx context.Context, id int64) (bool, error)
}

type ApiKeyQuery interface {
	GetByID(ctx context.Context, id int64) (*ApiKey, error)
	ListSystem(ctx context.Context) ([]ApiKey, error)
	ListByUser(ctx context.Context, userID int) ([]ApiKey, error)
	CountByUser(ctx context.Context, userID int) (int, error)
}

type ApiKeyService struct {
	repo   ApiKeyRepository
	query  ApiKeyQuery
	node   *snowflake.Node
	config ApiKeyConfig
}

func NewApiKeyService(repo ApiKeyRepository, query ApiKeyQuery, config ApiKeyConfig) *ApiKeyService {
	if config.MaxUserKeys <= 0 {
		config.MaxUserKeys = DefaultApiKeyConfig.MaxUserKeys
	}
	if config.SnowflakeNodeID == 0 {
		config.SnowflakeNodeID = DefaultApiKeyConfig.SnowflakeNodeID
	}
	node, err := snowflake.NewNode(config.SnowflakeNodeID)
	if err != nil {
		panic(err)
	}
	return &ApiKeyService{repo: repo, query: query, node: node, config: config}
}

func NewUserApiKey(name string, userID int, credential ApiKeyCredential, now time.Time) (*ApiKey, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, ErrApiKeyNameRequired
	}
	return &ApiKey{
		ID:         credential.KeyID,
		Name:       name,
		SecretHash: credential.SecretHash,
		Role:       ApiKeyRoleUser,
		UserID:     userID,
		CreatedAt:  now,
	}, nil
}

func NewSystemApiKey(name string, credential ApiKeyCredential, now time.Time) (*ApiKey, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, ErrApiKeyNameRequired
	}
	return &ApiKey{
		ID:         credential.KeyID,
		Name:       name,
		SecretHash: credential.SecretHash,
		Role:       ApiKeyRoleSystem,
		CreatedAt:  now,
	}, nil
}

func (k ApiKey) IsSystem() bool {
	return k.Role == ApiKeyRoleSystem && k.UserID == 0
}

func (k ApiKey) IsUser() bool {
	return k.Role == ApiKeyRoleUser && k.UserID > 0
}

func (k ApiKey) BelongsTo(userID int) bool {
	return k.IsUser() && k.UserID == userID
}

func (s *ApiKeyService) ValidateKey(ctx context.Context, key string) (*ApiKey, error) {
	credential, err := ParseApiKeyCredential(key)
	if err != nil {
		return nil, nil
	}
	apiKey, err := s.query.GetByID(ctx, credential.KeyID)
	if err != nil {
		return nil, err
	}
	if apiKey == nil {
		return nil, nil
	}
	if subtle.ConstantTimeCompare([]byte(apiKey.SecretHash), []byte(credential.SecretHash)) != 1 {
		return nil, nil
	}
	return apiKey, nil
}

func (s *ApiKeyService) ListUserKeys(ctx context.Context, userID int) ([]ApiKey, error) {
	return s.query.ListByUser(ctx, userID)
}

func (s *ApiKeyService) ListSystemKeys(ctx context.Context) ([]ApiKey, error) {
	return s.query.ListSystem(ctx)
}

func (s *ApiKeyService) CreateSystemKey(ctx context.Context, name string) (*ApiKey, *ApiKeyCredential, error) {
	credential, err := s.generateCredential()
	if err != nil {
		return nil, nil, err
	}
	apiKey, err := NewSystemApiKey(name, credential, time.Now())
	if err != nil {
		return nil, nil, err
	}
	if err := s.repo.Create(ctx, apiKey); err != nil {
		return nil, nil, err
	}
	return apiKey, &credential, nil
}

func (s *ApiKeyService) CreateUserKey(ctx context.Context, userID int, name string) (*ApiKey, *ApiKeyCredential, error) {
	credential, err := s.generateCredential()
	if err != nil {
		return nil, nil, err
	}
	apiKey, err := NewUserApiKey(name, userID, credential, time.Now())
	if err != nil {
		return nil, nil, err
	}

	count, err := s.query.CountByUser(ctx, userID)
	if err != nil {
		return nil, nil, err
	}
	if count >= s.config.MaxUserKeys {
		return nil, nil, ErrApiKeyLimitExceeded
	}
	if err := s.repo.Create(ctx, apiKey); err != nil {
		return nil, nil, err
	}
	return apiKey, &credential, nil
}

func (s *ApiKeyService) DeleteUserKey(ctx context.Context, userID int, id int64) error {
	key, err := s.query.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if key == nil || !key.BelongsTo(userID) {
		return ErrApiKeyNotFound
	}
	deleted, err := s.repo.Delete(ctx, id)
	if err != nil {
		return err
	}
	if !deleted {
		return ErrApiKeyNotFound
	}
	return nil
}

func (s *ApiKeyService) DeleteSystemKey(ctx context.Context, id int64) error {
	key, err := s.query.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if key == nil || !key.IsSystem() {
		return ErrApiKeyNotFound
	}
	deleted, err := s.repo.Delete(ctx, id)
	if err != nil {
		return err
	}
	if !deleted {
		return ErrApiKeyNotFound
	}
	return nil
}

func (s *ApiKeyService) generateCredential() (ApiKeyCredential, error) {
	return NewApiKeyCredential(int64(s.node.Generate()))
}
