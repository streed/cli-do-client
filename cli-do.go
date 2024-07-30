package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/aquilax/truncate"
	"github.com/go-resty/resty/v2"
	"github.com/rodaine/table"
	"github.com/urfave/cli/v2"
)

type Config struct {
	Endpoint string `json:"endpoint"`
	ClientId string `json:"client_id"`
}

type Login struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	ClientId string `json:"client_id"`
}

type Auth struct {
	Email        string `json:"email"`
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	CreatedAt    int    `json:"created_at"`
}

type Todos struct {
	Todos []Todo `json:"todos"`
}

type Todo struct {
	Id        string     `json:"id"`
	Subject   string     `json:"subject"`
	Ticket    int        `json:"ticket"`
	Body      string     `json:"body"`
	DueDate   *time.Time `json:"due_date"`
	Completed bool       `json:"completed"`
	PastDue   bool       `json:"past_due"`
}

type CreateTodo struct {
	Todo Todo `json:"todo"`
}

type Projects struct {
	Projects []Project `json:"projects"`
}

type Project struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Ticket      int    `json:"ticket"`
	Todos       []Todo `json:"todos"`
}

type DirectorySettings struct {
	ProjectId string `json:"project_id"`
}

func GetConfig() (Config, error) {
	var config Config

	homeDir, err := os.UserHomeDir()

	if err != nil {
		return config, err
	} else {
		var path = filepath.Join(homeDir, ".config", "cli-do", "config.json")
		byteValue, _ := os.ReadFile(path)
		json.Unmarshal(byteValue, &config)

		return config, nil
	}
}

func GetAuth() (Auth, error) {
	var auth Auth

	homeDir, err := os.UserHomeDir()

	if err != nil {
		fmt.Println("Please log back in.")
		return auth, err
	}

	var path = filepath.Join(homeDir, ".config", "cli-do", "auth.json")
	byteValue, _ := os.ReadFile(path)
	json.Unmarshal(byteValue, &auth)

	return auth, nil
}

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
				Action:  HandleLogin,
			},
			{
				Name:    "list",
				Aliases: []string{"ls"},
				Usage:   "List projects and todos",
				Subcommands: []*cli.Command{
					{
						Name:     "projects",
						Category: "List",
						Aliases:  []string{"p"},
						Usage:    "List projects",
						Action:   HandleProjectList,
					},
					{
						Name:     "todos",
						Category: "List",
						Aliases:  []string{"t"},
						Usage:    "List todos",
						Action:   HandleTodosList,
					},
				},
			},
			{
				Name:    "todo",
				Aliases: []string{"t"},
				Usage:   "Todo operations",
				Subcommands: []*cli.Command{
					{
						Name:      "get",
						ArgsUsage: "<ticket>",
						Aliases:   []string{"g"},
						Action:    HandleGetTodo,
					},
					{
						Name:    "create",
						Aliases: []string{"c"},
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
						Action: HandleCreateTodo,
					},
					{
						Name:    "complete",
						Aliases: []string{"co"},
						Action:  HandleCompleteTodo,
					},
				},
			},
		},
		Action: func(*cli.Context) error {
			fmt.Println("Hello, cli-do!")
			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Println(err)
	}
}

func HandleProjectList(ctx *cli.Context) error {
	var config Config
	config, _ = GetConfig()
	var auth, _ = GetAuth()

	resp, err := resty.New().R().
		SetHeader("Authorization", fmt.Sprintf("Bearer %s", auth.AccessToken)).
		SetHeader("Content-Type", "application/json").
		Get(fmt.Sprintf("%s/projects", config.Endpoint))

	if err != nil {
		fmt.Println("Error:", err)
		return err
	}

	var projects Projects
	errParseProjects := json.Unmarshal([]byte(resp.String()), &projects)

	if errParseProjects != nil {
		fmt.Println("Error:", errParseProjects)
		return nil
	}

	var tbl = table.New("ID", "Name")

	for _, project := range projects.Projects {
		tbl.AddRow(project.Ticket, project.Name)
	}

	tbl.Print()

	return nil
}

func HandleTodosList(ctx *cli.Context) error {
	var config, _ = GetConfig()
	var auth, _ = GetAuth()
	var directorySettings = ReadDirectorySettingsFile(ctx)

	if directorySettings.ProjectId == "" {
		fmt.Println("Project ID is required or init a new project in current directory.")
		return nil
	}

	resp, err := resty.New().R().
		SetHeader("Authorization", fmt.Sprintf("Bearer %s", auth.AccessToken)).
		SetHeader("Content-Type", "application/json").
		Get(fmt.Sprintf("%s/projects/%s", config.Endpoint, directorySettings.ProjectId))

	if err != nil {
		fmt.Println("Error:", err)
		return err
	}

	if resp.StatusCode() == 404 {
		fmt.Println("Project not found. Please check the project ID.")
		return nil
	}

	var project Project
	errParseProject := json.Unmarshal([]byte(resp.String()), &project)
	if errParseProject != nil {
		fmt.Println("Error:", errParseProject)
		return errParseProject
	}

	var tbl = table.New("Ticket", "Subject", "Body", "Due Date", "Completed", "Past Due")

	for _, todo := range project.Todos {
		var dueDate string
		if todo.DueDate == nil {
			dueDate = "-"
		} else {
			dueDate = todo.DueDate.String()
		}
		var trunacatedSubject = truncate.Truncate(todo.Subject, 24, "...", truncate.PositionEnd)
		var truncatedBody = truncate.Truncate(todo.Body, 32, "...", truncate.PositionEnd)
		tbl.AddRow(todo.Ticket, trunacatedSubject, truncatedBody, dueDate, todo.Completed, todo.PastDue)
	}

	tbl.Print()

	return nil
}

