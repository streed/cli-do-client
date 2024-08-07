package clido

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v2"
	"golang.org/x/term"
)

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

	var api = Api{}

	err := api.Login(Login{
		Email:    email,
		Password: string(password[:]),
		ClientId: config.ClientId,
	})

	if err != nil {
		return err
	}

	homeDir, _ := os.UserHomeDir()
	var path = filepath.Join(homeDir, ".config", "cli-do")

	_ = os.MkdirAll(path, os.ModeDir)

	path = filepath.Join(path, "auth.json")

	bytes, err := json.Marshal(api.auth)

	if err != nil {
		return err
	}

	saveCliDoConfigError := os.WriteFile(path, bytes, 0644)

	if saveCliDoConfigError != nil {
		return nil
	}

	fmt.Println("Welcome to cli-do!")

	return nil
}
