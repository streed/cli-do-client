package clido

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/urfave/cli/v2"
)

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

	defer file.Close()

	fileScanner := bufio.NewScanner(file)
	fileScanner.Split(bufio.ScanLines)
	var fileLines []string

	for fileScanner.Scan() {
		fileLines = append(fileLines, fileScanner.Text())
	}

	var updatedTodo = ParseHeaders(todo, fileLines[:4])

	var newBody = strings.Join(fileLines[4:], "\n")
	updatedTodo.Body = newBody

	return updatedTodo, nil
}

func ParseHeaders(todo Todo, fileLines []string) Todo {
	regex := regexp.MustCompile(`#\s+(Ticket|Subject|Completed|DueDate)\s*:\s+(.*)`)

	for _, line := range fileLines {
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
			if matches[2] != "none" {
				t, _ := time.Parse("2006-01-02", matches[2])
				todo.DueDate = &t
			}
		}
	}

	return todo
}

func WriteToTempFile(todo Todo) (string, error) {
	var file, err = os.CreateTemp("./", ".todo-*")
	defer file.Close()

	if err != nil {
		return "", err
	}

	var header = fmt.Sprintf("# Ticket: %d\n# Subject: %s\n# Completed: %t\n",
		todo.Ticket,
		todo.Subject,
		todo.Completed)

	if todo.DueDate != nil {
		header = fmt.Sprintf("%s# DueDate: %s\n", header, todo.DueDate.Format("2006-01-02"))
	} else {
		header = fmt.Sprintf("%s# DueDate: none\n", header)
	}

	if todo.Body == "" || todo.Body == "\n" {
		header = fmt.Sprintf("%s\n", header)
	}

	_, err = file.WriteString(header + todo.Body)

	if err != nil {
		return "", err
	}

	return file.Name(), nil
}