func HandleGetTodo(ctx *cli.Context) error {
	var config, _ = GetConfig()
	var auth, _ = GetAuth()
	var directorySettings = ReadDirectorySettingsFile(ctx)

	resp, err := resty.New().R().
		SetHeader("Authorization", fmt.Sprintf("Bearer %s", auth.AccessToken)).
		SetHeader("Content-Type", "application/json").
		Get(fmt.Sprintf("%s/projects/%s/todos/%s", config.Endpoint, directorySettings.ProjectId, ctx.Args().First()))

	if err != nil {
		fmt.Println("Error:", err)
		return err
	}

	if resp.StatusCode() == 404 {
		fmt.Println("Todo not found. Please check the Ticket number and project ID.")
		return nil
	}

	var todo Todo

	errParseTodo := json.Unmarshal([]byte(resp.String()), &todo)

	if errParseTodo != nil {
		fmt.Println("Error:", errParseTodo)
		return errParseTodo
	}

	fmt.Println("Ticket:", todo.Ticket)
	if todo.DueDate != nil {
		fmt.Println("Due Date:", todo.DueDate)
	}
	fmt.Println("Completed:", todo.Completed)
	fmt.Println("Subject:", todo.Subject)
	fmt.Println("Body:", todo.Body)
	return nil
}

func HandleCreateTodo(ctx *cli.Context) error {
	var config, _ = GetConfig()
	var auth, _ = GetAuth()
	var directorySettings = ReadDirectorySettingsFile(ctx)

	var createTodo = CreateTodo{
		Todo: Todo{
			Subject: ctx.String("subject"),
			Body:    ctx.String("body"),
			DueDate: ctx.Timestamp("due-date"),
		},
	}

	var createTodoJson, _ = json.Marshal(createTodo)

	resp, err := resty.New().R().
		SetHeader("Authorization", fmt.Sprintf("Bearer %s", auth.AccessToken)).
		SetHeader("Content-Type", "application/json").
		SetBody(createTodoJson).
		Post(fmt.Sprintf("%s/projects/%s/todos", config.Endpoint, directorySettings.ProjectId))

	if err != nil {
		return nil
	}

	if resp.StatusCode() == 200 {
		fmt.Println("Todo created successfully!")
	}

	return nil
}

func HandleCompleteTodo(ctx *cli.Context) error {
	var config, _ = GetConfig()
	var auth, _ = GetAuth()
	var directorySettings = ReadDirectorySettingsFile(ctx)

	resp, err := resty.New().R().
		SetHeader("Authorization", fmt.Sprintf("Bearer %s", auth.AccessToken)).
		SetHeader("Content-Type", "application/json").
		Post(fmt.Sprintf("%s/projects/%s/todos/%s/complete", config.Endpoint, directorySettings.ProjectId, ctx.Args().First()))

	if err != nil {
		return nil
	}

	if resp.StatusCode() == 404 {
		fmt.Println("Todo not found. Please check the Ticket number and project ID.")
	}

	if resp.StatusCode() == 200 {
		fmt.Println("Todo completed successfully!")
	}

	return nil
}

func HandleLogin(ctx *cli.Context) error {
	var config Config
	config, _ = GetConfig()

	var homeDir, _ = os.UserHomeDir()
	var path = filepath.Join(homeDir, ".config", "cli-do", "auth.json")

	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		err := LoginUser(config)

		if err != nil {
			fmt.Println("Error:", err)
		}
	} else {
		fmt.Println("You are already logged in!")
	}

	return nil
}

func LoginUser(config Config) error {
	var auth Auth
	var email string
	var password string

	fmt.Println("Login to cli-do")
	fmt.Print("Email: ")
	fmt.Scan(&email)

	fmt.Print("Password: ")
	fmt.Scan(&password)

	fmt.Print("Logging in...")

	var endpoint = fmt.Sprintf("%s/login", config.Endpoint)

	client := resty.New()
	resp, err := client.R().
		EnableTrace().
		SetHeader("Content-Type", "application/json").
		SetBody(Login{Email: email, Password: password, ClientId: config.ClientId}).
		Post(endpoint)

	if err != nil {
		fmt.Println("Error:", err)
		return err
	}

	fmt.Printf("Response Info:\n")
	fmt.Println("Body:\n", resp)

	homeDir, err := os.UserHomeDir()
	var path = filepath.Join(homeDir, ".config", "cli-do")

	_ = os.MkdirAll(path, os.ModeDir)

	path = filepath.Join(path, "auth.json")

	saveCliDoConfigError := os.WriteFile(path, []byte(resp.String()), 0644)

	if saveCliDoConfigError != nil {
		fmt.Println("Error:", err)
	}

	fmt.Printf("Welcome to cli-do!\n")

	json.Unmarshal([]byte(resp.String()), &auth)

	return nil
}

func ReadDirectorySettingsFile(ctx *cli.Context) DirectorySettings {
	var directorySettings DirectorySettings = DirectorySettings{
		ProjectId: ctx.String("project"),
	}
	var wd, _ = os.Getwd()
	var path = filepath.Join(wd, ".cli-do-project")

	if _, err := os.Stat(path); err == nil {
		byteValue, _ := os.ReadFile(path)
		json.Unmarshal(byteValue, &directorySettings)
	}

	return directorySettings
}
