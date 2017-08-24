package main

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
	"io/ioutil"
	"lastpass_provisioning/api"
	"lastpass_provisioning/lastpass_time"
	"lastpass_provisioning/logger"
	"os"
	"sync"
	"time"
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
	commandGet,
	commandDescribe,
	commandDelete,
	commandUpdate,
}

// Update command with subcommands
var commandUpdate = cli.Command{
	Name:  "update",
	Usage: "update specific object",
	Subcommands: []cli.Command{
		subCommandUpdateUser,
	},
}

var subCommandUpdateUser = cli.Command{
	Name:        "user",
	Usage:       "update user <email>",
	Description: `update a <email>`,
	ArgsUsage:   "<email>",
	Subcommands: []cli.Command{
		{
			Name:        "transfer",
			Usage:       "transfer user",
			ArgsUsage:   "[[--leave | -l <department>]...] [[--join | -j <department>]...]",
			Description: "User can specify either --leave or --join to move department",
			Action:      doTransferUser,
			Flags: []cli.Flag{
				cli.StringSliceFlag{Name: "leave, l", Value: &cli.StringSlice{}, Usage: "leave current department"},
				cli.StringSliceFlag{Name: "join, j", Value: &cli.StringSlice{}, Usage: "join new department"},
			},
		},
	},
}

func doTransferUser(c *cli.Context) error {
	argUserName := c.Args().Get(0)
	if argUserName == "" {
		logger.DieIf(errors.New("Email(username) has to be specified"))
	}

	client := NewLastPassClientFromContext(c)
	s := NewService(client)

	// Fetch User if he/she exists
	user, err := s.GetUserData(argUserName)
	logger.DieIf(err)

	// Join
	user.Groups = append(user.Groups, c.StringSlice("join")...)

	// Leave
	leave := c.StringSlice("leave")
	for i := 0; i < len(leave); i++ {
		newDeps := []string{}
		for _, dep := range user.Groups {
			if dep != leave[i] {
				newDeps = append(newDeps, dep)
			}
		}
		user.Groups = newDeps
	}

	// Update
	err = s.UpdateUser(user)
	logger.DieIf(err)
	logger.Log("updated", user.UserName)
	return nil
}

// Delete command with subcommands
var commandDelete = cli.Command{
	Name:  "delete",
	Usage: "delete specific object",
	Subcommands: []cli.Command{
		subCommandDeleteUser,
	},
}

var subCommandDeleteUser = cli.Command{
	Name:        "user",
	Usage:       "delete user <email>",
	Description: `delete a <email> by choosing either 'deactivate(default)' or 'delete'`,
	ArgsUsage:   "[--mode | -m <deleteMode>] <email>",
	Action:      doDeleteUser,
	Flags: []cli.Flag{
		cli.StringFlag{Name: "mode, m", Value: "deactivate", Usage: "deleteMode"},
	},
}

func doDeleteUser(c *cli.Context) error {
	argUserName := c.Args().Get(0)
	if argUserName == "" {
		logger.DieIf(errors.New("Email(username) has to be specified"))
	}

	var mode = DeactivationMode(Deactivate)
	switch c.String("mode") {
	case "deactivate":
		mode = Deactivate
	case "delete":
		mode = Delete
	default:
		mode = Deactivate
	}

	client := NewLastPassClientFromContext(c)
	err := NewService(client).DeleteUser(argUserName, mode)
	logger.DieIf(err)
	logger.Log(c.String("mode"), argUserName)
	return nil
}

// Describe command with subcommands
var commandDescribe = cli.Command{
	Name:  "describe",
	Usage: "describe specific object",
	Subcommands: []cli.Command{
		subCommandDescribeUser,
	},
}

var subCommandDescribeUser = cli.Command{
	Name:        "user",
	Usage:       "describe user",
	Description: `Show the information of the user with <email>`,
	ArgsUsage:   "<email>",
	Action:      doDescribeUser,
}

func doDescribeUser(c *cli.Context) error {
	argUserName := c.Args().Get(0)
	if argUserName == "" {
		logger.DieIf(errors.New("Email(username) has to be specified"))
	}

	client := NewLastPassClientFromContext(c)
	user, err := NewService(client).GetUserData(argUserName)
	logger.DieIf(err)

	PrintIndentedJSON(user)
	return nil
}

// Get command with subcommands
var commandGet = cli.Command{
	Name:  "get",
	Usage: "Get objects",
	Subcommands: []cli.Command{
		subcommandGetUsers,
	},
}

var subcommandGetUsers = cli.Command{
	Name:        "users",
	Usage:       "get users",
	ArgsUsage:   "[--filter, -f <option>]",
	Description: "Use --filter to filter users. You can choose from either `non2fa`, `inactive`, `disabled`, or `admin`",
	Action:      doGetUsers,
	Flags: []cli.Flag{
		cli.StringFlag{Name: "filter, f", Value: "all", Usage: "Filter fetching users"},
	},
}

func doGetUsers(c *cli.Context) (err error) {
	client := NewLastPassClientFromContext(c)
	s := NewService(client)

	var users []api.User

	switch c.String("filter") {
	case "non2fa":
		users, err = s.GetNon2faUsers()
	case "inactive":
		users, err = s.GetInactiveUsers()
	case "disabled":
		users, err = s.GetDisabledUsers()
	case "admin":
		users, err = s.GetAdminUserData()
	default:
		users, err = s.GetAllUsers()
	}
	logger.DieIf(err)
	api.PrintUserNames(users)
	return nil
}

