package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/harnyk/teamjerk/internal/app"
	"github.com/harnyk/teamjerk/internal/authstore"
	"github.com/harnyk/teamjerk/internal/twapi"
	"github.com/spf13/cobra"
)

//this will be replaced in the goreleaser build
var version = "development"

func getAuthFilePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(home, ".teamjerk", "auth.json"), nil
}

func main() {
	authFilePath, err := getAuthFilePath()
	if err != nil {
		log.Fatal(err)
	}

	tw := twapi.NewClient()
	store := authstore.NewAuthStore[twapi.AuthData](authFilePath)
	app := app.NewApp(tw, store)

	rootCmd := &cobra.Command{
		Use:   "teamjerk",
		Short: "A command line tool for Teamwork.com",
		Long:  `A command line tool for Teamwork.com`,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	loginCmd := &cobra.Command{
		Use:   "login",
		Short: "Login to Teamwork.com",
		Long:  `Login to Teamwork.com`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.LogIn()
		},
	}

	logoutCmd := &cobra.Command{
		Use:   "logout",
		Short: "Logout from Teamwork.com",
		Long:  `Logout from Teamwork.com`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.LogOut()
		},
	}

	whoamiCmd := &cobra.Command{
		Use:   "whoami",
		Short: "Show the currently logged in user",
		Long:  `Show the currently logged in user`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.WhoAmI()
		},
	}

	projectsCmd := &cobra.Command{
		Use:   "projects",
		Short: "List all projects",
		Long:  `List all projects`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.Projects()
		},
	}

	tasksCmd := &cobra.Command{
		Use:   "tasks",
		Short: "List all tasks",
		Long:  `List all tasks`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.Tasks()
		},
	}

	logCmd := &cobra.Command{
		Use:   "log",
		Short: "Log time",
		Long:  `Log time`,
		RunE: func(cmd *cobra.Command, args []string) error {
			dryRun, err := cmd.Flags().GetBool("dry-run")
			if err != nil {
				return err
			}
			nonBillable, err := cmd.Flags().GetBool("non-billable")
			if err != nil {
				return err
			}

			projectID, err := cmd.Flags().GetInt("project-id")
			if err != nil {
				return err
			}

			taskID, err := cmd.Flags().GetInt("task-id")
			if err != nil {
				return err
			}

			dateS, err := cmd.Flags().GetString("date")
			if err != nil {
				return err
			}
			date, err := time.Parse("2006-01-02", dateS)
			if err != nil {
				return err
			}

			timeS, err := cmd.Flags().GetString("time")
			if err != nil {
				return err
			}
			startTime, err := time.Parse("15:04", timeS)
			if err != nil {
				return err
			}

			hours, err := cmd.Flags().GetFloat64("hours")
			if err != nil {
				return err
			}
			if hours <= 0 || hours > 24 {
				return fmt.Errorf("hours must be between 0 and 24")
			}

			description, err := cmd.Flags().GetString("description")
			if err != nil {
				return err
			}

			//TODO: implement a structure parameter for the log command

			return app.Log()
		},
	}
	logCmd.Flags().BoolP("dry-run", "n", false, "Don't actually log time")
	logCmd.Flags().BoolP("non-billable", "B", false, "Log time as non-billable")
	logCmd.Flags().IntP("project-id", "p", 0, "Project ID")
	logCmd.Flags().IntP("task-id", "t", 0, "Task ID")
	logCmd.Flags().StringP("date", "d", "", "Date (e.g. 2020-01-31)")
	logCmd.Flags().StringP("time", "t", "", "Start time (e.g. 09:00)")
	logCmd.Flags().Float64P("hours", "h", 0, "Number of logged hours (e.g. 8.5)")
	logCmd.Flags().StringP("description", "d", "", "Description")

	reportCmd := &cobra.Command{
		Use:   "report",
		Short: "Report time",
		Long:  `Report time`,
		RunE: func(cmd *cobra.Command, args []string) error {
			now := time.Now()
			defaultYear := now.Year()
			defaultMonth := int(now.Month())

			year, err := cmd.Flags().GetInt("year")
			if err != nil {
				return err
			}
			if year == 0 {
				year = defaultYear
			}
			if year < 2000 || year > 2100 {
				return fmt.Errorf("invalid year: %d", year)
			}

			month, err := cmd.Flags().GetInt("month")
			if err != nil {
				return err
			}
			if month == 0 {
				month = defaultMonth
			}
			if month < 1 || month > 12 {
				return fmt.Errorf("invalid month: %d", month)
			}

			beginningOfMonth := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)

			return app.Report(beginningOfMonth)
		},
	}
	reportCmd.Flags().IntP("year", "y", time.Now().Year(), "Year to report")
	reportCmd.Flags().IntP("month", "m", int(time.Now().Month()), "Month to report")

	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version number of teamjerk",
		Long:  `All software has versions. This is teamjerk's`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.Println(version)
			return nil
		},
	}

	rootCmd.AddCommand(loginCmd)
	rootCmd.AddCommand(logoutCmd)
	rootCmd.AddCommand(whoamiCmd)
	rootCmd.AddCommand(projectsCmd)
	rootCmd.AddCommand(tasksCmd)
	rootCmd.AddCommand(logCmd)
	rootCmd.AddCommand(reportCmd)
	rootCmd.AddCommand(versionCmd)

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}

}
