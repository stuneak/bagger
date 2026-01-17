package main

import (
	"log"

	"github.com/stuneak/bagger/api"
	"github.com/stuneak/bagger/config"
	db "github.com/stuneak/bagger/db/sqlc"
)

func main() {
	config, err := config.LoadConfig()
	if err != nil {
		log.Fatal("cannot load config:", err)
	}

	conn, err := db.NewDB(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}

	store := db.New(conn)
	server := api.NewServer(store)

	err = server.Start(config.ServerAddress)
	if err != nil {
		log.Fatal("cannot start server:", err)
	}
}