var commandCreate = cli.Command{
	Name:  "create",
	Usage: "Create a new object",
	Subcommands: []cli.Command{
		subcommandCreateUser,
	},
}

var subcommandCreateUser = cli.Command{
	Name:        "user",
	Usage:       "create an users",
	ArgsUsage:   "[--bulk | -b <file>] [--dept | -d <department>] <username>",
	Description: `Create one or more users specifying either username or pre-configured file.`,
	Action:      doAddUser,
	Flags: []cli.Flag{
		cli.StringFlag{Name: "email, e", Value: "", Usage: "Create user with <email>"},
		cli.StringSliceFlag{Name: "dept, d", Value: &cli.StringSlice{}, Usage: "Create user with <email> in <department>"},
		cli.StringFlag{Name: "bulk, b", Value: "", Usage: "Load users from a JSON <file>"},
	},
}

func doAddUser(c *cli.Context) error {
	if c.String("bulk") != "" {
		return doAddUsersInBulk(c)
	}

	argUserName := c.Args().Get(0)
	if argUserName == "" {
		logger.DieIf(errors.New("Email(username) has to be specified"))
	}

	user := api.User{
		UserName: argUserName,
		Groups:   c.StringSlice("dept"),
	}

	client := NewLastPassClientFromContext(c)
	err := NewService(client).BatchAdd([]api.User{user})
	logger.DieIf(err)

	message := user.UserName
	for _, dep := range user.Groups {
		message += fmt.Sprintf(" in %v", dep)
	}
	logger.Log("created", message)
	return nil
}

func doAddUsersInBulk(c *cli.Context) error {
	users, err := loadAddingUsers(c.String("bulk"))
	if err != nil {
		logger.DieIf(err)
	}

	client := NewLastPassClientFromContext(c)
	err = NewService(client).BatchAdd(users)
	logger.DieIf(err)

	for _, user := range users {
		message := user.UserName
		for _, dep := range user.Groups {
			message += fmt.Sprintf(" in %v", dep)
		}
		logger.Log("created", message)
	}

	return err
}

func loadAddingUsers(usersFile string) (config []api.User, err error) {
	f, err := ioutil.ReadFile(usersFile)
	if err != nil {
		return
	}

	data := struct {
		Data []api.User `json:"data"`
	}{}

	if err = json.Unmarshal(f, &data); err != nil {
		return
	}
	config = data.Data
	return
}

var commandDashboards = cli.Command{
	Name:        "dashboard",
	Usage:       "Report summary",
	ArgsUsage:   "[--verbose | -v] [--period | -d <duration>]",
	Description: `show audit related dashboard`,
	Action:      doDashboard,
	Flags: []cli.Flag{
		cli.IntFlag{Name: "duration, d", Usage: "Audits for past <duration> day"},
		cli.BoolFlag{Name: "verbose, v", Usage: "Verbose output mode"},
	},
}

type dashBoard struct {
	From        lastpass_time.JsonLastPassTime `json:"from"`
	To          lastpass_time.JsonLastPassTime `json:"to"`
	Users       map[string][]api.User          `json:"users"`
	Departments map[string][]api.User          `json:"department"`
	Events      map[string][]api.Event         `json:"events"`
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

	d := &dashBoard{
		Users:       make(map[string][]api.User),
		Departments: make(map[string][]api.User),
		Events:      make(map[string][]api.Event),
	}

	loc, _ := time.LoadLocation("Asia/Tokyo")
	now := time.Now().In(loc)
	dayAgo := now.Add(-time.Duration(durationToAuditInDay) * time.Hour * 24)
	d.From = lastpass_time.JsonLastPassTime{JsonTime: dayAgo}
	d.To = lastpass_time.JsonLastPassTime{JsonTime: now}

	client := NewLastPassClientFromContext(c)
	s := NewService(client)

	AdminUsers, err := s.GetAdminUserData()
	logger.DieIf(err)
	d.Users["admin_users"] = AdminUsers

	disabledUsers, err := s.GetDisabledUsers()
	logger.DieIf(err)
	d.Users["disabled_users"] = disabledUsers

	inactiveUsers, err := s.GetInactiveUsers()
	logger.DieIf(err)
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
	logger.DieIf(err)

	var sharedFolders map[string]api.SharedFolder
	err = JSONBodyDecoder(res, &sharedFolders)
	logger.DieIf(err)

	for _, sf := range sharedFolders {
		if sf.ShareFolderName != "Super-Admins" {
			break
		}
		d.Users["super_shared_folder_users"] = sf.Users
	}

	res, err = client.GetEventReport("", "", d.From, d.To)
	logger.DieIf(err)

	var result api.Events
	err = JSONBodyDecoder(res, &result)
	logger.DieIf(err)

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
			logger.DieIf(err)

			err = JSONBodyDecoder(res, &result)
			logger.DieIf(err)
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
			out = out + fmt.Sprintf(u.UserName+", ")
		}
	}

	fmt.Println(out)
	return nil
}
