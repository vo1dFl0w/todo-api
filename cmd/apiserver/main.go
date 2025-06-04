package main

import (
	"log"

	"github.com/vo1dFl0w/taskmanager-api/internal/app/apiserver"
)

func main() {
	cfg := apiserver.NewConfig()

	if err := apiserver.Launch(cfg); err != nil {
		log.Fatalf("cannot launch the server: %s", err)
	}
}