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

type DashBoard struct {
	From   lastpass_time.JsonLastPassTime `json:"from"`
	To     lastpass_time.JsonLastPassTime `json:"to"`
	Users  map[string][]api.User	`json:"users"`
	Departments map[string][]api.User		`json:"department"`
	Events map[string][]api.Event	`json:"events"`
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

	d := &DashBoard{
		Users: make(map[string][]api.User),
		Departments: make(map[string][]api.User),
		Events: make(map[string][]api.Event),
	}

	loc, _ := time.LoadLocation("Asia/Tokyo")
	now := time.Now().In(loc)
	dayAgo := now.Add(-time.Duration(durationToAuditInDay) * time.Hour * 24)
	//t := lastpass_time.JsonLastPassTime{JsonTime: now}
	//f := lastpass_time.JsonLastPassTime{JsonTime: dayAgo}
	d.From = lastpass_time.JsonLastPassTime{JsonTime: dayAgo}
	d.To = lastpass_time.JsonLastPassTime{JsonTime: now}

	//fmt.Println(
	//	fmt.Sprintf("----------------- Audits(%v ~ %v) -----------------", f.Format(), t.Format()))

	client := NewLastPassClientFromContext(c)
	s := NewService(client)

	//fmt.Println("# Admin Users")
	AdminUsers, err := s.GetAdminUserData()
	logger.ErrorIf(err)
	d.Users["admin_users"] = AdminUsers
	//PrettyPrintJSON(AdminUsers)

	//fmt.Println("# Disabled Users")
	disabledUsers, err := s.GetDisabledUser()
	logger.DieIf(err)

	d.Users["disabled_users"] = disabledUsers
	//for _, user := range disabledUsers {
	//	fmt.Println(user.UserName)
	//}

	//fmt.Println("# Inactive Users")
	inactiveUsers, err := s.GetInactiveUser()
	logger.ErrorIf(err)
	inactiveDep := make(map[string][]api.User)
	for _, u := range inactiveUsers {
		for _, group := range u.Groups {
			inactiveDep[group] = append(inactiveDep[group], u)
		}
	}
	for dep, users := range inactiveDep {
		//fmt.Println(fmt.Sprintf("%v : %v", dep, len(users)))
		d.Departments[dep] = users
	}

	// Get Shared Folder Data
	//fmt.Println("# Super Shared Folders must be only shared within Admin Users")
	res, err := client.GetSharedFolderData()
	logger.ErrorIf(err)

	var sharedFolders map[string]api.SharedFolder
	err = JSONBodyDecoder(res, &sharedFolders)
	logger.ErrorIf(err)

	for _, sf := range sharedFolders {
		if sf.ShareFolderName != "Super-Admins" {
			break
		}

		d.Users["super_shared_folder_users"] = sf.Users

		if len(sf.Users) > len(AdminUsers) {
			//fmt.Println("Some non-admin who joins Super Shared Folders")
			//PrettyPrintJSON(sf.Users)
		} else {
			for _, admin := range AdminUsers {
				flag := false
				for _, susers := range sf.Users {
					if admin.UserName != susers.UserName {
						flag = true
					}
				}
				if !flag {
					//fmt.Println("Some one who is non-admin is in Super Shared Folders")
					//PrettyPrintJSON(sf.Users)
				}
			}
		}
	}

	//fmt.Println("# Events")
	res, err = client.GetEventReport("", "", d.From, d.To)
	logger.ErrorIf(err)

	var result api.Events
	err = JSONBodyDecoder(res, &result)
	logger.ErrorIf(err)

	for _, event := range result.Events {
		d.Events[event.Username] = result.Events

		if event.IsAuditEvent() {
			//fmt.Println(event)
		}
	}

	GetLoginEvent := func(wg *sync.WaitGroup, q chan string) {
		defer wg.Done()
		for {
			userName, ok := <-q
			if !ok {
				return
			}
			res, err = client.GetEventReport(userName, "", d.From, d.To)
			//fmt.Println(fmt.Sprintf("## %v Login History", userName))
			logger.ErrorIf(err)
			//if err != nil {
			//	fmt.Println(err)
			//	return
			//}
			err = JSONBodyDecoder(res, &result)
			logger.ErrorIf(err)
			//if err != nil {
			//	fmt.Println(err)
			//	return
			//}
			//for _, event := range result.Events {
			//	fmt.Println(event)
			//}
			d.Events[userName] = result.Events
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

	out := fmt.Sprintf("# Admin Users And Activities\n")
	for _, u := range d.Users["admin_users"] {
		out = out + fmt.Sprintf("## %v: \n", u.UserName)
		for us, events := range d.Events {
			if us == u.UserName {
				for _, event := range events{
					out = out + fmt.Sprintf("%v\n",event)
				}
			}
		}
	}
	out = out + fmt.Sprintf("\n# Disabled Users\n")
	for _, u := range d.Users["disabled_users"] {
		out = out + fmt.Sprintf("## %v: \n", u.UserName)
	}
	out = out + fmt.Sprintf("\n# Inactive Users")
	for dep, us := range d.Departments {
		out = out + fmt.Sprintf("\n## %v: %v\n", dep, len(us))
		for _, u := range us {
			out = out + fmt.Sprintf(u.UserName + ", ")
		}
	}

	//PrettyPrintJSON(d.Users)
	fmt.Println(out)
	return nil
}