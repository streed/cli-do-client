package clido

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/aquilax/truncate"
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
	var api = Api{
		config: config,
		auth:   auth,
	}
	var directorySettings = ReadDirectorySettingsFile(ctx)

	todo, err := api.GetTodo(directorySettings.ProjectId, ctx.Args().First())

	if err != nil {
		return err
	}

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

	err = api.UpdateTodo(directorySettings.ProjectId, ctx.Args().First(), updateTodoRequest)

	if err != nil {
		return nil
	}

	fmt.Println("Todo updated successfully!")

	_ = os.Remove(path)

	return nil
}

func HandleCreateTodo(ctx *cli.Context) error {
	var config, _ = GetConfig()
	var auth, _ = GetAuth()
	var api = Api{
		config: config,
		auth:   auth,
	}
	var directorySettings = ReadDirectorySettingsFile(ctx)

	var createTodo = CreateTodo{
		Todo: Todo{
			Subject: ctx.String("subject"),
			Body:    ctx.String("body"),
			DueDate: ctx.Timestamp("due-date"),
		},
	}

	todo, err := api.CreateTodo(directorySettings.ProjectId, createTodo)

	if err != nil {
		return err
	}

	fmt.Printf("Todo created successfully with ticket: %d\n", todo.Ticket)

	return nil
}

func HandleArchiveTodo(ctx *cli.Context) error {
	var config, _ = GetConfig()
	var auth, _ = GetAuth()
	var api = Api{
		config: config,
		auth:   auth,
	}
	var directorySettings = ReadDirectorySettingsFile(ctx)

	err := api.ArchiveTodo(directorySettings.ProjectId, ctx.Args().First())

	if err != nil {
		return err
	}

	fmt.Println("Todo archived successfully!")

	return nil
}

func HandleCompleteTodo(ctx *cli.Context) error {
	var config, _ = GetConfig()
	var auth, _ = GetAuth()
	var api = Api{
		config: config,
		auth:   auth,
	}
	var directorySettings = ReadDirectorySettingsFile(ctx)

	err := api.CompleteTodo(directorySettings.ProjectId, ctx.Args().First())

	if err != nil {
		return nil
	}

	fmt.Println("Todo completed successfully!")

	return nil
}
