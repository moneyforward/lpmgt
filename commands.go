package main

import (
	"github.com/urfave/cli"
	"fmt"
	"lastpass_provisioning/logger"
)

var client *LastpassClient

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
	Name:	"dashboard",
	Usage:  "Report summary",
	ArgsUsage: "",
	Description: `show audit related dashboard`,
	Action: doDashboard,
	Flags: []cli.Flag{},
}

func doDashboard(c *cli.Context) error {
	client := NewLastPassClientFromContext(c)
	s := NewService(client)

	fmt.Println(" --------------------  Admin Users -------------------- ")
	AdminUsers, err := s.GetAdminUserData()
	logger.DieIf(err)

	for _, admin := range AdminUsers {
		fmt.Println(admin.UserName)
	}
	return nil
}