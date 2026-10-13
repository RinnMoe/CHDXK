package bkjw

import (
	"net/url"
	"time"
)

// Only QR preparation GETs use this policy. CAS callbacks and form submissions
// can consume one-time tickets and are intentionally not replayed here.
func (c *connection) loginRead(raw string, form url.Values, headers map[string]string) ([]byte, string, error) {
	var data []byte
	var final string
	var err error
	for attempt := 0; attempt < 3; attempt++ {
		if c.ctx.Err() != nil {
			return nil, "", c.ctx.Err()
		}
		if attempt > 0 {
			delay := time.Duration(attempt) * time.Second
			if c.retryDelay != nil {
				delay = c.retryDelay(attempt)
			}
			if err := c.pause(max(delay, time.Until(c.retryAfter))); err != nil {
				return nil, "", err
			}
		}
		data, final, err = c.request("GET", raw, form, headers)
		if err == nil && !isMaintenancePage(data) {
			return data, final, nil
		}
		if err == nil {
			err = codeError("UPSTREAM_UNAVAILABLE")
		}
		switch errorCode(err, "") {
		case "AUTH_REQUIRED", "PERMISSION_DENIED", "URL_NOT_ALLOWED", "UPSTREAM_REJECTED":
			return nil, "", err
		}
	}
	return nil, "", err
}
