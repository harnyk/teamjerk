package app

import (
	"fmt"
	"os"
	"strconv"

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
	Log() error
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

	taskGroups, err := a.getProjectsAndTasks(projects, tasks)
	if err != nil {
		return err
	}

	timelogTarget, err := a.selectTimelogTarget(taskGroups)
	if err != nil {
		return err
	}

	fmt.Println("Selected:", timelogTarget)

	return nil
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

func (a *app) getProjectsAndTasks(
	projects *twapi.ProjectsResponse,
	tasks *twapi.TasksResponse,
) ([]twapi.TasksGroup, error) {
	taskGroups := tasks.GroupByProject()

	projectsIdsFromTaskGroups, _ := slices.Map(taskGroups,
		func(i int, taskGroup twapi.TasksGroup) (string, error) {
			idStr := strconv.Itoa(int(taskGroup.Project.ID))
			return idStr, nil
		},
	)

	projectsWithoutTasks, _ := slices.Filter(projects.Projects,
		func(project twapi.Project) (bool, error) {
			for _, id := range projectsIdsFromTaskGroups {
				if project.ID == id {
					return false, nil
				}
			}
			return true, nil
		},
	)

	taskGroupsWithoutTasks, err := slices.Map(projectsWithoutTasks,
		func(i int, project twapi.Project) (twapi.TasksGroup, error) {
			id, err := strconv.Atoi(project.ID)
			if err != nil {
				return twapi.TasksGroup{}, err
			}

			return twapi.TasksGroup{
				Project: twapi.TaskProject{
					ID:   uint64(id),
					Name: project.Name,
				},
				Tasks: []twapi.Task{},
			}, nil
		},
	)

	if err != nil {
		return nil, err
	}

	taskGroups = append(taskGroups, taskGroupsWithoutTasks...)

	return taskGroups, nil
}

func (a *app) selectTimelogTarget(taskGroups []twapi.TasksGroup) (*timelogTargetSelection, error) {
	timelogTargetSelections := []timelogTargetSelection{}

	for _, taskGroup := range taskGroups {
		timelogTargetSelections = append(timelogTargetSelections, timelogTargetSelection{
			Project: taskGroup.Project,
		})

		for _, task := range taskGroup.Tasks {
			timelogTargetSelections = append(timelogTargetSelections, timelogTargetSelection{
				Project: taskGroup.Project,
				Task:    task,
			})
		}
	}

	timelogTargetLabels, _ := slices.Map(timelogTargetSelections,
		func(i int, selection timelogTargetSelection) (string, error) {
			if selection.Task.ID == 0 {
				return selection.Project.Name, nil
			}

			return fmt.Sprintf("%s / %s", selection.Project.Name, selection.Task.Content), nil
		},
	)

	prompt := promptui.Select{
		Label: "Select timelog target",
		Items: timelogTargetLabels,
	}

	timelogTargetIndex, _, err := prompt.Run()
	if err != nil {
		return nil, err
	}

	return &timelogTargetSelections[timelogTargetIndex], nil
}

type timelogTargetSelection struct {
	Project twapi.TaskProject
	Task    twapi.Task
}
