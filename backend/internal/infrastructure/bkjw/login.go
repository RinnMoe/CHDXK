package bkjw

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"html"
	"jcourse/internal/domain/course"
	"net/url"
	"regexp"
	"strings"
	"time"
)

var wxLink = regexp.MustCompile(`https://open\.weixin\.qq\.com[^\s"'<>]+`)
var uuidPattern = regexp.MustCompile(`qrconnect\?uuid=([a-zA-Z0-9_-]{10,})`)
var callbackPattern = regexp.MustCompile(`(?:var\s+t\s*=|redirect_uri["']?\s*[=:])\s*["']([^"']+)["']`)
var statePattern = regexp.MustCompile(`state=([a-zA-Z0-9_-]+)`)
var codePattern = regexp.MustCompile(`window\.wx_code\s*=\s*['"]([^'"]+)['"]`)
var statusPattern = regexp.MustCompile(`window\.wx_errcode\s*=\s*(\d+)`)
var redirectPattern = regexp.MustCompile(`(?:window\.|document\.|top\.)?location(?:\.href)?\s*=\s*['"]([^'"]+)['"]|location\.replace\(\s*['"]([^'"]+)['"]`)

func match(re *regexp.Regexp, value string) string {
	m := re.FindStringSubmatch(value)
	if len(m) > 1 {
		return m[1]
	}
	return ""
}
func loginURL() string {
	return casOrigin + "/authserver/login?" + url.Values{"service": {eamsOrigin + "/eams/login.action"}}.Encode()
}
func (c *connection) login() error {
	data, _, err := c.request("GET", loginURL(), nil, nil)
	if err != nil {
		return err
	}
	iframe := ""
	for _, n := range tag(document(data), "iframe") {
		if iframe == "" || attr(n, "id") == "iframe" {
			iframe = attr(n, "src")
		}
		if attr(n, "id") == "iframe" {
			break
		}
	}
	entry, err := resolve(casOrigin+"/", iframe)
	if err != nil || iframe == "" || !strings.HasPrefix(entry, casOrigin+"/authserver/") {
		return codeError("QR_ENTRY_CHANGED")
	}
	data, final, err := c.request("GET", entry, nil, nil)
	if err != nil {
		return err
	}
	wx := final
	if !strings.HasPrefix(wx, wxOrigin+"/") {
		wx = wxLink.FindString(html.UnescapeString(strings.ReplaceAll(string(data), `\/`, "/")))
	}
	if wx == "" {
		return codeError("QR_ENTRY_CHANGED")
	}
	data, final, err = c.request("GET", wx, nil, map[string]string{"Referer": entry})
	if err != nil {
		return err
	}
	uuid := match(uuidPattern, string(data))
	if uuid == "" {
		return codeError("QR_PAGE_CHANGED")
	}
	callback := html.UnescapeString(match(callbackPattern, string(data)))
	if callback == "" {
		callback = casOrigin + "/authserver/callback"
	}
	if decoded, e := url.QueryUnescape(callback); e == nil && strings.HasPrefix(decoded, "https://") {
		callback = decoded
	}
	cb, err := url.Parse(callback)
	if err != nil || cb.Scheme != "https" || cb.Hostname() != "ids.chd.edu.cn" || !allowedURL(callback) || !strings.HasPrefix(cb.Path, "/authserver/") {
		return codeError("QR_CALLBACK_INVALID")
	}
	params := cb.Query()
	if !params.Has("state") {
		params.Set("state", match(statePattern, string(data)))
	}
	image := ""
	for _, n := range tag(document(data), "img") {
		if hasClass(n, "web_qrcode_img") {
			image = attr(n, "src")
			break
		}
	}
	imageURL, err := resolve(final, image)
	if err != nil || image == "" || !strings.HasPrefix(imageURL, wxOrigin+"/") {
		return codeError("QR_IMAGE_INVALID")
	}
	imageData, _, err := c.request("GET", imageURL, nil, nil)
	if err != nil {
		return err
	}
	if len(imageData) == 0 || len(imageData) > 1<<20 {
		return codeError("QR_IMAGE_INVALID")
	}
	c.emit(course.SourceEvent{Type: "qr", QR: "data:image/png;base64," + base64.StdEncoding.EncodeToString(imageData)})
	muid := ""
	wxParsed, _ := url.Parse(wxOrigin)
	for _, cookie := range c.client.Jar.Cookies(wxParsed) {
		if cookie.Name == "yyb_muid" {
			muid = cookie.Value
		}
	}
	if muid == "" {
		random := make([]byte, 16)
		if _, err = rand.Read(random); err != nil {
			return err
		}
		muid = hex.EncodeToString(random)
	}
	last := ""
	scanned := false
	deadline := time.Now().Add(3 * time.Minute)
	for time.Now().Before(deadline) {
		if err = c.pause(2 * time.Second); err != nil {
			return err
		}
		query := url.Values{"uuid": {uuid}}
		if last != "" && last != "408" {
			query.Set("last", last)
		}
		// Poll cancellation also interrupts a long-poll HTTP request.
		data, _, err = c.request("GET", "https://lp.open.weixin.qq.com/connect/l/qrconnect?"+query.Encode(), nil, map[string]string{"Referer": wxOrigin + "/", "Cookie": "yyb_muid=" + muid})
		if err != nil {
			if c.ctx.Err() != nil {
				return c.ctx.Err()
			}
			continue
		}
		code := match(codePattern, string(data))
		status := match(statusPattern, string(data))
		if status != "" {
			last = status
		}
		if code != "" {
			c.emit(course.SourceEvent{Type: "authenticating"})
			params.Set("code", code)
			cb.RawQuery = params.Encode()
			return c.completeLogin(cb.String())
		}
		if last == "404" && !scanned {
			c.emit(course.SourceEvent{Type: "scanned"})
			scanned = true
		}
		if last == "402" || last == "403" || last == "500" {
			break
		}
	}
	return codeError("QR_EXPIRED")
}

