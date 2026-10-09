package bkjw

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"time"
)

// Only public course metadata is stored. Cookies, login URLs, and personal
// timetables are never written to this cache. A new connection still logs in.
type checkpointStore struct{ root string }
type catalogCheckpoint struct {
	Version    int
	CreatedAt  time.Time
	Semester   string
	SemesterID string
	Total      int
	Size       int
	NextPage   int
	Signature  string
	Lessons    []map[string]any
}
type detailCheckpoint struct {
	Generation  time.Time
	Fingerprint string
	Detail      map[string]any
}

const checkpointLifetime = 24 * time.Hour

func cacheKey(value any) string {
	data, _ := json.Marshal(value)
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}
func (store *checkpointStore) dir(s semester) string {
	return filepath.Join(store.root, cacheKey([]string{s.ID, s.Name}))
}
func readCheckpoint(path string, value any, limit int64) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	data, err := io.ReadAll(io.LimitReader(f, limit+1))
	if err != nil {
		return err
	}
	if int64(len(data)) > limit {
		return errors.New("checkpoint too large")
	}
	return json.Unmarshal(data, value)
}
func writeCheckpoint(path string, value any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return codeError("CACHE_WRITE_FAILED")
	}
	f, err := os.CreateTemp(filepath.Dir(path), ".checkpoint-*")
	if err != nil {
		return codeError("CACHE_WRITE_FAILED")
	}
	name := f.Name()
	defer os.Remove(name)
	err = json.NewEncoder(f).Encode(value)
	if err == nil {
		err = f.Sync()
	}
	closeErr := f.Close()
	if err == nil {
		err = closeErr
	}
	if err == nil {
		err = os.Rename(name, path)
	}
	if err != nil {
		return codeError("CACHE_WRITE_FAILED")
	}
	return nil
}
func (store *checkpointStore) load(s semester) (*catalogCheckpoint, error) {
	if store == nil {
		return nil, nil
	}
	var v catalogCheckpoint
	if err := readCheckpoint(filepath.Join(store.dir(s), "catalog.json"), &v, 32<<20); err != nil {
		return nil, nil
	}
	if v.Version != 1 || v.Semester != s.Name || v.SemesterID != s.ID || time.Since(v.CreatedAt) > checkpointLifetime || v.CreatedAt.After(time.Now()) || v.Size < 1 || v.Size > 1000 || v.Total < 0 || v.Total > 100000 || v.NextPage < 2 || v.NextPage > 2001 || v.Signature == "" {
		return nil, nil
	}
	if len(v.Lessons) != min(v.Total, (v.NextPage-1)*v.Size) {
		return nil, nil
	}
	seen := map[string]bool{}
	for _, lesson := range v.Lessons {
		id, ok := lesson["id"].(string)
		links, linksOK := lesson["links"].(map[string]any)
		_, summaryOK := lesson["summary"].(map[string]any)
		path, pathOK := links["detail"].(string)
		if !ok || id == "" || seen[id] || !linksOK || !pathOK || !summaryOK {
			return nil, nil
		}
		if _, err := resolve(eamsOrigin+"/eams/", path); err != nil {
			return nil, nil
		}
		delete(lesson, "detail")
		seen[id] = true
	}
	return &v, nil
}
func (store *checkpointStore) saveCatalog(s semester, v *catalogCheckpoint) error {
	if store == nil {
		return nil
	}
	return writeCheckpoint(filepath.Join(store.dir(s), "catalog.json"), v)
}
func lessonFingerprint(lesson map[string]any) string {
	return cacheKey([]any{lesson["id"], lesson["summary"], lesson["links"]})
}
func (store *checkpointStore) loadDetail(s semester, catalog *catalogCheckpoint, lesson map[string]any) map[string]any {
	if store == nil || catalog == nil {
		return nil
	}
	fingerprint := lessonFingerprint(lesson)
	var v detailCheckpoint
	if readCheckpoint(filepath.Join(store.dir(s), fingerprint+".json"), &v, 1<<20) != nil || !v.Generation.Equal(catalog.CreatedAt) || v.Fingerprint != fingerprint {
		return nil
	}
	if _, ok := v.Detail["basic"].(map[string]any); !ok {
		return nil
	}
	return v.Detail
}
func (store *checkpointStore) saveDetail(s semester, catalog *catalogCheckpoint, lesson map[string]any) error {
	if store == nil {
		return nil
	}
	fingerprint := lessonFingerprint(lesson)
	return writeCheckpoint(filepath.Join(store.dir(s), fingerprint+".json"), detailCheckpoint{Generation: catalog.CreatedAt, Fingerprint: fingerprint, Detail: lesson["detail"].(map[string]any)})
}
