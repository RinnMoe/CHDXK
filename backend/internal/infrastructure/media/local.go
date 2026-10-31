package media

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"jcourse/internal/domain/canteen"
)

// LocalStorage stores media bytes under a fixed root directory. Keys are
// relative POSIX-style paths such as "review/2026/10/uuid.jpg".
type LocalStorage struct {
	root string
}

func NewLocalStorage(root string) (*LocalStorage, error) {
	root = strings.TrimSpace(root)
	if root == "" {
		return nil, errors.New("media directory is required")
	}
	absolute, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("resolve media directory: %w", err)
	}
	if err := os.MkdirAll(absolute, 0o750); err != nil {
		return nil, fmt.Errorf("create media directory: %w", err)
	}
	return &LocalStorage{root: absolute}, nil
}

func (s *LocalStorage) resolve(key string) (string, error) {
	key = strings.TrimSpace(key)
	if key == "" {
		return "", errors.New("media key is required")
	}
	cleaned := filepath.Clean(filepath.FromSlash(key))
	if cleaned == "." || filepath.IsAbs(cleaned) || strings.HasPrefix(cleaned, "..") {
		return "", errors.New("media key escapes the storage root")
	}
	path := filepath.Join(s.root, cleaned)
	relative, err := filepath.Rel(s.root, path)
	if err != nil || strings.HasPrefix(relative, "..") {
		return "", errors.New("media key escapes the storage root")
	}
	return path, nil
}

func (s *LocalStorage) Put(_ context.Context, key string, reader io.Reader) error {
	path, err := s.resolve(key)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return err
	}
	temp, err := os.CreateTemp(filepath.Dir(path), ".upload-*")
	if err != nil {
		return err
	}
	tempName := temp.Name()
	defer func() {
		_ = os.Remove(tempName)
	}()
	if _, err := io.Copy(temp, reader); err != nil {
		_ = temp.Close()
		return err
	}
	if err := temp.Chmod(0o640); err != nil {
		_ = temp.Close()
		return err
	}
	if err := temp.Close(); err != nil {
		return err
	}
	return os.Rename(tempName, path)
}

func (s *LocalStorage) Open(_ context.Context, key string) (io.ReadCloser, error) {
	path, err := s.resolve(key)
	if err != nil {
		return nil, err
	}
	return os.Open(path)
}

func (s *LocalStorage) Delete(_ context.Context, key string) error {
	path, err := s.resolve(key)
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

var _ canteen.MediaStorage = (*LocalStorage)(nil)
