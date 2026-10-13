package bkjw

import (
	"bytes"
	"context"
	"errors"
	"golang.org/x/net/html"
	"golang.org/x/net/html/charset"
	"io"
	"log/slog"
	"math/rand/v2"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

const casOrigin = "https://ids.chd.edu.cn"
const casLandingPath = "/personalInfo/personCenter/index.html"
const eamsOrigin = "http://bkjw.chd.edu.cn"
const wxOrigin = "https://open.weixin.qq.com"
const maxBody = 16 << 20

func allowedURL(raw string) bool {
	u, err := url.Parse(raw)
	if err != nil || u.User != nil {
		return false
	}
	if port := u.Port(); port != "" && !((u.Scheme == "https" && port == "443") || (u.Scheme == "http" && port == "80")) {
		return false
	}
	switch u.Hostname() {
	case "bkjw.chd.edu.cn":
		return (u.Scheme == "http" || u.Scheme == "https") && strings.HasPrefix(u.Path, "/eams/")
	case "ids.chd.edu.cn":
		return u.Scheme == "https" && (strings.HasPrefix(u.Path, "/authserver/") || u.Path == casLandingPath)
	case "open.weixin.qq.com", "lp.open.weixin.qq.com":
		return u.Scheme == "https"
	}
	return false
}
func checkRedirect(req *http.Request, via []*http.Request) error {
	if len(via) >= 12 || !allowedURL(req.URL.String()) {
		slog.Warn("academic redirect rejected", "scheme", req.URL.Scheme, "host", req.URL.Hostname(), "port", req.URL.Port(), "redirect_count", len(via))
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
		target := u.ResolveReference(r)
		slog.Warn("academic link rejected", "scheme", target.Scheme, "host", target.Hostname(), "port", target.Port())
		return "", codeError("URL_NOT_ALLOWED")
	}
	return full, nil
}
func (c *connection) request(method, raw string, form url.Values, headers map[string]string) ([]byte, string, error) {
	if !allowedURL(raw) {
		return nil, "", codeError("URL_NOT_ALLOWED")
	}
	u, _ := url.Parse(raw)
	if c.budget != nil && u.Hostname() == "bkjw.chd.edu.cn" {
		finish, err := c.budget.begin(c.ctx)
		if err != nil {
			return nil, "", err
		}
		defer finish()
	}
	if c.pacer != nil && u.Hostname() == "bkjw.chd.edu.cn" {
		finish, err := c.pacer.begin(c.ctx)
		if err != nil {
			return nil, "", err
		}
		defer finish()
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
	started := time.Now()
	resp, err := c.client.Do(req)
	if err != nil {
		var networkError net.Error
		timedOut := errors.As(err, &networkError) && networkError.Timeout()
		// WeChat long polling normally times out while waiting for a scan.
		if !(req.URL.Hostname() == "lp.open.weixin.qq.com" && timedOut && c.ctx.Err() == nil) {
			slog.Warn("academic HTTP request failed", "host", req.URL.Hostname(), "timeout", timedOut, "cancelled", c.ctx.Err() != nil, "method", method, "path", u.Path, "elapsed_ms", time.Since(started).Milliseconds())
		}
		return nil, "", err
	}
	defer resp.Body.Close()
	if c.pacer != nil && u.Hostname() != "bkjw.chd.edu.cn" && resp.Request.URL.Hostname() == "bkjw.chd.edu.cn" {
		// A CAS request can finish on EAMS through HTTP redirects. Account
		// for that response before the first authenticated business query.
		defer func() {
			c.pacer.mu.Lock()
			c.pacer.next = maxTime(c.pacer.next, time.Now().Add(c.pacer.interval))
			c.pacer.mu.Unlock()
		}()
	}
	if resp.StatusCode == 429 || resp.StatusCode == 503 {
		c.retryAfter = retryAfter(resp.Header.Get("Retry-After"), time.Now())
		if c.budget != nil && u.Hostname() == "bkjw.chd.edu.cn" {
			c.budget.observe(true, c.retryAfter)
		}
	}
	if resp.StatusCode == 429 {
		if c.pacer != nil {
			c.pacer.slowDown()
		}
		return nil, "", codeError("RATE_LIMITED")
	}
	if resp.StatusCode == 401 {
		return nil, "", codeError("AUTH_REQUIRED")
	}
	if resp.StatusCode == 403 {
		return nil, "", codeError("PERMISSION_DENIED")
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		slog.Warn("academic HTTP response failed", "host", req.URL.Hostname(), "status", resp.StatusCode)
		if resp.StatusCode >= 500 || resp.StatusCode == 408 {
			return nil, "", codeError("UPSTREAM_UNAVAILABLE")
		}
		return nil, "", codeError("UPSTREAM_REJECTED")
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, maxBody+1))
	if err != nil {
		return nil, "", err
	}
	if len(data) > maxBody {
		return nil, "", codeError("PAGE_STRUCTURE")
	}
	contentType := strings.ToLower(resp.Header.Get("Content-Type"))
	if strings.HasPrefix(contentType, "text/") || strings.Contains(contentType, "javascript") || strings.Contains(contentType, "json") {
		reader, decodeErr := charset.NewReader(bytes.NewReader(data), contentType)
		if decodeErr != nil {
			return nil, "", codeError("PAGE_STRUCTURE")
		}
		data, decodeErr = io.ReadAll(io.LimitReader(reader, maxBody+1))
		if decodeErr != nil || len(data) > maxBody {
			return nil, "", codeError("PAGE_STRUCTURE")
		}
	}
	slog.Debug("academic HTTP response", "host", u.Hostname(), "path", u.Path, "method", method, "status", resp.StatusCode, "elapsed_ms", time.Since(started).Milliseconds(), "response_bytes", len(data))
	return data, resp.Request.URL.String(), nil
}

// Serialize EAMS requests and pace them from response completion, including
// retries and the first business request after CAS redirects.
type requestPacer struct {
	gate     sync.Mutex
	mu       sync.Mutex
	next     time.Time
	interval time.Duration
	adaptive bool
	stable   int
}

func (p *requestPacer) begin(ctx context.Context) (func(), error) {
	p.gate.Lock()
	p.mu.Lock()
	next := p.next
	p.mu.Unlock()
	timer := time.NewTimer(max(time.Until(next), 0))
	defer timer.Stop()
	select {
	case <-ctx.Done():
		p.gate.Unlock()
		return nil, ctx.Err()
	case <-timer.C:
	}
	return func() {
		p.mu.Lock()
		p.next = maxTime(p.next, time.Now().Add(p.interval))
		p.mu.Unlock()
		p.gate.Unlock()
	}, nil
}
func maxTime(a, b time.Time) time.Time {
	if a.After(b) {
		return a
	}
	return b
}
func (p *requestPacer) slowDown() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.stable = 0
	p.interval = min(max(p.interval*2, 1250*time.Millisecond), 10*time.Second)
	p.next = time.Now().Add(p.interval)
}
func (c *connection) eams(method, path string, form url.Values) ([]byte, string, error) {
	data, final, err := c.eamsOnce(method, path, form)
	// Renew EAMS from the still-valid CAS jar once per retry operation. No
	// credentials or browser cookies are persisted, and a real login is never guessed.
	if err == codeError("AUTH_REQUIRED") && !c.refreshTried {
		c.refreshTried = true
		_, landing, refreshErr := c.navigate(loginURL())
		if refreshErr != nil {
			return nil, "", refreshErr
		}
		target, _ := url.Parse(landing)
		if target.Hostname() != "bkjw.chd.edu.cn" {
			return nil, "", err
		}
		return c.eamsOnce(method, path, form)
	}
	return data, final, err
}
func (c *connection) eamsOnce(method, path string, form url.Values) ([]byte, string, error) {
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
	requested, _ := url.Parse(raw)
	if u.Path == "/eams/home.action" && requested.Path != u.Path {
		// CAS can renew an expired EAMS session and land on home rather than
		// replaying the original read-only query. Retry with the updated jar.
		return nil, "", codeError("SESSION_REFRESHED")
	}
	if isRateLimitPage(data) {
		if c.pacer != nil {
			c.pacer.slowDown()
		}
		if c.budget != nil {
			c.budget.observe(true, time.Time{})
		}
		slog.Warn("academic request rate limited", "response_bytes", len(data))
		return nil, "", codeError("RATE_LIMITED")
	}
	if isPermissionPage(data) {
		return nil, "", codeError("PERMISSION_DENIED")
	}
	if isMaintenancePage(data) {
		if c.budget != nil {
			c.budget.observe(true, time.Time{})
		}
		return nil, "", codeError("UPSTREAM_UNAVAILABLE")
	}
	if c.pacer != nil {
		c.pacer.success()
	}
	if c.budget != nil {
		c.budget.observe(false, time.Time{})
	}
	return data, final, nil
}

