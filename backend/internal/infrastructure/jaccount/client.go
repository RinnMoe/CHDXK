package jaccount

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/go-resty/resty/v2"
	"golang.org/x/oauth2"

	"jcourse/config"
	domainjaccount "jcourse/internal/domain/jaccount"
)

type OAuthClient struct {
	oauthConfig *oauth2.Config
	resty       *resty.Client
	apiBaseURL  string
	enabled     bool
}

func NewOAuthClient(conf config.JAccountConfig) *OAuthClient {
	return &OAuthClient{
		oauthConfig: &oauth2.Config{
			ClientID:     conf.ClientID,
			ClientSecret: conf.ClientSecret,
			RedirectURL:  conf.RedirectURL,
			Scopes:       conf.Scopes,
			Endpoint: oauth2.Endpoint{
				AuthURL:  conf.AuthorizeURL,
				TokenURL: conf.TokenURL,
			},
		},
		resty:      resty.New(),
		apiBaseURL: strings.TrimRight(conf.APIBaseURL, "/") + "/",
		enabled:    conf.CourseSyncEnabled,
	}
}

func (c *OAuthClient) Enabled() bool {
	return c != nil && c.enabled && c.oauthConfig.ClientID != "" && c.oauthConfig.ClientSecret != "" && c.oauthConfig.RedirectURL != "" && c.apiBaseURL != "/"
}

func (c *OAuthClient) AuthCodeURL(state string) (string, error) {
	if !c.Enabled() {
		return "", domainjaccount.ErrDisabled
	}
	return c.oauthConfig.AuthCodeURL(state), nil
}

func (c *OAuthClient) Exchange(ctx context.Context, code string) (*oauth2.Token, error) {
	if !c.Enabled() {
		return nil, domainjaccount.ErrDisabled
	}
	return c.oauthConfig.Exchange(ctx, code)
}

func (c *OAuthClient) Lessons(ctx context.Context, token *oauth2.Token, semester string) ([]domainjaccount.LessonCourse, error) {
	if !c.Enabled() {
		return nil, domainjaccount.ErrDisabled
	}
	semester = strings.TrimSpace(semester)
	if semester == "" {
		return nil, domainjaccount.ErrSemesterRequired
	}
	endpoint, err := url.JoinPath(c.apiBaseURL, "v1", "me", "lessons", semester, "/")
	if err != nil {
		return nil, err
	}

	var body lessonsResponse
	res, err := c.resty.R().
		SetContext(ctx).
		SetAuthToken(token.AccessToken).
		SetQueryParam("classes", "false").
		SetResult(&body).
		Get(endpoint)
	if err != nil {
		return nil, err
	}
	if res.StatusCode() < http.StatusOK || res.StatusCode() >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("jaccount lessons request failed: %s", res.Status())
	}
	if body.Errno != 0 {
		msg := strings.TrimSpace(body.Error)
		if msg == "" {
			msg = strings.TrimSpace(body.Message)
		}
		if msg == "" {
			msg = fmt.Sprintf("errno %d", body.Errno)
		}
		return nil, fmt.Errorf("jaccount lessons error: %s", msg)
	}

	lessons := make([]domainjaccount.LessonCourse, 0, len(body.Entities))
	for _, entity := range body.Entities {
		if entity.Course.Code == "" || len(entity.Teachers) == 0 {
			continue
		}
		lessons = append(lessons, domainjaccount.LessonCourse{
			Code:        entity.Course.Code,
			TeacherName: entity.Teachers[0].Name,
		})
	}
	return lessons, nil
}

type lessonsResponse struct {
	Errno    int            `json:"errno"`
	Error    string         `json:"error"`
	Message  string         `json:"message"`
	Entities []lessonEntity `json:"entities"`
}

type lessonEntity struct {
	Course struct {
		Code string `json:"code"`
	} `json:"course"`
	Teachers []struct {
		Name string `json:"name"`
	} `json:"teachers"`
}

var _ domainjaccount.Client = (*OAuthClient)(nil)
