package main

import (
	"fmt"
	"os"
	"time"

	"github.com/streed/cli-do-client/internal/clido"

	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:    "cli-do",
		Usage:   "A CLI for CLI Do",
		Version: "0.1.0",
		Authors: []*cli.Author{
			{
				Name:  "Reed",
				Email: "support@cli-do.com",
			},
		},
		Compiled:             time.Now(),
		EnableBashCompletion: true,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "project",
				Aliases: []string{"p"},
				Usage:   "Project ID",
			},
		},
		Commands: []*cli.Command{
			{
				Name:    "login",
				Aliases: []string{"l"},
				Usage:   "Login to cli-do",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "force",
						Usage:   "Force to refresh login if already logged in",
						Aliases: []string{"f"},
					},
				},
				Action: clido.HandleLogin,
			},
			{
				Name:    "todo",
				Aliases: []string{"t"},
				Usage:   "Todo operations",
				Subcommands: []*cli.Command{
					{
						Name:    "list",
						Aliases: []string{"ls"},
						Flags: []cli.Flag{
							&cli.BoolFlag{
								Name:    "all",
								Aliases: []string{"a"},
							},
						},
						Action: clido.HandleTodosList,
					},
					{
						Name:      "get",
						ArgsUsage: "<ticket>",
						Aliases:   []string{"g"},
						Action:    clido.HandleGetTodo,
					},
					{
						Name:    "edit",
						Aliases: []string{"e"},
						Action:  clido.HandleEditTodo,
					},
					{
						Name:    "new",
						Aliases: []string{"n"},
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:    "subject",
								Aliases: []string{"s"},
								Usage:   "Subject of the todo",
							},
							&cli.StringFlag{
								Name:    "body",
								Aliases: []string{"b"},
								Usage:   "Body of the todo",
							},
							&cli.TimestampFlag{
								Name:    "due-date",
								Aliases: []string{"d"},
								Usage:   "Due date of the todo",
								Layout:  "2006-01-02",
							},
						},
						Action: clido.HandleCreateTodo,
					},
					{
						Name:    "archive",
						Aliases: []string{"a"},
						Action:  clido.HandleArchiveTodo,
					},
					{
						Name:    "complete",
						Aliases: []string{"co"},
						Action:  clido.HandleCompleteTodo,
					},
				},
			},
			{
				Name:    "project",
				Usage:   "Project operations",
				Aliases: []string{"p"},
				Subcommands: []*cli.Command{
					{
						Name:    "init",
						Aliases: []string{"i"},
						Action: func(ctx *cli.Context) error {
							return clido.HandleInitProjectDirectory(ctx)
						},
					},
					{
						Name:    "new",
						Aliases: []string{"n"},
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:    "name",
								Aliases: []string{"n"},
								Usage:   "Name of the project",
							},
							&cli.StringFlag{
								Name:    "description",
								Aliases: []string{"d"},
								Usage:   "Description of the project",
							},
						},
						Action: clido.HandleProjectNew,
					},
					{
						Name:    "list",
						Aliases: []string{"ls"},
						Action:  clido.HandleProjectList,
					},
					{
						Name:      "archive",
						Aliases:   []string{"a"},
						ArgsUsage: "<project_id>",
						Action:    clido.HandleProjectArchive,
					},
				},
			},
		},
		Action: func(*cli.Context) error {
			fmt.Println("Hello, cli-do! Run 'cli-do help' for more information.")
			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Println(err)
	}
}
