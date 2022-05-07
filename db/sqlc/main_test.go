package db

import (
	"database/sql"
	"log"
	"os"
	"testing"

	"github.com/kokoichi206/account-book-api/util"
)

// テスト用のクエリ。
var testQueries *Queries

// テスト用のDB。
var testDB *sql.DB

// 全てのテストケースの前後で実行したい処理。
// db パッケージ内でのみ有効。
func TestMain(m *testing.M) {
	config, err := util.LoadConfig("../..")
	if err != nil {
		log.Fatal("cannot load config:", err)
	}

	testDB, err = sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatal("Cannot connect to db: ", err)
	}

	testQueries = New(testDB)

	os.Exit(m.Run())
}
