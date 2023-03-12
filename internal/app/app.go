package app

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"time"

	"github.com/fatih/color"
	"github.com/harnyk/teamjerk/internal/authstore"
	"github.com/harnyk/teamjerk/internal/twapi"
	"github.com/olekukonko/tablewriter"
)

type App interface {
	LogIn() error
	WhoAmI() error
	LogOut() error
	Projects() error
	Tasks() error
	Log(options LogOptions) error
	Report(beginningOfMonth time.Time) error
}

type app struct {
	tw    twapi.Client
	store authstore.AuthStore[twapi.AuthData]
}

func NewApp(tw twapi.Client, store authstore.AuthStore[twapi.AuthData]) App {
	return &app{tw: tw, store: store}
}

func (a *app) Log(options LogOptions) error {
	if !a.store.Exists() {
		return fmt.Errorf("not logged in")
	}

	auth, err := a.store.Load()
	if err != nil {
		return err
	}

	user, err := a.tw.GetMe(auth)
	if err != nil {
		return err
	}

	userId, err := strconv.ParseUint(user.Person.ID, 10, 64)
	if err != nil {
		return err
	}

	var taskID uint64
	var projectID uint64
	var prettyPrint string

	if options.ProjectID == 0 && options.TaskID == 0 {
		projectID, taskID, prettyPrint, err = a.getProjectAndTaskInteractively(auth)
		if err != nil {
			return err
		}
	} else {
		projectID = options.ProjectID
		taskID = options.TaskID
	}
	fmt.Println("Target:", prettyPrint)

	var duration time.Duration
	if options.Duration == 0 {
		duration = askDuration()
	} else {
		duration = options.Duration
	}
	fmt.Println("Duration:", duration.Hours())

	var startTime time.Time
	if options.StartTime.IsZero() {
		startTime = askStartTime()
	} else {
		startTime = options.StartTime
	}
	fmt.Println("Start time:", startTime.Format("15:04:05"))

	var date time.Time
	if options.Date.IsZero() {
		date = askDate()
	} else {
		date = options.Date
	}
	fmt.Println("Date:", date.Format("2006-01-02"))

	description := options.Description

	if options.DryRun {
		fmt.Println("Dry run, not logging anything")
		return nil
	}

	request := &twapi.LogtimeRequestWithProjectID{
		LogtimeRequest: twapi.LogtimeRequest{
			Timelog: twapi.LogtimeTimelog{
				TaskID:      taskID,
				Hours:       uint64(duration.Hours()),
				Minutes:     uint64(duration.Minutes()) % 60,
				Date:        date.Format("2006-01-02"),
				Time:        startTime.Format("15:04:05"),
				Description: description, //TODO: would be nice to take this from the GitHub activity or at least from a command line argument
				IsBillable:  !options.NonBillable,
				UserID:      userId,
				TagIDs:      []uint64{},
			},
			TimelogOptions: twapi.LogtimeTimelogOptions{
				MarkTaskComplete: false,
			},
		},
		ProjectID: projectID,
	}

	err = a.tw.LogTime(auth, request)
	if err != nil {
		return err
	}

	return nil
}

func (a *app) getProjectAndTaskInteractively(auth *twapi.AuthData) (projectID, taskId uint64, prettyPrint string, err error) {
	projectID = 0
	taskId = 0

	projects, err := a.tw.GetProjects(auth)
	if err != nil {
		return
	}

	tasks, err := a.tw.GetTasks(auth)
	if err != nil {
		return
	}

	taskGroups, err := getProjectsAndTasks(projects, tasks)
	if err != nil {
		return
	}

	timelogTarget, err := selectTimelogTarget(taskGroups)
	if err != nil {
		return
	}

	projectID = timelogTarget.Project.ID
	taskId = timelogTarget.Task.ID
	prettyPrint = timelogTarget.PrettyPrint()

	return
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
	if err != nil {
		return err
	}

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

type chartSeries struct {
	Date            time.Time     `json:"date"`
	BillableTime    time.Duration `json:"billable"`
	NonBillableTime time.Duration `json:"nonBillable"`
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

	rangeStart := beginningOfMonth
	rangeEnd := beginningOfMonth.AddDate(0, 1, 0)

	chartSeriesMap := make(map[string]*chartSeries)

	date := rangeStart
	for date.Before(rangeEnd) {
		dateStr := date.Format("2006-01-02")
		chartSeriesMap[dateStr] = &chartSeries{
			Date:            date,
			BillableTime:    0,
			NonBillableTime: 0,
		}
		date = date.AddDate(0, 0, 1)
	}

	for _, item := range res.User.Billable {
		dateStr := item.Epoch.Format("2006-01-02")
		d := time.Duration(item.Min) * time.Minute
		if existing, ok := chartSeriesMap[dateStr]; ok {
			existing.BillableTime = d
		} else {
			return fmt.Errorf("date %s out of range", dateStr)
		}
	}

	for _, item := range res.User.NonBillable {
		dateStr := item.Epoch.Format("2006-01-02")
		d := time.Duration(item.Min) * time.Minute
		if existing, ok := chartSeriesMap[dateStr]; ok {
			existing.NonBillableTime = d
		} else {
			return fmt.Errorf("date %s out of range", dateStr)
		}
	}

	var chartSeriesList []chartSeries
	for _, v := range chartSeriesMap {
		chartSeriesList = append(chartSeriesList, *v)
	}

	sort.Slice(chartSeriesList, func(i, j int) bool {
		return chartSeriesList[i].Date.Before(chartSeriesList[j].Date)
	})

	renderReportAsTable(chartSeriesList)

	return nil
}

func renderReportAsTable(chartSeriesList []chartSeries) {
	tableRows := [][]string{}

	var totalBillableTime time.Duration
	var totalNonBillableTime time.Duration

	for _, item := range chartSeriesList {
		totalBillableTime += item.BillableTime
		totalNonBillableTime += item.NonBillableTime

		var billableTime string
		if item.BillableTime > 0 {
			billableTime = formatDuration(item.BillableTime)
		}

		var nonBillableTime string
		if item.NonBillableTime > 0 {
			nonBillableTime = formatDuration(item.NonBillableTime)
		}

		var dayColor func(a ...interface{}) string
		switch item.Date.Weekday() {
		case time.Saturday, time.Sunday:
			dayColor = color.New(color.FgRed).SprintFunc()
		default:
			dayColor = color.New(color.FgWhite).SprintFunc()
		}

		tableRows = append(tableRows, []string{
			dayColor(item.Date.Format("2006-01-02")),
			billableTime,
			nonBillableTime,
		})
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetAlignment(tablewriter.ALIGN_RIGHT)
	table.SetHeaderAlignment(tablewriter.ALIGN_CENTER)
	table.SetFooterAlignment(tablewriter.ALIGN_RIGHT)

	table.SetHeader([]string{"Date", "Billable", "Non-Billable"})
	table.AppendBulk(tableRows)
	table.SetFooter([]string{"Total", formatDuration(totalBillableTime), formatDuration(totalNonBillableTime)})

	table.Render()
}

func formatDuration(d time.Duration) string {
	return fmt.Sprintf("%2d:%02d", int(d.Hours()), int(d.Minutes())%60)
}
