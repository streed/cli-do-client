package clido

import (
	"encoding/json"
	"fmt"

	"github.com/go-resty/resty/v2"
)

type ApiError struct {
	Message    string `json:"message"`
	StatusCode int    `json:"status_code"`
}

func (e *ApiError) Error() string {
	return fmt.Sprintf("Cli-do API Error: %s", e.Message)
}

type Api struct {
	auth   Auth
	config Config
}

func (api *Api) Login(login Login) error {
	var endpoint = fmt.Sprintf("%s/login", api.config.Endpoint)
	resp, err := HandlePostNoAuth(endpoint, login, "User")

	if err != nil {
		return err
	}

	if resp.StatusCode() != 200 {
		var apiError ApiError
		apiError.StatusCode = resp.StatusCode()

		if resp.StatusCode() == 401 {
			apiError.Message = "Invalid email or password."
		}

		if resp.StatusCode() == 500 {
			apiError.Message = "Internal server error."
		}

		return &apiError
	}

	var auth Auth
	err = json.Unmarshal(resp.Body(), &auth)

	if err != nil {
		return err
	}

	api.auth = auth

	return nil
}

func (api *Api) GetProjects() (Projects, error) {
	var endpoint = fmt.Sprintf("%s/projects", api.config.Endpoint)
	resp, err := HandleGetAuth(endpoint, api.auth, "Projects")

	if err != nil {
		return Projects{}, err
	}

	var projects Projects
	err = json.Unmarshal(resp.Body(), &projects)

	if err != nil {
		return Projects{}, err
	}

	return projects, nil
}

func (api *Api) GetProject(projectId string) (Project, error) {
	var endpoint = fmt.Sprintf("%s/projects/%s", api.config.Endpoint, projectId)

	resp, err := HandleGetAuth(endpoint, api.auth, "Project")

	if err != nil {
		return Project{}, err
	}

	if resp.StatusCode() != 200 {
		var apiError ApiError
		apiError.StatusCode = resp.StatusCode()

		if resp.StatusCode() == 404 {
			apiError.Message = "Project not found."
		}

		return Project{}, &apiError
	}

	var project Project
	err = json.Unmarshal(resp.Body(), &project)

	if err != nil {
		return Project{}, err
	}

	return project, nil
}

func (api *Api) CreateProject(createProject CreateProject) (Project, error) {
	var endpoint = fmt.Sprintf("%s/projects", api.config.Endpoint)
	resp, err := HandlePostAuth(endpoint, createProject, api.auth, "Project")

	if err != nil {
		return Project{}, err
	}

	var createdProject Project
	err = json.Unmarshal(resp.Body(), &createdProject)

	if err != nil {
		return Project{}, err
	}

	return createdProject, nil
}

func (api *Api) ArchiveProject(projectId string) error {
	var endpoint = fmt.Sprintf("%s/projects/%s", api.config.Endpoint, projectId)
	_, err := HandleDeleteAuth(endpoint, api.auth, "Project")

	if err != nil {
		return err
	}

	return nil
}

func (api *Api) ListTodos(projectId string, all bool) (Todos, error) {
	var endpoint = fmt.Sprintf("%s/projects/%s/todos?all=%t", api.config.Endpoint, projectId, all)
	resp, err := HandleGetAuth(endpoint, api.auth, "Todos")

	if err != nil {
		return Todos{}, err
	}

	var todos Todos
	err = json.Unmarshal(resp.Body(), &todos)

	if err != nil {
		return Todos{}, err
	}

	return todos, nil
}

func (api *Api) GetTodo(projectId string, ticket string) (Todo, error) {
	var endpoint = fmt.Sprintf("%s/projects/%s/todos/%s", api.config.Endpoint, projectId, ticket)
	resp, err := HandleGetAuth(endpoint, api.auth, "Todo")

	if err != nil {
		return Todo{}, err
	}

	var todo Todo
	err = json.Unmarshal(resp.Body(), &todo)

	if err != nil {
		return Todo{}, err
	}

	return todo, nil
}

