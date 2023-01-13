package twapi

import (
	"fmt"
	"net/http"

	"github.com/go-resty/resty/v2"
)

type AuthData struct {
	APIEndPoint string
	Token       string
}

type Client interface {
	GetAccountsToLogIn(email, password string) (*AccountsResponse, error)
	LogIn(apiEndPoint, email, password string) (*AuthData, error)
	GetMe(authData *AuthData) (*ProfileResponse, error)
	GetProjects(authData *AuthData) (*ProjectsResponse, error)
}

type client struct {
}

func NewClient() Client {
	return &client{}
}

func (c *client) baseURL() string {
	return "https://www.teamwork.com"
}

func (c *client) GetAccountsToLogIn(email, password string) (*AccountsResponse, error) {
	client := resty.New()

	accountsResponse := &AccountsResponse{}

	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(map[string]interface{}{
			"email":      email,
			"password":   password,
			"rememberMe": true,
		}).
		SetResult(&accountsResponse).
		Post(c.baseURL() + "/launchpad/v1/accounts.json?generic=true")

	if err != nil {
		return nil, err
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("status code: %d", resp.StatusCode())
	}

	return accountsResponse, nil
}

func (c *client) LogIn(apiEndPoint, email, password string) (*AuthData, error) {
	// Plan:
	// 1. POST to https://{{apiEndPoint}}launchpad/v1/login.json
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
		Post(apiEndPoint + "launchpad/v1/login.json")

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
			return &AuthData{
				APIEndPoint: apiEndPoint,
				Token:       cookie.Value,
			}, nil
		}
	}

	return nil, fmt.Errorf("cookie 'tw-auth' not found")
}

func (c *client) GetMe(authData *AuthData) (*ProfileResponse, error) {
	client := resty.New()

	user := &ProfileResponse{}

	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetCookie(&http.Cookie{
			Name:  "tw-auth",
			Value: authData.Token,
		}).
		SetResult(user).
		Get(authData.APIEndPoint + "me.json")

	if err != nil {
		return nil, err
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("status code: %d", resp.StatusCode())
	}

	return user, nil
}

func (c *client) GetProjects(authData *AuthData) (*ProjectsResponse, error) {
	client := resty.New()

	projects := &ProjectsResponse{}

	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetCookie(&http.Cookie{
			Name:  "tw-auth",
			Value: authData.Token,
		}).
		SetResult(projects).
		Get(authData.APIEndPoint + "projects.json")

	if err != nil {
		return nil, err
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("status code: %d", resp.StatusCode())
	}

	return projects, nil
}