// Match the school's standalone error notice and back link, not phrases
// embedded in scripts or normal course content.
func isRateLimitPage(data []byte) bool {
	// Ordinary pages need no second DOM parse for an absent error notice.
	if !strings.Contains(string(data), "返回前页") {
		return false
	}
	doc := document(data)
	if strings.Contains(text(doc), "课程代码") {
		return false
	}
	back := false
	for _, a := range tag(doc, "a") {
		if strings.Trim(text(a), "[] ") == "返回前页" {
			back = true
		}
	}
	if !back {
		return false
	}
	for _, n := range nodes(doc, func(n *html.Node) bool { return n.Type == html.TextNode }) {
		if n.Parent != nil && (n.Parent.Data == "script" || n.Parent.Data == "style") {
			continue
		}
		switch strings.TrimSpace(n.Data) {
		case "请不要过快点击", "访问过于频繁", "操作过于频繁", "请求过于频繁", "请不要重复提交":
			return true
		}
	}
	return false
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
	started := time.Now()
	c.refreshTried = false
	c.retryAfter = time.Time{}
	defer func() { c.retryAt = nil }()
	attempts := 50
	if c.shortRetries {
		attempts = 3
	}
	for attempt := 0; attempt < attempts && time.Since(started) < 20*time.Minute; attempt++ {
		if c.ctx.Err() != nil {
			return c.ctx.Err()
		}
		if attempt > 0 {
			ceiling := min(2*time.Second<<min(attempt-1, 5), time.Minute)
			delay := ceiling/2 + time.Duration(rand.Int64N(int64(ceiling/2)+1))
			if c.retryDelay != nil {
				delay = c.retryDelay(attempt)
			}
			delay = max(delay, time.Until(c.retryAfter))
			next := time.Now().Add(delay)
			c.retryAt = &next
			if onRetry != nil {
				onRetry(attempt)
			}
			if err = c.pause(delay); err != nil {
				return err
			}
			c.retryAt = nil
		}
		err = operation()
		if err == nil {
			return nil
		}
		var permanent permanentSourceError
		if errors.As(err, &permanent) {
			return err
		}
		var code codeError
		if errors.As(err, &code) {
			switch code {
			case "AUTH_REQUIRED", "PERMISSION_DENIED":
				if attempt >= 2 {
					return err
				}
			case "PAGE_STRUCTURE", "PERSONAL_PARSE_FAILED":
				if attempt >= 7 {
					return err
				}
			case "RATE_LIMITED", "SESSION_REFRESHED", "UPSTREAM_UNAVAILABLE":
			default:
				return err
			}
		}
		if c.ctx.Err() != nil {
			return c.ctx.Err()
		}
	}
	return err
}

