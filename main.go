package main

import (
	"github.com/aaronbolcerek/BlogAggregator/internal/config"
	"fmt"
	"os"
)

type state struct {
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
		return fmt.Errorf("No command argument")
	}
	err := s.config.SetUser(cmd.args[0]) 
	if err != nil {
		return err
	}
	fmt.Printf("User has been set")
	return nil

}

func main() {
	original_config, err := config.Read()
	if err != nil {
		fmt.Print(err)
	}
	user_state := state{
		config: &original_config,
	}
	cmds := commands{
		commands: make(map[string]func(*state, command) error),
	}
	cmds.register("login", handlerLogin)
	arguments := os.Args
	if len(arguments) < 2 {
		fmt.Printf("Please provide more than 1 argument")
		os.Exit(1)
	}
	cmd := command{name: arguments[1], args: arguments[2:]}
	err = cmds.run(&user_state, cmd)
	if err != nil {
		fmt.Print(err)
		os.Exit(1)
	}
}