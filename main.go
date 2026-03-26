package main

import (
	"github.com/aaronbolcerek/BlogAggregator/internal/config"
	"fmt"
)

func main() {
	original_config, err := config.Read()
	if err != nil {
		fmt.Print(err)
	}
	err = original_config.SetUser("aaron")
	if err != nil {
		fmt.Print(err)
	}
	updated_config, err := config.Read()
	if err != nil {
		fmt.Print(err)
	}
	fmt.Print(updated_config)
}