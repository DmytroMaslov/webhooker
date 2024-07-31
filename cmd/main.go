package main

import (
	"log"
	"webhooker/config"

	"webhooker/cmd/app"
)

func main() {
	c, err := config.GetConfig()
	if err != nil {
		log.Fatal("failed to get config")
	}
	app := app.App{
		Config: c,
	}

	app.Run()
}
