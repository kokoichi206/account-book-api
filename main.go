package main

import (
	"database/sql"
	"log"

	"github.com/kokoichi206/account-book-api/api"
	db "github.com/kokoichi206/account-book-api/db/sqlc"
	"github.com/kokoichi206/account-book-api/util"
)

func main() {
	config, err := util.LoadConfig("e")
	if err != nil {
		log.Fatal("cannot load config:", err)
	}

	conn, err := sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatal("Cannot connect to db: ", err)
	}

	querier := db.New(conn)
	server := api.NewServer(config, querier)

	err = server.Start(config.ServerAddress)
	if err != nil {
		log.Fatal("cannot start server:", err)
	}
}
