package main

import (
	"fmt"
	"os"

	"github.com/boxproject/companion/commands"
	"gopkg.in/urfave/cli.v1"
)

func main() {
	commands.InitLogger()
	app := newApp()
	app.Run(os.Args)
}

func newApp() *cli.App {
	app := cli.NewApp()
	app.Version = PrintVersion(gitCommit, stage, version)
	app.Name = "Blockchain monitor"
	app.Usage = "The blockchain monitor command line interface"
	app.Author = "2SE Group"
	app.Copyright = "Copyright 2017-2018 The exchange Authors"
	app.Email = "support@2se.com"
	app.Description = "blockchain monitor"

	app.Commands = []cli.Command{
		// 启动
		{
			Name:   "start",
			Usage:  "start the monitor",
			Action: commands.StartCmd,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "config,c",
					Usage: "Path of the config.json file",
					Value: "",
				},
				cli.StringFlag{
					Name:  "block-file,b",
					Usage: "Check point block number",
					Value: "",
				},
			},
		},
		// 停止
		{
			Name:   "stop",
			Usage:  "stop the monitor",
			Action: commands.StopCmd,
			Flags:  []cli.Flag{},
		},
	}

	return app
}

func PrintVersion(gitCommit, stage, version string) string {
	if gitCommit != "" {
		return fmt.Sprintf("%s-%s-%s", stage, version, gitCommit)
	}
	return fmt.Sprintf("%s-%s", stage, version)
}
