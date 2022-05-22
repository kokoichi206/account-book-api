package main

import (
	"database/sql"
	"log"

	"github.com/kokoichi206/account-book-api/api"
	"github.com/kokoichi206/account-book-api/auth"
	db "github.com/kokoichi206/account-book-api/db/sqlc"
	"github.com/kokoichi206/account-book-api/util"
	"go.uber.org/zap"

	_ "github.com/lib/pq"
)

func main() {
	config, err := util.LoadConfig(".")
	if err != nil {
		log.Fatal("cannot load config:", err)
	}

	conn, err := sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatal("Cannot connect to db: ", err)
	}

	querier := db.New(conn)
	manager := auth.NewManager(querier)

	logger := util.InitLogger()

	server := api.NewServer(config, querier, manager, logger)

	err = server.Start(config.ServerAddress)
	if err != nil {
		log.Fatal("cannot start server:", err)
	}
}
