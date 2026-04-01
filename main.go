package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"
	"github.com/aaronbolcerek/BlogAggregator/internal/config"
	"github.com/aaronbolcerek/BlogAggregator/internal/database"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

type state struct {
	db *database.Queries
	config *config.Config
}

type command struct {
	name string
	args []string
}

type commands struct {
	commands map[string]func(*state, command) error
}

func (c *commands) run(s *state, cmd command) error {
	handler, ok := c.commands[cmd.name]
	if !ok {
		return fmt.Errorf("Error, command name not found")
	} 
	return handler(s, cmd)
}

func (c *commands) register(name string, f func(*state, command) error) {
	c.commands[name] = f
}

func handlerLogin(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return fmt.Errorf("No command argument\n")
	}
	ctx := context.Background()
	_, err := s.db.GetUser(ctx, cmd.args[0])
	if err != nil {
		return fmt.Errorf("User does not exist\n")
	}
	err = s.config.SetUser(cmd.args[0]) 
	if err != nil {
		return err
	}
	fmt.Printf("User has been set")
	return nil
}

func handlerRegister(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return fmt.Errorf("No command argument\n")
	}
	ctx := context.Background()
	created_at := time.Now()
	updated_at := time.Now()
	user := database.CreateUserParams{
		ID: uuid.New(),
		CreatedAt: created_at,
		UpdatedAt: updated_at,
		Name: cmd.args[0],
	}
	_, err := s.db.CreateUser(ctx, user)
	if err != nil {
		return err
	}
	err = s.config.SetUser(cmd.args[0]) 
	if err != nil {
		return err
	}
	fmt.Printf("User has been created\n")
	ctx = context.Background()
	user_details, err := s.db.GetUser(ctx, cmd.args[0])
	if err != nil {
		return err
	}
	fmt.Print(user_details)
	return nil
}

func handlerReset(s *state, cmd command) error {
	ctx := context.Background()
	err := s.db.ResetUsers(ctx)
	if err != nil {
		return fmt.Errorf("Error when resetting database\n")
	}
	return nil
}

func handlerUsers(s *state, cmd command) error {
	ctx := context.Background()
	users, err := s.db.GetUsers(ctx)
	if err != nil {
		return err
	}
	for _, user := range users {
		if user.Name == s.config.CurrentUserName {
			fmt.Printf("* %v (current)\n", user.Name)
		} else {
			fmt.Printf("* %v\n", user.Name)
		}
	}
	return nil
}

func main() {
	original_config, err := config.Read()
	if err != nil {
		fmt.Print(err)
		os.Exit(1)
	}
	db, err := sql.Open("postgres", original_config.DbUrl)
	if err != nil {
		fmt.Print(err)
		os.Exit(1)
	}
	dbQueries := database.New(db)
	user_state := state{
		db: dbQueries,
		config: &original_config,
	}
	cmds := commands{
		commands: make(map[string]func(*state, command) error),
	}
	cmds.register("login", handlerLogin)
	cmds.register("register", handlerRegister)
	cmds.register("reset", handlerReset)
	cmds.register("users", handlerUsers)
	arguments := os.Args
	if len(arguments) < 2 {
		fmt.Printf("Please provide more than 1 argument\n")
		os.Exit(1)
	}
	cmd := command{name: arguments[1], args: arguments[2:]}
	err = cmds.run(&user_state, cmd)
	if err != nil {
		fmt.Print(err)
		os.Exit(1)
	}
}