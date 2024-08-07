package clido

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/rodaine/table"
	"github.com/urfave/cli/v2"
)

func HandleProjectList(ctx *cli.Context) error {
	var config Config
	config, _ = GetConfig()
	var auth, _ = GetAuth()
	var api = Api{
		config: config,
		auth:   auth,
	}

	var projects, err = api.GetProjects()

	if err != nil {
		return err
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

	var api = Api{
		config: config,
		auth:   auth,
	}

	var directorySettings = ReadDirectorySettingsFile(ctx)

	if directorySettings.ProjectId != "" {
		fmt.Println("Project directory already initialized!")
		return nil
	}

	project, err := api.GetProject(ctx.Args().First())

	if err != nil {
		return err
	}

	var wd, _ = os.Getwd()
	var path = filepath.Join(wd, ".cli-do-project")

	_ = os.WriteFile(path, []byte(fmt.Sprintf(`{"project_id": "%s"}`, project.Id)), 0644)

	fmt.Println("Project directory initialized successfully!")

	return nil
}

func HandleProjectNew(ctx *cli.Context) error {
	var config, _ = GetConfig()
	var auth, _ = GetAuth()
	var api = Api{
		config: config,
		auth:   auth,
	}

	var createProject = CreateProject{
		Project: Project{
			Name:        ctx.String("name"),
			Description: ctx.String("description"),
		},
	}

	var _, err = api.CreateProject(createProject)

	if err != nil {
		return err
	}

	fmt.Println("Project created successfully!")

	return nil
}

func HandleProjectArchive(ctx *cli.Context) error {
	var config, _ = GetConfig()
	var auth, _ = GetAuth()
	var api = Api{
		config: config,
		auth:   auth,
	}

	err := api.ArchiveProject(ctx.Args().First())

	if err != nil {
		return err
	}

	return nil
}
