package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/aquilax/truncate"
	"github.com/go-resty/resty/v2"
	"github.com/rodaine/table"
	"github.com/urfave/cli/v2"
	"golang.org/x/term"
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

type UpdateTodo struct {
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

type CreateProject struct {
	Project Project `json:"project"`
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
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "force",
						Usage:   "Force to refresh login if already logged in",
						Aliases: []string{"f"},
					},
				},
				Action: HandleLogin,
			},
			{
				Name:    "todo",
				Aliases: []string{"t"},
				Usage:   "Todo operations",
				Subcommands: []*cli.Command{
					{
						Name:    "list",
						Aliases: []string{"ls"},
						Action:  HandleTodosList,
					},
					{
						Name:      "get",
						ArgsUsage: "<ticket>",
						Aliases:   []string{"g"},
						Action:    HandleGetTodo,
					},
					{
						Name:    "edit",
						Aliases: []string{"e"},
						Action:  HandleEditTodo,
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
						Action: HandleCreateTodo,
					},
					{
						Name:    "archive",
						Aliases: []string{"a"},
						Action:  HandleArchiveTodo,
					},
					{
						Name:    "complete",
						Aliases: []string{"co"},
						Action:  HandleCompleteTodo,
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
						Action:  HandleInitProjectDirectory,
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
						Action: HandleProjectNew,
					},
					{
						Name:    "list",
						Aliases: []string{"ls"},
						Action:  HandleProjectList,
					},
					{
						Name:      "archive",
						Aliases:   []string{"a"},
						ArgsUsage: "<project_id>",
						Action:    HandleProjectArchive,
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

func HandleInitProjectDirectory(ctx *cli.Context) error {
	var config, _ = GetConfig()
	var auth, _ = GetAuth()

	var directorySettings = ReadDirectorySettingsFile(ctx)

	if directorySettings.ProjectId != "" {
		fmt.Println("Project directory already initialized!")
		return nil
	}

	resp, err := resty.New().R().
		SetHeader("Authorization", fmt.Sprintf("Bearer %s", auth.AccessToken)).
		SetHeader("Content-Type", "application/json").
		Get(fmt.Sprintf("%s/projects/%s", config.Endpoint, ctx.Args().First()))

	if err != nil {
		return err
	}

	if resp.StatusCode() == 404 {
		fmt.Println("Project not found. Please check the project ID.")
	}

	if resp.StatusCode() == 200 {
		var wd, _ = os.Getwd()
		var path = filepath.Join(wd, ".cli-do-project")

		_ = os.WriteFile(path, []byte(fmt.Sprintf(`{"project_id": "%s"}`, ctx.Args().First())), 0644)

		fmt.Println("Project directory initialized successfully!")
	} else {
		fmt.Println("Project not found. Please check the project ID.")
	}

	return nil
}

func HandleTodosList(ctx *cli.Context) error {
	var config, _ = GetConfig()
	var auth, _ = GetAuth()
	var directorySettings = ReadDirectorySettingsFile(ctx)

	if directorySettings.ProjectId == "" {
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
		fmt.Println("Due Date:", todo.DueDate.Format("2006-01-02"))
	}
	fmt.Println("Completed:", todo.Completed)
	fmt.Println("Subject:", todo.Subject)
	fmt.Printf("\n%s\n", todo.Body)
	return nil
}

func HandleEditTodo(ctx *cli.Context) error {
	var config, _ = GetConfig()
	var auth, _ = GetAuth()
	var directorySettings = ReadDirectorySettingsFile(ctx)

	resp, err := resty.New().R().
		SetHeader("Authorization", fmt.Sprintf("Bearer %s", auth.AccessToken)).
		SetHeader("Content-Type", "application/json").
		Get(fmt.Sprintf("%s/projects/%s/todos/%s", config.Endpoint, directorySettings.ProjectId, ctx.Args().First()))

	if err != nil {
		return err
	}

	if resp.StatusCode() == 404 {
		fmt.Println("Todo not found. Please check the Ticket number and project ID.")
		return nil
	}

	var todo Todo
	json.Unmarshal([]byte(resp.String()), &todo)

	var path string
	path, err = WriteToTempFile(todo)

	if err != nil {
		return err
	}

	editorPath := os.Getenv("EDITOR")
	if editorPath == "" {
		editorPath = "vim"
	}

	cmd := exec.Command(editorPath, path)

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return nil
	}

	updatedTodo, err := ParseTempTodoFile(todo, path)

	if err != nil {
		return err
	}

	var updateTodoRequest = UpdateTodo{}
	updateTodoRequest.Todo = updatedTodo
	var updatedTodoJson, _ = json.Marshal(updateTodoRequest)

	resp, err = resty.New().R().
		SetHeader("Authorization", fmt.Sprintf("Bearer %s", auth.AccessToken)).
		SetHeader("Content-Type", "application/json").
		SetBody(updatedTodoJson).
		Put(fmt.Sprintf("%s/projects/%s/todos/%s", config.Endpoint, directorySettings.ProjectId, ctx.Args().First()))

	if err != nil {
		return nil
	}

	if resp.StatusCode() == 200 {
		fmt.Println("Todo updated successfully!")
	} else {
		fmt.Println("Error updating todo.")
	}

	_ = os.Remove(path)

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
		return err
	}

	if resp.StatusCode() == 200 {
		fmt.Println("Todo created successfully!")
	}

	return nil
}

func HandleArchiveTodo(ctx *cli.Context) error {
	var config, _ = GetConfig()
	var auth, _ = GetAuth()
	var directorySettings = ReadDirectorySettingsFile(ctx)

	resp, err := resty.New().R().
		SetHeader("Authorization", fmt.Sprintf("Bearer %s", auth.AccessToken)).
		Delete(fmt.Sprintf("%s/projects/%s/todos/%s", config.Endpoint, directorySettings.ProjectId, ctx.Args().First()))

	if err != nil {
		return err
	}

	if resp.StatusCode() == 404 {
		fmt.Println("Todo not found. Please check the Ticket number and project ID.")
	}

	if resp.StatusCode() == 200 {
		fmt.Println("Todo archived successfully!")
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

func HandleProjectNew(ctx *cli.Context) error {
	var config, _ = GetConfig()
	var auth, _ = GetAuth()

	var createProject = CreateProject{
		Project: Project{
			Name:        ctx.String("name"),
			Description: ctx.String("description"),
		},
	}

	var createProjectJson, _ = json.Marshal(createProject)

	rest, err := resty.New().R().
		SetHeader("Authorization", fmt.Sprintf("Bearer %s", auth.AccessToken)).
		SetHeader("Content-Type", "application/json").
		SetBody(createProjectJson).
		Post(fmt.Sprintf("%s/projects", config.Endpoint))

	if err != nil {
		return err
	}

	if rest.StatusCode() == 200 {
		fmt.Println("Project created successfully!")
	}

	return nil
}

func HandleProjectArchive(ctx *cli.Context) error {
	var config, _ = GetConfig()
	var auth, _ = GetAuth()

	resp, err := resty.New().R().
		SetHeader("Authorization", fmt.Sprintf("Bearer %s", auth.AccessToken)).
		Delete(fmt.Sprintf("%s/projects/%s", config.Endpoint, ctx.Args().First()))

	if err != nil {
		return err
	}

	if resp.StatusCode() == 404 {
		fmt.Println("Project not found. Please check the project ID.")
	}

	if resp.StatusCode() == 200 {
		fmt.Println("Project archived successfully!")
	}

	return nil
}

func HandleLogin(ctx *cli.Context) error {
	var config Config
	config, _ = GetConfig()

	var homeDir, _ = os.UserHomeDir()
	var path = filepath.Join(homeDir, ".config", "cli-do", "auth.json")
	_, err := os.Stat(path)

	if errors.Is(err, os.ErrNotExist) || ctx.Bool("force") {
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
	var email string
	var password []byte

	fmt.Println("Login to cli-do")
	fmt.Print("Email: ")
	fmt.Scan(&email)

	fmt.Print("Password: ")
	password, _ = term.ReadPassword(1)

	fmt.Println("Logging in...")

	var endpoint = fmt.Sprintf("%s/login", config.Endpoint)

	client := resty.New()
	resp, err := client.R().
		EnableTrace().
		SetHeader("Content-Type", "application/json").
		SetBody(Login{Email: email, Password: string(password[:]), ClientId: config.ClientId}).
		Post(endpoint)

	if err != nil {
		return err
	}

	homeDir, _ := os.UserHomeDir()
	var path = filepath.Join(homeDir, ".config", "cli-do")

	_ = os.MkdirAll(path, os.ModeDir)

	path = filepath.Join(path, "auth.json")

	saveCliDoConfigError := os.WriteFile(path, []byte(resp.String()), 0644)

	if saveCliDoConfigError != nil {
		return nil
	}

	fmt.Println("Welcome to cli-do!")

	return nil
}

func ReadDirectorySettingsFile(ctx *cli.Context) DirectorySettings {
	var directorySettings DirectorySettings = DirectorySettings{
		ProjectId: ctx.String("project"),
	}
	var wd, _ = os.Getwd()
	var path = filepath.Join(wd, ".cli-do-project")

	if directorySettings.ProjectId != "" {
		return directorySettings
	}

	if _, err := os.Stat(path); err == nil {
		byteValue, _ := os.ReadFile(path)
		json.Unmarshal(byteValue, &directorySettings)
	} else {
		fmt.Println("Project flag not provided and project directory not initialized.")
	}

	return directorySettings
}

func ParseTempTodoFile(todo Todo, path string) (Todo, error) {
	file, err := os.Open(path)

	if err != nil {
		return todo, err
	}

	fileScanner := bufio.NewScanner(file)
	fileScanner.Split(bufio.ScanLines)
	var fileLines []string

	for fileScanner.Scan() {
		fileLines = append(fileLines, fileScanner.Text())
	}

	file.Close()

	var updatedTodo, headerEnd = ParseHeaders(todo, fileLines)

	var newBody = strings.Join(fileLines[headerEnd:], "\n")
	updatedTodo.Body = newBody

	return updatedTodo, nil
}

func ParseHeaders(todo Todo, fileLines []string) (Todo, int) {
	regex := regexp.MustCompile(`#\s+(Ticket|Subject|Completed|DueDate)\s*:\s+(.*)`)
	var lineNum = 0

	for _, line := range fileLines {
		lineNum = lineNum + 1
		if line == "" {
			break
		}

		matches := regex.FindStringSubmatch(line)

		if len(matches) <= 0 {
			continue
		}

		if matches[1] == "Subject" {
			todo.Subject = matches[2]
		}

		if matches[1] == "Completed" {
			todo.Completed = matches[2] == "true"
		}

		if matches[1] == "DueDate" {
			t, _ := time.Parse("2006-01-02", matches[2])
			todo.DueDate = &t
		}
	}

	return todo, lineNum
}

func WriteToTempFile(todo Todo) (string, error) {
	var file, err = os.CreateTemp("./", ".todo-*")
	if err != nil {
		return "", err
	}

	var header = fmt.Sprintf("# Ticket: %d\n# Subject: %s\n# Completed: %t\n", todo.Ticket, todo.Subject, todo.Completed)

	if todo.DueDate != nil {
		header = fmt.Sprintf("%s# DueDate: %s", header, todo.DueDate.Format("2006-01-02"))
	}

	header = fmt.Sprintf("%s\n\n", header)

	_, err = file.WriteString(header + todo.Body)

	if err != nil {
		return "", err
	}

	return file.Name(), nil
}
