package main

import (
	"github.com/urfave/cli"
)

const (
	databaseFile = "tinycloud.sqlite"
)

var (
	Version string
)

func main() {
	app := cli.NewApp()

	app.Name = "tinycloud"
	app.Version = Version

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "data",
			EnvVar: "TINYCLOUD_DATA",
			Value:  "./data",
		},
	}

	app.Commands = []cli.Command{
		start(),
		users(),
	}

	app.RunAndExitOnError()

}
