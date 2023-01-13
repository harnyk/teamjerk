package app

import (
	"fmt"
	"os"

	"github.com/harnyk/teamjerk/internal/authstore"
	"github.com/harnyk/teamjerk/internal/twapi"
	"github.com/howeyc/gopass"
)

type App interface {
	LogIn() error
	WhoAmI() error
	LogOut() error
	Projects() error
}

type app struct {
	tw    twapi.Client
	store authstore.AuthStore[twapi.AuthData]
}

func NewApp(tw twapi.Client, store authstore.AuthStore[twapi.AuthData]) App {
	return &app{tw: tw, store: store}
}

func (a *app) LogIn() error {
	//read email from stdin
	// email, err := gopass.GetPasswdPrompt("Email: ",
	// 	false, os.Stdin, os.Stdout)
	// This is wrong. Do not use gopass for email. Use fmt.Scanln instead.

	var email string
	fmt.Print("Email: ")
	_, err := fmt.Scanln(&email)
	if err != nil {
		return err
	}

	//read password from stdin hiding the input
	password, err := gopass.GetPasswdPrompt("Password: ",
		false, os.Stdin, os.Stdout)
	if err != nil {
		return err
	}

	passwordStr := string(password)

	accounts, err := a.tw.GetAccountsToLogIn(email, passwordStr)

	fmt.Println("Select account:")
	for i, account := range accounts.Accounts {
		fmt.Printf("%d) %s %s @ %s\n", i, account.User.FirstName, account.User.LastName, account.Installation.Company.Name)
	}

	var accountIndex int

	for {
		fmt.Print("Account: ")
		_, err = fmt.Scanln(&accountIndex)
		if err != nil {
			return err
		}

		if accountIndex < 0 || accountIndex >= len(accounts.Accounts) {
			fmt.Println("Invalid account index")
			continue
		}

		break
	}

	account := accounts.Accounts[accountIndex]

	auth, err := a.tw.LogIn(account.Installation.ApiEndPoint, email, passwordStr)

	if err != nil {
		return err
	}

	err = a.store.Save(auth)
	if err != nil {
		return err
	}

	return nil
}

func (a *app) WhoAmI() error {
	exists := a.store.Exists()
	if !exists {
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
	panic("implement me")
}
