package twapi

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

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
	GetTasks(authData *AuthData) (*TasksResponse, error)
	LogTime(authData *AuthData, timeLog *LogtimeRequestWithProjectID) error
	GetLoggedTime(authData *AuthData, beginningOfMonth time.Time) (*TimeChartResponse, error)
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
			"rememberMe": true,
		}).
		Post(apiEndPoint + "launchpad/v1/login.json")

	if err != nil {
		return nil, err
	}

	if resp.StatusCode() != http.StatusOK {
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
	user := &ProfileResponse{}

	resp, err := c.getAuthenticatedRequest(authData).
		SetResult(user).
		Get(authData.APIEndPoint + "me.json")

	if err != nil {
		return nil, err
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("status code: %d", resp.StatusCode())
	}

	return user, nil
}

func (c *client) GetProjects(authData *AuthData) (*ProjectsResponse, error) {
	projects := &ProjectsResponse{}

	resp, err := c.getAuthenticatedRequest(authData).
		SetResult(projects).
		Get(authData.APIEndPoint + "projects.json")

	if err != nil {
		return nil, err
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("status code: %d", resp.StatusCode())
	}

	return projects, nil
}

func (c *client) GetTasks(authData *AuthData) (*TasksResponse, error) {
	tasks := &TasksResponse{}

	resp, err := c.getAuthenticatedRequest(authData).
		SetResult(tasks).
		Get(authData.APIEndPoint + "tasks.json")

	if err != nil {
		return nil, err
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("status code: %d", resp.StatusCode())
	}

	return tasks, nil
}

func (c *client) LogTime(authData *AuthData, timeLog *LogtimeRequestWithProjectID) error {
	var url string
	if timeLog.Timelog.TaskID != 0 {
		url = fmt.Sprintf("%sprojects/api/v3/tasks/%d/time.json", authData.APIEndPoint, timeLog.Timelog.TaskID)
	} else {
		url = fmt.Sprintf("%sprojects/api/v3/projects/%d/time.json", authData.APIEndPoint, timeLog.ProjectID)
	}

	resp, err := c.getAuthenticatedRequest(authData).
		SetBody(timeLog.LogtimeRequest).
		Post(url)

	if err != nil {
		return err
	}

	if resp.StatusCode() != http.StatusCreated {
		return fmt.Errorf("status code: %d", resp.StatusCode())
	}

	return nil
}

func (c *client) getAuthenticatedRequest(authData *AuthData) *resty.Request {
	client := resty.New()

	return client.R().
		SetHeader("Content-Type", "application/json").
		SetCookie(&http.Cookie{
			Name:  "tw-auth",
			Value: authData.Token,
		})
}

func (c *client) GetLoggedTime(authData *AuthData, beginningOfMonth time.Time) (*TimeChartResponse, error) {
	timeChart := &TimeChartResponse{}

	month := beginningOfMonth.Month()
	year := beginningOfMonth.Year()
	projectID := 0
	page := 1
	pageSize := 50

	user, err := c.GetMe(authData)
	if err != nil {
		return nil, err
	}

	userID := user.Person.ID

	resp, err := c.getAuthenticatedRequest(authData).
		SetResult(timeChart).
		SetQueryParam("m", strconv.Itoa(int(month))).
		SetQueryParam("y", strconv.Itoa(year)).
		SetQueryParam("projectId", strconv.Itoa(projectID)).
		SetQueryParam("page", strconv.Itoa(page)).
		SetQueryParam("pageSize", strconv.Itoa(pageSize)).
		Get(authData.APIEndPoint + "people/" + userID + "/loggedtime.json")

	if err != nil {
		return nil, err
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("status code: %d", resp.StatusCode())
	}

	return timeChart, nil
}
