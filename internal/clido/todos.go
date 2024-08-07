package clido

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	"github.com/aquilax/truncate"
	"github.com/go-resty/resty/v2"
	"github.com/rodaine/table"
	"github.com/urfave/cli/v2"
)

func HandleTodosList(ctx *cli.Context) error {
	var config, _ = GetConfig()
	var auth, _ = GetAuth()
	var api = Api{
		config: config,
		auth:   auth,
	}
	var directorySettings = ReadDirectorySettingsFile(ctx)

	if directorySettings.ProjectId == "" {
		return nil
	}

	var all = ctx.Bool("all")

	todos, err := api.ListTodos(directorySettings.ProjectId, all)

	if err != nil {
		return err
	}

	var tbl = table.New("Ticket", "Subject", "Body", "Due Date", "Completed", "Past Due")

	for _, todo := range todos.Todos {
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
	var api = Api{
		config: config,
		auth:   auth,
	}
	var directorySettings = ReadDirectorySettingsFile(ctx)

	todo, err := api.GetTodo(directorySettings.ProjectId, ctx.Args().First())

	if err != nil {
		return err
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
