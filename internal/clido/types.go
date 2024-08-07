package clido

import "time"

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
