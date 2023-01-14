package twapi

import "fmt"

/*

Example of AccountsResponse:

{
    "accounts": [
        {
            "installation": {
                "id": 328092,
                "name": "Skelia sarl",
                "url": "https://skeliasarl.eu.teamwork.com/",
                "region": "EU",
                "logo": "",
                "loginStartText": "",
                "apiEndPoint": "https://skeliasarl.eu.teamwork.com/",
                "company": {
                    "id": 97396,
                    "name": "Skelia sarl",
                    "logo": "https://s3-eu-west-1.amazonaws.com/tw-eu-files/328092/companyLogo/tf_A30F508C-922A-D288-C424FAFA8A332D5B.elogo.png"
                },
                "projectsEnabled": true,
                "deskEnabled": true,
                "chatEnabled": false
            },
            "user": {
                "id": 459261,
                "firstName": "Mark",
                "lastName": "Harnyk",
                "email": "markh@sweepbright.com",
                "avatar": "https://tw-eu-files.s3.eu-west-1.amazonaws.com/328092/userAvatar/twia_c3f9bbe53bcd85236f872c71a42e8651.png",
                "company": {
                    "id": 170467,
                    "name": "Skelia ext",
                    "logo": "https://s3-eu-west-1.amazonaws.com/tw-eu-files/"
                }
            }
        }
    ],
    "ignoredAccounts": 0,
    "status": "ok"
}
*/

type AccountsResponse struct {
	Accounts        []Account `json:"accounts"`
	IgnoredAccounts int       `json:"ignoredAccounts"`
	Status          string    `json:"status"`
}

type Account struct {
	Installation Installation `json:"installation"`
	User         User         `json:"user"`
}

type Installation struct {
	ID                int               `json:"id"`
	Name              string            `json:"name"`
	URL               string            `json:"url"`
	Region            string            `json:"region"`
	Logo              string            `json:"logo"`
	LoginStartText    string            `json:"loginStartText"`
	ApiEndPoint       string            `json:"apiEndPoint"`
	Company           Company           `json:"company"`
	ProjectsEnabled   bool              `json:"projectsEnabled"`
	DeskEnabled       bool              `json:"deskEnabled"`
	ChatEnabled       bool              `json:"chatEnabled"`
	ProjectTemplates  []ProjectTemplate `json:"projectTemplates"`
	ProjectCategories []ProjectCategory `json:"projectCategories"`
}

type User struct {
	ID        int    `json:"id"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Email     string `json:"email"`
	Avatar    string `json:"avatar"`
	Company   Company
}

type Company struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Logo string `json:"logo"`
}

type ProjectTemplate struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func (a *Account) String() string {
	return fmt.Sprintf("%s %s @ %s", a.User.FirstName, a.User.LastName, a.Installation.Company.Name)
}
