package twapi

import (
	"fmt"
	"net/http"

	"github.com/go-resty/resty/v2"
)

type AuthData struct {
	Token string
}

type User struct {
	ID        string `json:"id"`
	FirstName string `json:"first-name"`
	LastName  string `json:"last-name"`
}

type Client interface {
	LogIn(email, password string) (*AuthData, error)
	GetMe(authData *AuthData) (*User, error)
}

type client struct {
	companySlug string
	zone        string
}

func NewClient(companySlug, zone string) Client {
	return &client{companySlug: companySlug, zone: zone}
}

func (c *client) baseURL() string {
	return "https://" + c.companySlug + "." + c.zone + ".teamwork.com"
}

func (c *client) LogIn(email, password string) (*AuthData, error) {
	// Plan:
	// 1. POST to https://{{.CompanySlug}}.{{.Zone}}.teamwork.com/launchpad/v1/login.json
	//  with body: {"email": email, "password": password, "rememberMe": false}
	// 2. If response status is 200, return AuthData with token from response cookie 'tw-auth'
	// 3. If response status is not 200, return error

	client := resty.New()

	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(map[string]interface{}{
			"email":      email,
			"password":   password,
			"rememberMe": false,
		}).
		Post(c.baseURL() + "/launchpad/v1/login.json")

	if err != nil {
		return nil, err
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("status code: %d", resp.StatusCode())
	}

	// return &AuthData{Token: resp.Cookies()[0].Value}, nil
	// This is wrong, because we can have multiple cookies. Let's use the cookie package to parse the cookies:
	cookies := resp.Cookies()
	for _, cookie := range cookies {
		if cookie.Name == "tw-auth" {
			return &AuthData{Token: cookie.Value}, nil
		}
	}

	return nil, fmt.Errorf("cookie 'tw-auth' not found")
}

func (c *client) GetMe(authData *AuthData) (*User, error) {
	// Plan:
	// 1. GET to https://{{.CompanySlug}}.{{.Zone}}.teamwork.com/launchpad/v1/me.json
	//  with cookie 'tw-auth' set to authData.Token
	// 2. If response status is 200, return User from response body
	// 3. If response status is not 200, return error

	client := resty.New()

	var user *User

	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetCookie(&http.Cookie{
			Name:  "tw-auth",
			Value: authData.Token,
		}).
		SetResult(user).
		Get(c.baseURL() + "/launchpad/v1/me.json")

	if err != nil {
		return nil, err
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("status code: %d", resp.StatusCode())
	}

	return user, nil
}
