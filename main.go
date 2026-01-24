package main

import (
	"github.com/stuneak/sopeko/api"
	"github.com/stuneak/sopeko/config"
	"github.com/stuneak/sopeko/cron"
	db "github.com/stuneak/sopeko/db/sqlc"
	"github.com/stuneak/sopeko/pkg/logger"
)

var fatal = logger.NewFatalLogger("MAIN")

func main() {
	config, err := config.LoadConfig()
	if err != nil {
		fatal("cannot load config: %v", err)
	}

	conn, err := db.NewDB(config.DBDriver, config.DBSource)
	if err != nil {
		fatal("cannot connect to db: %v", err)
	}
	defer conn.Close()

	store := db.New(conn)

	// Initialize and start cron scheduler
	scheduler, err := cron.NewScheduler(store)
	if err != nil {
		fatal("cannot create scheduler: %v", err)
	}

	err = scheduler.RegisterJobs()
	if err != nil {
		fatal("cannot register cron jobs: %v", err)
	}

	scheduler.Start()
	defer scheduler.Stop()

	server := api.NewServer(store, config.GINMode)

	err = server.Start(config.ServerAddress)
	if err != nil {
		fatal("cannot start server: %v", err)
	}
}
