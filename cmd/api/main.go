package main

import (
	"log"

	"github.com/muchirisworld/terminal/internal/server"
)

func main() {
	cfg := server.NewConfig()

	if err := run(cfg); err != nil {
		log.Fatal(err.Error())
	}
}

func run(cfg *server.Config) error {
	app := server.NewApplication(cfg)
	return app.Run()
}