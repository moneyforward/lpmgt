package main

import (
	"github.com/urfave/cli"
	"os"
)

func main() {
	app := cli.NewApp()
	app.Name = "lastpass"
	app.Version = version
	app.Usage = "A CLI tool for Lastpass(Enterprise)"
	app.Author = "Money Forward Co., Ltd."
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "config, c",
			Usage: "Load configuration from `FILE`",
		},
	}
	app.Commands = Commands
	app.Run(os.Args)
}