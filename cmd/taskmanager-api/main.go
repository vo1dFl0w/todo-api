package main

import (
	"github.com/vo1dFl0w/taskmanager-api/internal/app/apiserver"
	"github.com/vo1dFl0w/taskmanager-api/internal/app/apiserver/config"
	"github.com/vo1dFl0w/taskmanager-api/internal/app/services/logger"
)

func main() {
	cfg := config.InitConfig()

	log := logger.InitLogger(cfg.Env)

	if err := apiserver.Run(cfg, log); err != nil {
		log.Error("failed to start server")
	}
}