func retryAfter(value string, now time.Time) time.Time {
	if seconds, err := strconv.Atoi(strings.TrimSpace(value)); err == nil && seconds >= 0 {
		return now.Add(time.Duration(min(seconds, 1200)) * time.Second)
	}
	if t, err := http.ParseTime(value); err == nil {
		return maxTime(now, minTime(t, now.Add(20*time.Minute)))
	}
	return time.Time{}
}
func minTime(a, b time.Time) time.Time {
	if a.Before(b) {
		return a
	}
	return b
}
func (p *requestPacer) success() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if !p.adaptive {
		return
	}
	p.stable++
	if p.stable >= 12 {
		p.stable = 0
		p.interval = max(150*time.Millisecond, p.interval-50*time.Millisecond)
	}
}
func isMaintenancePage(data []byte) bool {
	notices := []string{"系统维护中", "系统正在维护", "系统繁忙", "服务暂不可用", "Service Unavailable"}
	candidate := false
	for _, notice := range notices {
		if strings.Contains(string(data), notice) {
			candidate = true
			break
		}
	}
	if !candidate {
		return false
	}
	doc := document(data)
	for _, script := range tag(doc, "script") {
		if script.Parent != nil {
			script.Parent.RemoveChild(script)
		}
	}
	for _, style := range tag(doc, "style") {
		if style.Parent != nil {
			style.Parent.RemoveChild(style)
		}
	}
	visible := text(doc)
	if len([]rune(visible)) > 300 {
		return false
	}
	for _, notice := range notices {
		if strings.Contains(visible, notice) && len(tag(doc, "table")) == 0 && len(tag(doc, "form")) == 0 {
			return true
		}
	}
	return false
}

func isPermissionPage(data []byte) bool {
	content := string(data)
	if !strings.Contains(content, "没有权限") && !strings.Contains(content, "无权访问") {
		return false
	}
	doc := document(data)
	for _, n := range append(tag(doc, "script"), tag(doc, "style")...) {
		if n.Parent != nil {
			n.Parent.RemoveChild(n)
		}
	}
	visible := text(doc)
	// Reject only an actual short permission notice, not a script message or
	// these words in course descriptions. HTTP 403 is classified separately.
	if len([]rune(visible)) > 300 || strings.Contains(visible, "课程代码") || strings.Contains(visible, "课程序号") {
		return false
	}
	for _, table := range tag(doc, "table") {
		if hasClass(table, "gridtable") {
			return false
		}
	}
	return strings.Contains(visible, "没有权限") || strings.Contains(visible, "无权访问")
}

type permanentSourceError struct{ err error }

func (e permanentSourceError) Error() string { return e.err.Error() }
func (e permanentSourceError) Unwrap() error { return e.err }
