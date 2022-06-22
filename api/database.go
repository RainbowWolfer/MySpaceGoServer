package api

import (
	"database/sql"
	// "fmt"
	"time"
)

func GetDatabase() (*sql.DB, error) {
	var err error
	database, err := sql.Open("mysql", "wjx:123456@tcp(www.cqtest.top:3306)/wjx")
	// println(fmt.Sprintf("Connection in use %d", database.Stats().InUse))
	// println("Open new Database Connection")
	if err != nil {
		return nil, err
	}
	database.SetConnMaxLifetime(time.Second * 2)
	// database.SetMaxOpenConns(500)
	// database.SetMaxIdleConns(500)
	return database, nil
}
