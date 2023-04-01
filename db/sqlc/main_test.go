package db

import (
	"database/sql"
	"log"
	"os"
	"simplebank/util"
	"testing"

	_ "github.com/lib/pq"
)

var testQueries *Queries
// var testAccount Account
var testDB *sql.DB

func TestMain(m *testing.M) {
	config, err := util.LoadConfig("../../")
	if err != nil {
		log.Fatal("cannot load config:", err)
	}

	testDB, err = sql.Open(config.DBDriver, config.DBSource)

	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}

	testQueries = New(testDB)

	// arg := CreateAccountParams{
	// 	Owner:    util.RandomOwner(),
	// 	Balance:  util.RandomMoney(),
	// 	Currency: util.RandomCurrency(),
	// }

	// testAccount, err = testQueries.CreateAccount(context.Background(), arg)

	os.Exit(m.Run())
}