package credential

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"

	"jcourse/pkg/apperr"

	"golang.org/x/crypto/pbkdf2"
)

var ErrPasswordRequired = apperr.ErrPasswordRequired

const (
	djangoPBKDF2SHA256Algorithm  = "pbkdf2_sha256"
	djangoPBKDF2SHA256Iterations = 1200000
	djangoSaltLength             = 22
)

type PasswordHashConfig struct {
	Iterations int
	SaltLength int
}

var DefaultPasswordHashConfig = PasswordHashConfig{
	Iterations: djangoPBKDF2SHA256Iterations,
	SaltLength: djangoSaltLength,
}

type PasswordHasher interface {
	Hash(password string) (string, error)
	Verify(password, encoded string) bool
}

type DjangoPBKDF2SHA256PasswordHasher struct {
	Iterations int
	SaltLength int
}

func NewDjangoPBKDF2SHA256PasswordHasher(config PasswordHashConfig) *DjangoPBKDF2SHA256PasswordHasher {
	defaults := DefaultPasswordHashConfig
	if config.Iterations <= 0 {
		config.Iterations = defaults.Iterations
	}
	if config.SaltLength <= 0 {
		config.SaltLength = defaults.SaltLength
	}
	return &DjangoPBKDF2SHA256PasswordHasher{Iterations: config.Iterations, SaltLength: config.SaltLength}
}

func (h *DjangoPBKDF2SHA256PasswordHasher) Hash(password string) (string, error) {
	salt, err := randomString(h.SaltLength)
	if err != nil {
		return "", err
	}
	return h.hashWithSalt(password, salt), nil
}

func (h *DjangoPBKDF2SHA256PasswordHasher) Verify(password, encoded string) bool {
	parts := strings.Split(encoded, "$")
	if len(parts) != 4 || parts[0] != djangoPBKDF2SHA256Algorithm {
		return false
	}
	iterations, err := strconv.Atoi(parts[1])
	if err != nil || iterations <= 0 {
		return false
	}
	expected := pbkdf2.Key([]byte(password), []byte(parts[2]), iterations, sha256.Size, sha256.New)
	actual, err := base64.StdEncoding.DecodeString(parts[3])
	if err != nil {
		return false
	}
	return subtle.ConstantTimeCompare(expected, actual) == 1
}

func (h *DjangoPBKDF2SHA256PasswordHasher) hashWithSalt(password, salt string) string {
	encoded := pbkdf2.Key([]byte(password), []byte(salt), h.Iterations, sha256.Size, sha256.New)
	return fmt.Sprintf("%s$%d$%s$%s", djangoPBKDF2SHA256Algorithm, h.Iterations, salt, base64.StdEncoding.EncodeToString(encoded))
}

func randomString(n int) (string, error) {
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	out := make([]byte, n)
	for i, b := range buf {
		out[i] = alphabet[int(b)%len(alphabet)]
	}
	return string(out), nil
}