func (api *Api) CreateTodo(projectId string, createTodo CreateTodo) (Todo, error) {
	var endpoint = fmt.Sprintf("%s/projects/%s/todos", api.config.Endpoint, projectId)
	resp, err := HandlePostAuth(endpoint, createTodo, api.auth, "Todo")

	if err != nil {
		return Todo{}, err
	}

	var createdTodo Todo
	err = json.Unmarshal(resp.Body(), &createdTodo)

	if err != nil {
		return Todo{}, err
	}

	return createdTodo, nil
}

func (api *Api) UpdateTodo(projectId string, ticket string, updateTodo UpdateTodo) error {
	var endpoint = fmt.Sprintf("%s/projects/%s/todos/%s", api.config.Endpoint, projectId, ticket)
	_, err := HandlePutAuth(endpoint, updateTodo, api.auth, "Todo")

	if err != nil {
		return err
	}

	return nil
}

func (api *Api) ArchiveTodo(projectId string, ticket string) error {
	var endpoint = fmt.Sprintf("%s/projects/%s/todos/%s", api.config.Endpoint, projectId, ticket)
	_, err := HandleDeleteAuth(endpoint, api.auth, "Todo")

	if err != nil {
		return err
	}

	return nil
}

func (api *Api) CompleteTodo(projectId string, ticket string) error {
	var endpoint = fmt.Sprintf("%s/projects/%s/todos/%s/complete", api.config.Endpoint, projectId, ticket)
	_, err := HandlePostAuth(endpoint, nil, api.auth, "Todo")

	if err != nil {
		return err
	}

	return nil
}

func HandleResponseNotOk(resp *resty.Response, entity string) error {
	var apiError ApiError
	apiError.StatusCode = resp.StatusCode()

	if resp.StatusCode() == 200 {
		return nil
	}

	if resp.StatusCode() == 404 {
		apiError.Message = fmt.Sprintf("%s not found. Please check the %s's ID.", entity, entity)
	}

	if resp.StatusCode() == 401 {
		apiError.Message = "Invalid email or password."
	}

	if resp.StatusCode() == 500 {
		apiError.Message = "Internal server error."
	}

	return &apiError
}

func HandlePostAuth(endpoint string, body interface{}, auth Auth, entity string) (*resty.Response, error) {
	client := resty.New()
	resp, err := client.R().
		SetHeader("Authorization", fmt.Sprintf("Bearer %s", auth.AccessToken)).
		SetHeader("Content-Type", "application/json").
		SetBody(body).
		Post(endpoint)

	if err != nil {
		return resp, err
	}

	return resp, HandleResponseNotOk(resp, entity)
}

func HandlePutAuth(endpoint string, body interface{}, auth Auth, entity string) (*resty.Response, error) {
	client := resty.New()
	resp, err := client.R().
		SetHeader("Authorization", fmt.Sprintf("Bearer %s", auth.AccessToken)).
		SetHeader("Content-Type", "application/json").
		SetBody(body).
		Put(endpoint)

	if err != nil {
		return resp, err
	}

	return resp, HandleResponseNotOk(resp, entity)
}

func HandlePostNoAuth(endpoint string, body interface{}, entity string) (*resty.Response, error) {
	client := resty.New()
	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(body).
		Post(endpoint)

	if err != nil {
		return resp, err
	}

	return resp, HandleResponseNotOk(resp, entity)
}

func HandleGetAuth(endpoint string, auth Auth, entity string) (*resty.Response, error) {
	client := resty.New()
	resp, err := client.R().
		SetHeader("Authorization", fmt.Sprintf("Bearer %s", auth.AccessToken)).
		SetHeader("Content-Type", "application/json").
		Get(endpoint)

	if err != nil {
		return resp, err
	}

	return resp, HandleResponseNotOk(resp, entity)
}

func HandleDeleteAuth(endpoint string, auth Auth, entity string) (*resty.Response, error) {
	client := resty.New()
	resp, err := client.R().
		SetHeader("Authorization", fmt.Sprintf("Bearer %s", auth.AccessToken)).
		Delete(endpoint)

	if err != nil {
		return resp, err
	}

	return resp, HandleResponseNotOk(resp, entity)
}
