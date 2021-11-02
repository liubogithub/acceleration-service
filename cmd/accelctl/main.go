// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"
	"os"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"

	"github.com/goharbor/acceleration-service/pkg/client"
)

var versionGitCommit string
var versionBuildTime string

var ctl *client.Client

func main() {
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: time.RFC3339Nano,
	})

	version := fmt.Sprintf("%s.%s", versionGitCommit, versionBuildTime)

	app := &cli.App{
		Name:    "accelctl",
		Usage:   "A CLI tool to manage acceleration service",
		Version: version,
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "addr", Value: "localhost:2077", Usage: "Service address in format <host:port>."},
		},
		Before: func(c *cli.Context) error {
			ctl = client.NewClient(c.String("addr"))
			return nil
		},
		Commands: []*cli.Command{
			{
				Name:  "task",
				Usage: "Manage conversion tasks",
				Subcommands: []*cli.Command{
					{
						Name:  "create",
						Usage: "Create an acceleration image conversion task",
						Flags: []cli.Flag{
							&cli.StringFlag{Name: "source", Required: true, Usage: "Source image reference"},
							&cli.BoolFlag{Name: "sync", Value: false},
						},
						Action: func(c *cli.Context) error {
							sync := c.Bool("sync")
							if sync {
								logrus.Info("Waiting task to be completed...")
							}
							if err := ctl.CreateTask(c.String("source"), sync); err != nil {
								return err
							}
							if sync {
								logrus.Info("Task has been completed.")
							} else {
								logrus.Info("Submitted asynchronous task, check status by `task list`.")
							}
							return nil
						},
					},
					{
						Name:    "list",
						Aliases: []string{"ls"},
						Usage:   "List acceleration image conversion tasks",
						Action: func(c *cli.Context) error {
							tasks, err := ctl.ListTask()
							if err != nil {
								return err
							}

							for _, task := range tasks {
								created := task.Created.Format("2006-01-02 15:04:05")
								fmt.Printf("%d\t%s\t%s\t%s\t%s\t\n", task.ID, created, task.Status, task.Source, task.Reason)
							}

							return nil
						},
					},
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		logrus.Fatal(err)
	}
}