package main

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"math/rand"
	"time"

	//"context"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

var db *sqlx.DB //Переменная сессии БД

const (
	host     = "localhost"
	port     = 5432
	user     = "testadmin"
	password = "12345678"
	dbname   = "godb"
)

// Структура для хранения данных из таблицы testone
type TestOne struct {
	ID  int    `db:"id"`
	One string `db:"one"`
}

// Структура для хранения эхо
type EchoMessage struct {
	ID      int    `db:"id"`
	Message string `db:"message"`
}

func ConnectDB() error {
	var err error
	db, err = sqlx.Connect("postgres", fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname))
	fmt.Println("Successfully connected!")
	return err
}

func getEchomessageList() ([]EchoMessage, error) {
	query := `SELECT id, message FROM echomessage`
	var messages []EchoMessage
	err := db.Select(&messages, query)
	if err != nil {
		return nil, fmt.Errorf("error selecting rows in echomessage: %w", err)
	}
	return messages, nil
}

func addEchotoDB(message string) error {
	query := `INSERT INTO echomessage (message) VALUES ($1)`
	_, err := db.Exec(query, message)
	if err != nil {
		return fmt.Errorf("error inserting echo: %w", err)
	}
	fmt.Println("Echo added in table: ", message)
	return nil
}

func generateSalt() string {
	rand.Seed(time.Now().UnixNano())
	salt := make([]byte, 16)
	for i := range salt {
		salt[i] = byte(rand.Intn(256))
	}
	return hex.EncodeToString(salt)
}

// Функция добавления аккаунта в базу данных
func addAccounttoDB(username, password string) error {

	salt := generateSalt()

	// Хеширование пароля с использованием SHA-256
	h := sha256.New()
	h.Write([]byte(password + salt))
	hashedPassword := hex.EncodeToString(h.Sum(nil))
	//Хеширование с bcrypt
	/*hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("error hashing password: %w", err)
	}*/
	// SQL-запрос для вставки данных
	query := "INSERT INTO account (username, password, salt) VALUES (($1), ($2), ($3))"
	_, err := db.Exec(query, username, hashedPassword, salt)
	if err != nil {
		return fmt.Errorf("error inserting account: %w", err)
	}

	fmt.Println("Account added in table: ", username)
	return nil
}

func userExist(username, password string) (bool, error) {
	var hashedPassword, salt string

	err := db.QueryRow("SELECT password, salt FROM account WHERE username = $1", username).Scan(&hashedPassword, &salt)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil // Пользователь не найден
		}
		return false, err // Ошибка при выполнении запроса
	}
	//sha256
	h := sha256.New()
	h.Write([]byte(password + salt))
	hashedInputPassword := hex.EncodeToString(h.Sum(nil))

	if hashedInputPassword != hashedPassword {
		return false, nil // Неверный пароль
	}

	//bcrypt
	//err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	/*if err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			return false, nil // Неверный пароль
		}
		return false, err // Ошибка при сравнении паролей
	}*/

	return true, nil // Успешная аутентификация

}

/*var dbPassword string
query := `SELECT password FROM account WHERE username = $1`
err := db.QueryRow(query, username).Scan(&dbPassword)
if err != nil {
	if err == sql.ErrNoRows {
		return false, nil //Пользователь не найден
	}
	return false, nil //Ошибка при выполнении запроса
}
return password == dbPassword, nil //Сравнение паролей*/

/*func getTestoneList() ([]TestOne, error) {
	query := `SELECT id, one FROM testone`
	var rows []TestOne
	err := db.Select(&rows, query)
	if err != nil {
		return nil, fmt.Errorf("error selecting rows in testone: %w", err)
	}
	return rows, nil
}*/

/*
func addRowTestone(id int, one string) error {
	query := `INSERT INTO testone (id, one) VALUES ($1, $2)`
	_, err := db.Exec(query, id, one)
	if err != nil {
		return fmt.Errorf("error inserting row: %w", err)
	}
	return nil
}*/

/*
func getTable() ([]string, error) {
	query := `SELECT tablename FROM pg_catalog.pg_tables WHERE table_schema = 'public';`
	var tables []string
	err := db.Select(&tables, query)
	if err != nil {
		log.Fatalf("Error selected tables list: %v", err)
	}
	return tables, nil
}*/

/*func GetDB() *sqlx.DB {
	return db
}*/
