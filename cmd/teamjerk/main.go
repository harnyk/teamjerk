package main

import (
	"fmt"
	"log"

	"github.com/docopt/docopt-go"
	"github.com/harnyk/teamjerk/internal/app"
	"github.com/harnyk/teamjerk/internal/twapi"
)

//this will be replaced in the goreleaser build
var version = "development"

type command string

const (
	login    command = "login"
	logout   command = "logout"
	whoami   command = "whoami"
	projects command = "projects"
)

func main() {

	usage := `teamjerk
	
Usage:
  teamjerk login
  teamjerk logout
  teamjerk whoami
  teamjerk projects
`

	arguments, err := docopt.ParseArgs(usage, nil, version)
	if err != nil {
		log.Fatal(err)
	}

	cmd, err := getCommand(arguments)
	if err != nil {
		log.Fatal(err)
	}

	tw := twapi.NewClient("skeliasarl", "eu")
	app := app.NewApp(tw)

	switch cmd {
	case login:
		err = app.LogIn()
	case logout:
		err = app.LogOut()
	case whoami:
		err = app.WhoAmI()
	case projects:
		err = app.Projects()
	}

	if err != nil {
		log.Fatal(err)
	}

}

func getCommand(arguments docopt.Opts) (command, error) {
	for _, c := range []command{login, logout, whoami, projects} {
		cmdSelected, err := arguments.Bool(string(c))
		if err != nil {
			return "", err
		}
		if cmdSelected {
			return c, nil
		}
	}
	return "", fmt.Errorf("no command found")
}
