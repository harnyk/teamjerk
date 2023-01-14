package app

import (
	"fmt"
	"os"

	"github.com/bobg/go-generics/slices"
	"github.com/harnyk/teamjerk/internal/authstore"
	"github.com/harnyk/teamjerk/internal/twapi"
	"github.com/howeyc/gopass"
	"github.com/manifoldco/promptui"
)

type App interface {
	LogIn() error
	WhoAmI() error
	LogOut() error
	Projects() error
	Tasks() error
}

type app struct {
	tw    twapi.Client
	store authstore.AuthStore[twapi.AuthData]
}

func NewApp(tw twapi.Client, store authstore.AuthStore[twapi.AuthData]) App {
	return &app{tw: tw, store: store}
}

func (a *app) LogIn() error {
	email, err := a.askEmail()
	if err != nil {
		return err
	}

	password, err := a.askPassword()
	if err != nil {
		return err
	}

	accounts, err := a.tw.GetAccountsToLogIn(email, password)
	if err != nil {
		return err
	}

	account, err := a.selectAccount(*accounts)

	auth, err := a.tw.LogIn(account.Installation.ApiEndPoint, email, password)
	if err != nil {
		return err
	}

	err = a.store.Save(auth)
	if err != nil {
		return err
	}

	fmt.Printf("Logged in successfully as %s\n", account.String())

	return nil
}

func (a *app) WhoAmI() error {
	if !a.store.Exists() {
		return fmt.Errorf("not logged in")
	}

	auth, err := a.store.Load()
	if err != nil {
		return err
	}

	res, err := a.tw.GetMe(auth)
	if err != nil {
		return err
	}

	fmt.Println("First Name :", res.Person.FirstName)
	fmt.Println("Last Name  :", res.Person.LastName)
	fmt.Println("Email      :", res.Person.EmailAddress)
	fmt.Println("Company    :", res.Person.CompanyName)

	return nil
}

func (a *app) LogOut() error {
	panic("implement me")
}

func (a *app) Projects() error {
	if !a.store.Exists() {
		return fmt.Errorf("not logged in")
	}

	auth, err := a.store.Load()
	if err != nil {
		return err
	}

	res, err := a.tw.GetProjects(auth)
	if err != nil {
		return err
	}

	for _, project := range res.Projects {
		fmt.Printf("[ID: %s] %s\n", project.ID, project.Name)
	}

	return nil
}

func (a *app) Tasks() error {
	if !a.store.Exists() {
		return fmt.Errorf("not logged in")
	}

	auth, err := a.store.Load()
	if err != nil {
		return err
	}

	res, err := a.tw.GetTasks(auth)
	if err != nil {
		return err
	}

	taskGroups := res.GroupByProject()

	for _, taskGroup := range taskGroups {
		fmt.Printf("[ProjectID: %d] %s\n", taskGroup.Project.ID, taskGroup.Project.Name)
		for _, task := range taskGroup.Tasks {
			fmt.Printf("  [ID: %d] %s\n", task.ID, task.Content)
		}
	}

	return nil
}

func (a *app) selectAccount(accounts twapi.AccountsResponse) (twapi.Account, error) {
	if len(accounts.Accounts) == 1 {
		return accounts.Accounts[0], nil
	}

	accountLabels, _ := slices.Map(accounts.Accounts,
		func(i int, account twapi.Account) (string, error) {
			return account.String(), nil
		})

	prompt := promptui.Select{
		Label: "Select account",
		Items: accountLabels,
	}

	accountIndex, _, err := prompt.Run()

	if err != nil {
		return twapi.Account{}, err
	}

	return accounts.Accounts[accountIndex], nil
}

func (a *app) askEmail() (string, error) {
	var email string
	fmt.Print("Email: ")
	_, err := fmt.Scanln(&email)
	if err != nil {
		return "", err
	}

	return email, nil
}

func (a *app) askPassword() (string, error) {
	password, err := gopass.GetPasswdPrompt("Password: ",
		false, os.Stdin, os.Stdout)
	if err != nil {
		return "", err
	}

	return string(password), nil
}
