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
	"github.com/pkg/errors"
	"io/ioutil"
	"encoding/json"
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
	commandDashboards,
	commandCreate,
}

var commandCreate = cli.Command {
	Name:	"create",
	Usage:	"Create a new object",
	Subcommands: []cli.Command {
		subcommandCreateUser,
	},
}

var subcommandCreateUser = cli.Command{
	Name:	"user",
	Usage:	"hoge",
	ArgsUsage: "[--bulk | -b] <file>",
	Description:`To register users in a bulk, please specify a yaml file.`,
	Action: doAddUserFroms,
	Flags:	[]cli.Flag{
		cli.StringFlag{Name: "bulk, b", Value: "", Usage: "Load users from a JSON <file>"},
	},
}

func doAddUserFroms(c *cli.Context) error {
	if c.String("bulk") == "" {
		fmt.Println(errors.New("Need to specify file name."))
		return errors.New("Need to specify file name.")
	}

	hoge, err := loadConfir(c.String("bulk"))
	if err != nil {
		fmt.Println(err)
		return err
	}

	client := NewLastPassClientFromContext(c)
	s := NewService(client)
	fmt.Println(hoge)
	if err := s.BatchAdd(hoge); err != nil {
		fmt.Println(err)
	}
	return err
}

func loadConfir(configFile string) (config []api.User, err error) {
	f, err := ioutil.ReadFile(configFile)
	if err != nil {
		return
	}

	hoge := struct {
		Data []api.User `json:"data"`
	}{}

	//buf := new(bytes.Buffer)
	err = json.Unmarshal(f, &hoge)
	config = hoge.Data
	//err = yaml.Unmarshal(f, &config)
	return
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
	d.From = lastpass_time.JsonLastPassTime{JsonTime: dayAgo}
	d.To = lastpass_time.JsonLastPassTime{JsonTime: now}

	client := NewLastPassClientFromContext(c)
	s := NewService(client)

	AdminUsers, err := s.GetAdminUserData()
	logger.ErrorIf(err)
	d.Users["admin_users"] = AdminUsers

	disabledUsers, err := s.GetDisabledUser()
	logger.DieIf(err)
	d.Users["disabled_users"] = disabledUsers

	inactiveUsers, err := s.GetInactiveUser()
	logger.ErrorIf(err)
	inactiveDep := make(map[string][]api.User)
	for _, u := range inactiveUsers {
		for _, group := range u.Groups {
			inactiveDep[group] = append(inactiveDep[group], u)
		}
	}
	for dep, users := range inactiveDep {
		d.Departments[dep] = users
	}

	// Get Shared Folder Data
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
	}

	res, err = client.GetEventReport("", "", d.From, d.To)
	logger.ErrorIf(err)

	var result api.Events
	err = JSONBodyDecoder(res, &result)
	logger.ErrorIf(err)

	for _, event := range result.Events {
		d.Events[event.Username] = append(d.Events[event.Username], event)
	}

	GetLoginEvent := func(wg *sync.WaitGroup, q chan string) {
		defer wg.Done()
		for {
			userName, ok := <-q
			if !ok {
				return
			}
			res, err = client.GetEventReport(userName, "", d.From, d.To)
			logger.ErrorIf(err)

			err = JSONBodyDecoder(res, &result)
			logger.ErrorIf(err)
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

	admins := []string{}
	out := fmt.Sprintf("# Admin Users And Activities\n")
	for _, admin := range d.Users["admin_users"] {
		out = out + fmt.Sprintf("## %v: \n", admin.UserName)
		admins = append(admins, admin.UserName)
		for user, events := range d.Events {
			if admin.UserName == user {
				for _, event := range events {
					out = out + fmt.Sprintf("%v\n", event)
				}
			}
		}
	}

	out = out + fmt.Sprintf("\n# Audit Events\n")
	for _, events := range d.Events {
		for _, event := range events {
			if event.IsAuditEvent() {
				out = out + fmt.Sprintf("%v\n", event)
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

	fmt.Println(out)
	return nil
}