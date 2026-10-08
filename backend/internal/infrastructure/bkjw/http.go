package bkjw

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const casOrigin = "https://ids.chd.edu.cn"
const eamsOrigin = "http://bkjw.chd.edu.cn"
const wxOrigin = "https://open.weixin.qq.com"
const maxBody = 16 << 20

func allowedURL(raw string) bool {
	u, err := url.Parse(raw)
	if err != nil || u.User != nil || u.Port() != "" {
		return false
	}
	switch u.Hostname() {
	case "bkjw.chd.edu.cn":
		return (u.Scheme == "http" || u.Scheme == "https") && strings.HasPrefix(u.Path, "/eams/")
	case "ids.chd.edu.cn":
		return u.Scheme == "https" && strings.HasPrefix(u.Path, "/authserver/")
	case "open.weixin.qq.com", "lp.open.weixin.qq.com":
		return u.Scheme == "https"
	}
	return false
}
func checkRedirect(req *http.Request, via []*http.Request) error {
	if len(via) >= 12 || !allowedURL(req.URL.String()) {
		return codeError("URL_NOT_ALLOWED")
	}
	return nil
}
func resolve(base, ref string) (string, error) {
	u, err := url.Parse(base)
	if err != nil {
		return "", err
	}
	r, err := url.Parse(ref)
	if err != nil {
		return "", err
	}
	full := u.ResolveReference(r).String()
	if !allowedURL(full) {
		return "", codeError("URL_NOT_ALLOWED")
	}
	return full, nil
}
func (c *connection) request(method, raw string, form url.Values, headers map[string]string) ([]byte, string, error) {
	if !allowedURL(raw) {
		return nil, "", codeError("URL_NOT_ALLOWED")
	}
	var body io.Reader
	if form != nil {
		body = strings.NewReader(form.Encode())
	}
	requestCtx := c.ctx
	if strings.HasPrefix(raw, "https://lp.open.weixin.qq.com/") {
		var cancel context.CancelFunc
		requestCtx, cancel = context.WithTimeout(c.ctx, 10*time.Second)
		defer cancel()
	}
	req, err := http.NewRequestWithContext(requestCtx, method, raw, body)
	if err != nil {
		return nil, "", err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 CHDXK/1.0")
	if form != nil {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode == 401 {
		return nil, "", codeError("AUTH_REQUIRED")
	}
	if resp.StatusCode == 403 {
		return nil, "", codeError("PERMISSION_DENIED")
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, "", fmt.Errorf("upstream status %d", resp.StatusCode)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, maxBody+1))
	if err != nil {
		return nil, "", err
	}
	if len(data) > maxBody {
		return nil, "", codeError("PAGE_STRUCTURE")
	}
	return data, resp.Request.URL.String(), nil
}
func (c *connection) eams(method, path string, form url.Values) ([]byte, string, error) {
	raw, err := resolve(eamsOrigin+"/eams/", path)
	if err != nil {
		return nil, "", err
	}
	data, final, err := c.request(method, raw, form, map[string]string{"Referer": eamsOrigin + "/eams/home.action", "X-Requested-With": "XMLHttpRequest"})
	if err != nil {
		return nil, "", err
	}
	u, _ := url.Parse(final)
	if u.Hostname() != "bkjw.chd.edu.cn" || strings.Contains(u.Path, "/login") || strings.Contains(string(data), "pwdEncryptSalt") {
		return nil, "", codeError("AUTH_REQUIRED")
	}
	if strings.Contains(string(data), "没有权限") || strings.Contains(string(data), "无权访问") {
		return nil, "", codeError("PERMISSION_DENIED")
	}
	return data, final, nil
}
func (c *connection) pause(d time.Duration) error {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-c.ctx.Done():
		return c.ctx.Err()
	case <-timer.C:
		return nil
	}
}
func (c *connection) retry(operation func() error, onRetry func(int)) error {
	var err error
	for attempt := 0; attempt < 3; attempt++ {
		if attempt > 0 {
			onRetry(attempt)
			if err = c.pause(time.Duration(attempt) * time.Second); err != nil {
				return err
			}
		}
		err = operation()
		if err == nil {
			return nil
		}
		if _, fatal := err.(codeError); fatal || c.ctx.Err() != nil {
			return err
		}
	}
	return err
}
