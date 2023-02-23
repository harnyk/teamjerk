package app

import (
	"fmt"
	"time"

	"github.com/harnyk/teamjerk/internal/authstore"
	"github.com/harnyk/teamjerk/internal/twapi"
)

type App interface {
	LogIn() error
	WhoAmI() error
	LogOut() error
	Projects() error
	Tasks() error
	Log() error
	Report(beginningOfMonth time.Time) error
}

type app struct {
	tw    twapi.Client
	store authstore.AuthStore[twapi.AuthData]
}

func NewApp(tw twapi.Client, store authstore.AuthStore[twapi.AuthData]) App {
	return &app{tw: tw, store: store}
}

func (a *app) Log() error {
	if !a.store.Exists() {
		return fmt.Errorf("not logged in")
	}

	auth, err := a.store.Load()
	if err != nil {
		return err
	}

	tasks, err := a.tw.GetTasks(auth)
	if err != nil {
		return err
	}

	projects, err := a.tw.GetProjects(auth)
	if err != nil {
		return err
	}

	taskGroups, err := getProjectsAndTasks(projects, tasks)
	if err != nil {
		return err
	}

	timelogTarget, err := selectTimelogTarget(taskGroups)
	if err != nil {
		return err
	}

	fmt.Println("Selected:", timelogTarget.PrettyPrint())

	duration := askDuration()

	fmt.Println("Duration:", duration.Hours())

	return nil
}

func (a *app) LogIn() error {
	email, err := askEmail()
	if err != nil {
		return err
	}

	password, err := askPassword()
	if err != nil {
		return err
	}

	accounts, err := a.tw.GetAccountsToLogIn(email, password)
	if err != nil {
		return err
	}

	account, err := selectAccount(*accounts)

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

	fmt.Println("ID         :", res.Person.ID)
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

func (a *app) Report(beginningOfMonth time.Time) error {
	if !a.store.Exists() {
		return fmt.Errorf("not logged in")
	}

	auth, err := a.store.Load()
	if err != nil {
		return err
	}

	res, err := a.tw.GetLoggedTime(auth, beginningOfMonth)
	if err != nil {
		return err
	}

	fmt.Println("Logged time for", beginningOfMonth.Format("2006-01"))

	//just output the result for debugging (with property names)
	fmt.Printf("%+v\n", res)

	return nil
}