// Follow only declarative redirects and auto-submitted forms, never execute JS.
func (c *connection) navigate(raw string) ([]byte, string, error) {
	method := "GET"
	var form url.Values
	for hop := 0; hop < 8; hop++ {
		data, final, err := c.request(method, raw, form, nil)
		if err != nil {
			return nil, "", err
		}
		// WeChat login can finish at the CAS account portal. Its scripts are not
		// part of the service login; completeLogin will request the EAMS service
		// with the CAS cookie established by this redirect chain.
		u, _ := url.Parse(final)
		if u.Hostname() == "ids.chd.edu.cn" && u.Path == casLandingPath {
			return data, final, nil
		}
		page := string(data)
		if strings.Contains(page, "首次使用微信联合登录") {
			return nil, "", codeError("WECHAT_UNBOUND")
		}
		doc := document(data)
		next := ""
		method = "GET"
		form = nil
		if strings.Contains(page, ".submit(") {
			for _, n := range tag(doc, "form") {
				if attr(n, "action") == "" {
					continue
				}
				next = attr(n, "action")
				if strings.EqualFold(attr(n, "method"), "post") {
					method = "POST"
				}
				form = url.Values{}
				for _, input := range tag(n, "input") {
					if name := attr(input, "name"); name != "" {
						form.Add(name, attr(input, "value"))
					}
				}
				break
			}
		}
		if next == "" {
			for _, m := range tag(doc, "meta") {
				if strings.EqualFold(attr(m, "http-equiv"), "refresh") {
					v := attr(m, "content")
					if i := strings.Index(strings.ToLower(v), "url="); i >= 0 {
						next = strings.Trim(v[i+4:], " '\"")
						break
					}
				}
			}
		}
		if next == "" {
			for _, script := range tag(doc, "script") {
				m := redirectPattern.FindStringSubmatch(rawText(script))
				if len(m) > 1 {
					next = m[1]
					if next == "" {
						next = m[2]
					}
					break
				}
			}
		}
		if next == "" {
			return data, final, nil
		}
		raw, err = resolve(final, html.UnescapeString(next))
		if err != nil {
			return nil, "", err
		}
		if method == "GET" && form != nil {
			u, _ := url.Parse(raw)
			q := u.Query()
			for k, v := range form {
				q[k] = v
			}
			u.RawQuery = q.Encode()
			raw = u.String()
			form = nil
		}
	}
	return nil, "", codeError("QR_LOGIN_FAILED")
}
func (c *connection) completeLogin(callback string) error {
	_, _, err := c.navigate(callback)
	if err != nil {
		return err
	}
	_, final, err := c.navigate(loginURL())
	if err != nil {
		return err
	}
	u, _ := url.Parse(final)
	if u.Hostname() != "bkjw.chd.edu.cn" {
		return codeError("AUTH_REQUIRED")
	}
	// Verify an authenticated business page, not just a CAS ticket redirect.
	return c.retry(func() error { _, _, e := c.eams("GET", "/eams/home.action", nil); return e }, func(int) {})
}
