package main

import (
	"fmt"
	"github.com/urfave/cli"
	"lastpass_provisioning/api"
	"lastpass_provisioning/lastpass_time"
	"lastpass_provisioning/logger"
	"sync"
	"time"
	"os"
)

func init() {
	// Requirements:
	// - .Description: First and last line is blank.
	// - .ArgsUsage: ArgsUsage includes flag usages (e.g. [-v|verbose] <hostId>).
	//   All cli.Command should have ArgsUsage field.
	cli.CommandHelpTemplate = `NAME:
   {{.HelpName}} - {{.Usage}}
USAGE:
   {{.HelpName}} {{if .ArgsUsage}}{{.ArgsUsage}}{{else}}[arguments...]{{end}}{{if .Description}}
DESCRIPTION:{{.Description}}{{end}}{{if .VisibleFlags}}
OPTIONS:
   {{range .VisibleFlags}}{{.}}
   {{end}}{{end}}
`
}

// Commands cli.Command object list
var Commands = []cli.Command{
	//commandStatus,
	//commandHosts,
	//commandCreate,
	//commandUpdate,
	//commandThrow,
	//commandFetch,
	//commandRetire,
	//commandServices,
	//commandMonitors,
	//commandAlerts,
	commandDashboards,
	//commandAnnotations,
}

var commandDashboards = cli.Command{
	Name:        "dashboard",
	Usage:       "Report summary",
	ArgsUsage:   "[--verbose | -v] [--period | -d <duration>]",
	Description: `show audit related dashboard`,
	Action:      doDashboard,
	Flags:       []cli.Flag{
		cli.IntFlag{Name:  "duration, d", Usage: "Audits for past <duration> day"},
		cli.BoolFlag{Name: "verbose, v", Usage: "Verbose output mode"},
	},
}

// TODO refactor,
func doDashboard(c *cli.Context) error {
	if c.Bool("verbose") {
		os.Setenv("DEBUG", "1")
	}

	durationToAuditInDay := 1
	if c.Int("duration") >= 1 {
		durationToAuditInDay = c.Int("duration")
	}

	loc, _ := time.LoadLocation("Asia/Tokyo")
	now := time.Now().In(loc)
	dayAgo := now.Add(-time.Duration(durationToAuditInDay) * time.Hour * 24)
	t := lastpass_time.JsonLastPassTime{JsonTime: now}
	f := lastpass_time.JsonLastPassTime{JsonTime: dayAgo}
	fmt.Println(
		fmt.Sprintf("----------------- Audits(%v ~ %v) -----------------", f.Format(), t.Format()))

	client := NewLastPassClientFromContext(c)
	s := NewService(client)

	fmt.Println("# Disabled Users")
	disabledUsers, err := s.GetDisabledUser()

	for _, user := range disabledUsers {
		fmt.Println(user.UserName)
	}

	fmt.Println("# Inactive Users")
	inactiveUsers, err := s.GetInactiveUser()
	inactiveDep := make(map[string][]api.User)
	for _, u := range inactiveUsers {
		for _, group := range u.Groups {
			inactiveDep[group] = append(inactiveDep[group], u)
		}
	}
	for dep, users := range inactiveDep {
		fmt.Println(fmt.Sprintf("%v : %v", dep, len(users)))
	}

	fmt.Println("# Admin Users")
	AdminUsers, err := s.GetAdminUserData()
	logger.DieIf(err)
	PrettyPrintJSON(AdminUsers)

	// Get Shared Folder Data
	fmt.Println("# Super Shared Folders must be only shared within Admin Users")
	res, err := client.GetSharedFolderData()
	if err != nil {
		logger.ErrorIf(err)
	}
	var sharedFolders map[string]api.SharedFolder
	err = DecodeBody(res, &sharedFolders)
	if err != nil {
		logger.ErrorIf(err)
	}
	for _, sf := range sharedFolders {
		if sf.ShareFolderName != "Super-Admins" {
			break
		}

		if len(sf.Users) > len(AdminUsers) {
			fmt.Println("Some non-admin who joins Super Shared Folders")
			PrettyPrintJSON(sf.Users)
		} else {
			for _, admin := range AdminUsers {
				flag := false
				for _, susers := range sf.Users {
					if admin.UserName != susers.UserName {
						flag = true
					}
				}
				if !flag {
					fmt.Println("Some one who is non-admin is in Super Shared Folders")
					PrettyPrintJSON(sf.Users)
				}
			}
		}
	}

	fmt.Println("# Events")
	res, err = client.GetEventReport("", "", f, t)
	if err != nil {
		logger.ErrorIf(err)
	}

	var result api.Events
	err = DecodeBody(res, &result)
	if err != nil {
		logger.ErrorIf(err)
	}
	for _, event := range result.Events {
		if event.IsAuditEvent() {
			fmt.Println(event)
		}
	}

	GetLoginEvent := func(wg *sync.WaitGroup, q chan string) {
		defer wg.Done()
		for {
			userName, ok := <-q
			if !ok {
				return
			}
			res, err = client.GetEventReport(userName, "", f, t)
			fmt.Println(fmt.Sprintf("## %v Login History", userName))
			if err != nil {
				fmt.Println(err)
				return
			}
			err = DecodeBody(res, &result)
			if err != nil {
				fmt.Println(err)
				return
			}
			for _, event := range result.Events {
				fmt.Println(event)
			}
		}
	}

	var wg sync.WaitGroup
	q := make(chan string, 5)
	for i := 0; i < len(AdminUsers); i++ {
		wg.Add(1)
		go GetLoginEvent(&wg, q)
	}

	for _, admin := range AdminUsers {
		q <- admin.UserName
	}
	close(q)
	wg.Wait()
	return nil
}
