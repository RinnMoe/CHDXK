package bkjw

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"

	"jcourse/internal/domain/course"
)

type jobStore struct {
	root string
	mu   sync.Mutex
}
type savedJob struct {
	UpdatedAt time.Time       `json:"updated_at"`
	Status    json.RawMessage `json:"status"`
}

func NewJobStore(root string) course.SourceJobStore {
	if root == "" {
		return nil
	}
	return &jobStore{root: filepath.Join(root, "jobs")}
}
func (s *jobStore) path(owner int, personal bool) string {
	return filepath.Join(s.root, cacheKey([]any{owner, personal})+".json")
}
func (s *jobStore) Save(ctx context.Context, owner int, personal bool, value json.RawMessage) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if len(value) > 2<<20 || !json.Valid(value) {
		return codeError("CACHE_WRITE_FAILED")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	// A cancelled connection can finish after its replacement starts. Do not
	// let its final progress overwrite the newer job for this owner.
	var incoming struct {
		CreatedAt time.Time `json:"created_at"`
	}
	_ = json.Unmarshal(value, &incoming)
	var previous savedJob
	if readCheckpoint(s.path(owner, personal), &previous, 3<<20) == nil {
		var old struct {
			CreatedAt time.Time `json:"created_at"`
		}
		if json.Unmarshal(previous.Status, &old) == nil && incoming.CreatedAt.Before(old.CreatedAt) {
			return nil
		}
	}
	return writeCheckpoint(s.path(owner, personal), savedJob{time.Now(), value})
}
func (s *jobStore) Load(ctx context.Context, owner int, personal bool) (json.RawMessage, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	var saved savedJob
	err := readCheckpoint(s.path(owner, personal), &saved, 3<<20)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if time.Since(saved.UpdatedAt) > 72*time.Hour || saved.UpdatedAt.After(time.Now()) {
		return nil, nil
	}
	return saved.Status, nil
}
