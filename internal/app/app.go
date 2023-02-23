package app

import (
	"fmt"
	"os"
	"sort"
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

	type chartSeries struct {
		date            time.Time
		billableTime    time.Duration
		nonBillableTime time.Duration
	}

	chartSeriesMap := make(map[string]*chartSeries)

	date := rangeStart
	for date.Before(rangeEnd) {
		dateStr := date.Format("2006-01-02")
		chartSeriesMap[dateStr] = &chartSeries{
			date:            date,
			billableTime:    0,
			nonBillableTime: 0,
		}
		date = date.AddDate(0, 0, 1)
	}

	for _, item := range res.User.Billable {
		dateStr := item.Epoch.Format("2006-01-02")
		d := time.Duration(item.Min) * time.Minute
		if existing, ok := chartSeriesMap[dateStr]; ok {
			existing.billableTime = d
		} else {
			return fmt.Errorf("date %s out of range", dateStr)
		}
	}

	for _, item := range res.User.NonBillable {
		dateStr := item.Epoch.Format("2006-01-02")
		d := time.Duration(item.Min) * time.Minute
		if existing, ok := chartSeriesMap[dateStr]; ok {
			existing.nonBillableTime = d
		} else {
			return fmt.Errorf("date %s out of range", dateStr)
		}
	}

	var chartSeriesList []chartSeries
	for _, v := range chartSeriesMap {
		chartSeriesList = append(chartSeriesList, *v)
	}

	sort.Slice(chartSeriesList, func(i, j int) bool {
		return chartSeriesList[i].date.Before(chartSeriesList[j].date)
	})

	//Output

	tableRows := [][]string{}

	var totalBillableTime time.Duration
	var totalNonBillableTime time.Duration

	for _, item := range chartSeriesList {
		totalBillableTime += item.billableTime
		totalNonBillableTime += item.nonBillableTime

		var billableTime string
		if item.billableTime > 0 {
			billableTime = formatDuration(item.billableTime)
		}

		var nonBillableTime string
		if item.nonBillableTime > 0 {
			nonBillableTime = formatDuration(item.nonBillableTime)
		}

		var dayColor func(a ...interface{}) string
		switch item.date.Weekday() {
		case time.Saturday, time.Sunday:
			dayColor = color.New(color.FgRed).SprintFunc()
		default:
			dayColor = color.New(color.FgWhite).SprintFunc()
		}

		tableRows = append(tableRows, []string{
			dayColor(item.date.Format("2006-01-02")),
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

	return nil
}

func formatDuration(d time.Duration) string {
	return fmt.Sprintf("%2d:%02d", int(d.Hours()), int(d.Minutes())%60)
}
