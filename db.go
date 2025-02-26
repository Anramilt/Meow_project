package main

import (
	"fmt"

	//"context"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

var db *sqlx.DB //Переменная сессии БД

const (
	host     = "localhost"
	port     = 5432
	user     = "admin"
	password = "12345678"
	dbname   = "meowdb"
)

func ConnectDB() error {
	var err error
	db, err = sqlx.Connect("postgres", fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname))
	fmt.Println("Successfully connected!")
	return err
}
