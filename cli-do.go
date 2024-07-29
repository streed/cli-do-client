package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/go-resty/resty/v2"
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
	Id        int    `json:"id"`
	Subject   string `json:"name"`
	Body      string `json:"body"`
	DueDate   string `json:"due_date"`
	Completed bool   `json:"completed"`
	PastDue   bool   `json:"past_due"`
}

type Projects struct {
	Projects []Project `json:"projects"`
}

type Project struct {
	Id          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Todos       []Todo `json:"todos"`
}

func GetConfig() Config {
	var config Config

	byteValue, _ := os.ReadFile("config.json")
	json.Unmarshal(byteValue, &config)

	return config
}

func GetAuth() Auth {
	var auth Auth

	byteValue, _ := os.ReadFile(".cli-do")
	json.Unmarshal(byteValue, &auth)

	return auth
}

func main() {
	app := &cli.App{
		Name:  "cli-do",
		Usage: "Interact with cli-do API",
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
				Usage:   "List objects",
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
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:    "project",
								Aliases: []string{"p"},
								Usage:   "Project name",
							},
						},
						Action: HandleTodosList,
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
	var config Config = GetConfig()
	var auth Auth = GetAuth()

	resp, err := resty.New().R().
		SetHeader("Authorization", fmt.Sprintf("Bearer %s", auth.AccessToken)).
		SetHeader("Content-Type", "application/json").
		Get(fmt.Sprintf("%s/projects", config.Endpoint))

	if err != nil {
		fmt.Println("Error:", err)
		return err
	}

	fmt.Println(resp)

	var projects Projects
	errParseProjects := json.Unmarshal([]byte(resp.String()), &projects)

	if errParseProjects != nil {
		fmt.Println("Error:", errParseProjects)
		return nil
	}

	fmt.Println("Projects:")
	for _, project := range projects.Projects {
		fmt.Println(project.Name)
	}

	return nil
}

func HandleTodosList(ctx *cli.Context) error {
	var projectId = ctx.String("project")
	fmt.Println("Project ID:", projectId)
	return nil
}

func HandleLogin(ctx *cli.Context) error {
	var config = GetConfig()

	if _, err := os.Stat("./.cli-do"); errors.Is(err, os.ErrNotExist) {
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

	saveCliDoConfigError := os.WriteFile("./.cli-do", []byte(resp.String()), 0644)

	if saveCliDoConfigError != nil {
		fmt.Println("Error:", err)
	}

	fmt.Printf("Welcome to cli-do!\n")

	json.Unmarshal([]byte(resp.String()), &auth)

	return nil
}
