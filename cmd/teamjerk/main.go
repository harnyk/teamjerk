package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/docopt/docopt-go"
	"github.com/harnyk/teamjerk/internal/app"
	"github.com/harnyk/teamjerk/internal/authstore"
	"github.com/harnyk/teamjerk/internal/twapi"
)

//this will be replaced in the goreleaser build
var version = "development"

type command string

const (
	cmdLogin    command = "login"
	cmdLogout   command = "logout"
	cmdWhoami   command = "whoami"
	cmdProjects command = "projects"
	cmdTasks    command = "tasks"
	cmdLog      command = "log"
)

func getAuthFilePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(home, ".teamjerk", "auth.json"), nil
}

func main() {

	usage := `teamjerk
	
Usage:
  teamjerk login
  teamjerk logout
  teamjerk whoami
  teamjerk projects
  teamjerk tasks
  teamjerk log
`

	arguments, err := docopt.ParseArgs(usage, nil, version)
	if err != nil {
		log.Fatal(err)
	}

	cmd, err := getCommand(arguments)
	if err != nil {
		log.Fatal(err)
	}

	authFilePath, err := getAuthFilePath()
	if err != nil {
		log.Fatal(err)
	}

	tw := twapi.NewClient()
	store := authstore.NewAuthStore[twapi.AuthData](authFilePath)
	app := app.NewApp(tw, store)

	switch cmd {
	case cmdLogin:
		err = app.LogIn()
	case cmdLogout:
		err = app.LogOut()
	case cmdWhoami:
		err = app.WhoAmI()
	case cmdProjects:
		err = app.Projects()
	case cmdTasks:
		err = app.Tasks()
	case cmdLog:
		err = app.Log()
	}

	if err != nil {
		log.Fatal(err)
	}

}

func getCommand(arguments docopt.Opts) (command, error) {
	for _, c := range []command{cmdLogin, cmdLogout, cmdWhoami, cmdProjects, cmdTasks, cmdLog} {
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
