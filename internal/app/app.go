package app

import (
	"fmt"
	"os"

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
	tw twapi.Client
}

func NewApp(tw twapi.Client) App {
	return &app{tw: tw}
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

	auth, err := a.tw.LogIn(email, string(password))
	if err != nil {
		return err
	}

	//TODO: save auth to file
	fmt.Printf("auth: %+v", auth)

	return nil
}

func (a *app) WhoAmI() error {
	panic("implement me")
}

func (a *app) LogOut() error {
	panic("implement me")
}

func (a *app) Projects() error {
	panic("implement me")
}
