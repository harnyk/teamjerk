package app

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/fatih/color"

	"github.com/bobg/go-generics/slices"
	"github.com/harnyk/teamjerk/internal/twapi"
	"github.com/howeyc/gopass"
	"github.com/manifoldco/promptui"
)

type timelogTargetSelection struct {
	Project twapi.TaskProject
	Task    twapi.Task
}

func (tts *timelogTargetSelection) Serialize() string {
	return fmt.Sprintf("%d:%d", tts.Project.ID, tts.Task.ID)
}

func (tts *timelogTargetSelection) PrettyPrint() string {
	return fmt.Sprintf("[%s] %s: %s", tts.Serialize(), tts.Project.Name, tts.Task.Content)
}

// getProjectsAndTasks returns a slice of twapi.TasksGroup
// with tasks grouped by project.
// If a project has no tasks, it will be added to the slice
// as a twapi.TasksGroup with an empty slice of tasks.
func getProjectsAndTasks(
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

// selectTimelogTarget returns a timelogTargetSelection
// based on the user's selection.
func selectTimelogTarget(taskGroups []twapi.TasksGroup) (*timelogTargetSelection, error) {
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

func selectAccount(accounts twapi.AccountsResponse) (twapi.Account, error) {
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

// askStartTime returns a time.Time in the format of HH:mm
// by default (if a user just hits Return) it returns 09:00
// Loops until a valid input is given
func askStartTime() time.Time {
	//TODO: make this configurable
	defaultStartTime := time.Date(0, 0, 0, 9, 0, 0, 0, time.UTC)

	for {
		var startTimeStr string
		fmt.Print("Start time (HH:mm): ")
		color.Yellow(" (default: %s)", defaultStartTime.Format("15:04"))
		_, err := fmt.Scanln(&startTimeStr)
		//unexpected newline is returned when user hits Return, so we ignore it
		if err != nil && err.Error() != "unexpected newline" {
			fmt.Println("Invalid input")
			continue
		}

		if startTimeStr == "" {
			return time.Date(0, 0, 0, 9, 0, 0, 0, time.UTC)
		}

		startTime, err := time.Parse("15:04", startTimeStr)
		if err != nil {
			fmt.Println("Invalid input")
			continue
		}

		return startTime
	}
}

// askDate returns a time.Time in the format of YYYY-MM-DD
// by default (if a user just hits Return) it returns today's date
// Loops until a valid input is given
func askDate() time.Time {
	defaultDate := time.Now()

	for {
		var dateStr string
		fmt.Print("Date (YYYY-MM-DD): ")
		color.Yellow(" (default: %s)", defaultDate.Format("2006-01-02"))
		_, err := fmt.Scanln(&dateStr)
		if err != nil && err.Error() != "unexpected newline" {
			fmt.Println("Invalid input")
			continue
		}

		if dateStr == "" {
			return defaultDate
		}

		date, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			fmt.Println("Invalid input")
			continue
		}

		return date
	}
}

//returns a time.Duration between 0 and 24 hours
//expects a string in the format of "8.5" for 8 hours and 30 minutes
//Loops until a valid input is given
func askDuration() time.Duration {
	for {
		var durationStr string
		fmt.Print("Duration (in hours): ")
		_, err := fmt.Scanln(&durationStr)
		if err != nil {
			fmt.Println("Invalid input")
			continue
		}

		duration, err := strconv.ParseFloat(durationStr, 64)
		if err != nil {
			fmt.Println("Invalid input")
			continue
		}

		if duration < 0 || duration > 24 {
			fmt.Println("Duration must be between 0 and 24 hours")
			continue
		}

		return time.Duration(duration * float64(time.Hour))
	}
}

func askEmail() (string, error) {
	var email string
	fmt.Print("Email: ")
	_, err := fmt.Scanln(&email)
	if err != nil {
		return "", err
	}

	return email, nil
}

func askPassword() (string, error) {
	password, err := gopass.GetPasswdPrompt("Password: ",
		false, os.Stdin, os.Stdout)
	if err != nil {
		return "", err
	}

	return string(password), nil
}
