package main

import (
	"github.com/getsentry/raven-go"
	"os"
	"proctor/internal/app/service/infra/config"
	"proctor/internal/app/service/infra/db/migration"
	"proctor/internal/app/service/infra/logger"
	"proctor/internal/app/service/server"
	"proctor/internal/app/service/worker"

	"github.com/urfave/cli"
)

func main() {
	logger.Setup()
	_ = raven.SetDSN(config.SentryDSN())

	proctord := cli.NewApp()
	proctord.Name = "proctord"
	proctord.Usage = "Handle executing jobs and maintaining their configuration"
	proctord.Version = "0.2.0"
	proctord.Commands = []cli.Command{
		{
			Name:        "migrate",
			Description: "Run database migrations for proctord",
			Action: func(c *cli.Context) {
				err := migration.Up()
				if err != nil {
					panic(err.Error())
				}
				logger.Info("Migration successful")
			},
		},
		{
			Name:        "rollback",
			Description: "Rollback database migrations by one step for proctord",
			Action: func(c *cli.Context) {
				err := migration.DownOneStep()
				if err != nil {
					panic(err.Error())
				}
				logger.Info("Rollback successful")
			},
		},
		{
			Name:    "start",
			Aliases: []string{"s"},
			Usage:   "starts server",
			Action: func(c *cli.Context) error {
				return server.Start()
			},
		},
		{
			Name:  "start-scheduler",
			Usage: "starts scheduler",
			Action: func(c *cli.Context) error {
				return worker.Start()
			},
		},
	}

	_ = proctord.Run(os.Args)
}
