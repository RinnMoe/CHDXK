package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/snowflake"
)

const (
	ApiKeyRoleSystem = "system"

	apiKeyPrefix       = "jc"
	apiKeySecretLength = 16
	apiKeyIDLength     = 8
)

type ApiKeyConfig struct {
	SnowflakeNodeID int64
}

var DefaultApiKeyConfig = ApiKeyConfig{SnowflakeNodeID: 1}

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

	keyIDBytes, err := hex.DecodeString(parts[1])
	if err != nil || len(keyIDBytes) != apiKeyIDLength {
		return nil, ErrInvalidApiKey
	}
	secret, err := hex.DecodeString(parts[2])
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
	return fmt.Sprintf("%s_%s_%s", apiKeyPrefix, EncodeApiKeyID(c.KeyID), hex.EncodeToString(c.Secret))
}

func (c ApiKeyCredential) HashSecret() string {
	sum := sha256.Sum256(c.Secret)
	return hex.EncodeToString(sum[:])
}

func EncodeApiKeyID(id int64) string {
	var b [apiKeyIDLength]byte
	binary.BigEndian.PutUint64(b[:], uint64(id))
	return hex.EncodeToString(b[:])
}

type ApiKey struct {
	ID         int64
	Name       string
	SecretHash string
	Role       string
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
}

type ApiKeyService struct {
	repo  ApiKeyRepository
	query ApiKeyQuery
	node  *snowflake.Node
}

func NewApiKeyService(repo ApiKeyRepository, query ApiKeyQuery, config ApiKeyConfig) *ApiKeyService {
	if config.SnowflakeNodeID == 0 {
		config.SnowflakeNodeID = DefaultApiKeyConfig.SnowflakeNodeID
	}
	node, err := snowflake.NewNode(config.SnowflakeNodeID)
	if err != nil {
		panic(err)
	}
	return &ApiKeyService{repo: repo, query: query, node: node}
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
	return k.Role == ApiKeyRoleSystem
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

func (s *ApiKeyService) ListSystemKeys(ctx context.Context) ([]ApiKey, error) {
	return s.query.ListSystem(ctx)
}

func (s *ApiKeyService) GetKey(ctx context.Context, id int64) (*ApiKey, error) {
	return s.query.GetByID(ctx, id)
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
