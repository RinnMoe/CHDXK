package chdteachers

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"jcourse/internal/domain/teacher"
)

type collectionKey struct{}
type collectionOptions struct {
	fresh    bool
	progress func(teacher.FetchProgress)
}

func (c *Client) CollectionContext(ctx context.Context, fresh bool, progress func(teacher.FetchProgress)) context.Context {
	return context.WithValue(ctx, collectionKey{}, collectionOptions{fresh, progress})
}
func notify(ctx context.Context, p teacher.FetchProgress) {
	if options, ok := ctx.Value(collectionKey{}).(collectionOptions); ok && options.progress != nil {
		options.progress(p)
	}
}
func validResponse(data []byte, body url.Values) error {
	if body == nil {
		_, err := ParseDetail(data, teacher.Profile{})
		return err
	}
	var page struct {
		Total int `json:"total"`
		Pages int `json:"pageCount"`
		Data  []struct {
			Name       string `json:"title"`
			Department string `json:"department"`
			URL        string `json:"cnUrl"`
			Published  int    `json:"publishStatus"`
		} `json:"data"`
	}
	if json.Unmarshal(data, &page) != nil || page.Data == nil || page.Total < 0 || page.Pages < 0 || page.Pages > 500 || (page.Total > 0 && (page.Pages == 0 || len(page.Data) == 0)) {
		return fmt.Errorf("教师目录格式异常")
	}
	if page.Total == 0 {
		var conditions []struct {
			Field string `json:"field"`
			Value string `json:"value"`
		}
		_ = json.Unmarshal([]byte(body.Get("conditions")), &conditions)
		search := false
		for _, condition := range conditions {
			if condition.Field == "title" && strings.TrimSpace(condition.Value) != "" {
				search = true
			}
		}
		if !search || len(page.Data) != 0 {
			return fmt.Errorf("教师目录意外为空")
		}
	}
	for _, p := range page.Data {
		if p.Published != 1 {
			continue
		}
		if _, err := sourceURL(p.URL); err != nil || p.Name == "" || p.Department == "" {
			return fmt.Errorf("教师目录资料不完整")
		}
	}
	return nil
}
func (c *Client) cachePath(address string, body url.Values) string {
	hash := sha256.Sum256([]byte(address + "\n" + body.Encode()))
	return filepath.Join(c.CacheDir, hex.EncodeToString(hash[:])+".json")
}

type cachedResponse struct {
	CreatedAt time.Time
	Data      []byte
}

func (c *Client) cached(ctx context.Context, address string, body url.Values) []byte {
	options, _ := ctx.Value(collectionKey{}).(collectionOptions)
	if c.CacheDir == "" || options.fresh {
		return nil
	}
	path := c.cachePath(address, body)
	info, err := os.Stat(path)
	if err != nil || info.Size() > 3<<20 {
		return nil
	}
	data, err := os.ReadFile(path)
	var v cachedResponse
	if err != nil || json.Unmarshal(data, &v) != nil || time.Since(v.CreatedAt) > 24*time.Hour || v.CreatedAt.After(time.Now()) || validResponse(v.Data, body) != nil {
		return nil
	}
	return v.Data
}
func (c *Client) saveCache(address string, body url.Values, data []byte) error {
	if c.CacheDir == "" {
		return nil
	}
	if err := os.MkdirAll(c.CacheDir, 0700); err != nil {
		return fmt.Errorf("教师采集缓存无法保存: %w", err)
	}
	f, err := os.CreateTemp(c.CacheDir, ".checkpoint-*")
	if err != nil {
		return fmt.Errorf("教师采集缓存无法保存: %w", err)
	}
	defer os.Remove(f.Name())
	err = json.NewEncoder(f).Encode(cachedResponse{time.Now(), data})
	if err == nil {
		err = f.Sync()
	}
	closeErr := f.Close()
	if err == nil {
		err = closeErr
	}
	if err == nil {
		err = os.Rename(f.Name(), c.cachePath(address, body))
	}
	if err != nil {
		return fmt.Errorf("教师采集缓存无法保存: %w", err)
	}
	return nil
}
