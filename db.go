package main

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
	"math/rand"

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

func generateSalt() string {
	salt := make([]byte, 16) // 16 байт соли
	_, err := rand.Read(salt)
	if err != nil {
		log.Fatal("Ошибка генерации соли:", err)
	}
	return hex.EncodeToString(salt)
}

// Функция для проверки введенного ключа
func verifyKey(userKey string) (int, bool, error) {
	// 1. Извлекаем все записи с ключами и солями из базы данных
	rows, err := db.Query(`SELECT id_key, content_key, salt FROM key`)
	if err != nil {
		return 0, false, fmt.Errorf("ошибка при извлечении ключей из базы данных: %w", err)
	}
	defer rows.Close()
	fmt.Println("Записи считаны")
	// 2. Перебираем все записи в базе данных
	for rows.Next() {
		var keyID int
		var storedHashedKey, storedSalt string
		if err := rows.Scan(&keyID, &storedHashedKey, &storedSalt); err != nil {
			return 0, false, fmt.Errorf("ошибка при чтении данных из базы: %w", err)
		}
		fmt.Println(storedHashedKey)
		//fmt.Println(storedSalt)
		// 3. Захешируем введенный ключ с этой солью
		h := sha256.New()
		h.Write([]byte(userKey + storedSalt))
		hashedInputKey := hex.EncodeToString(h.Sum(nil))
		fmt.Println(hashedInputKey)
		// 4. Сравниваем полученный хеш с тем, что хранится в базе данных
		if storedHashedKey == hashedInputKey {
			fmt.Println("Ключ совпал блин!!1")
			return keyID, true, nil // ключ совпал
		}
	}
	// 5. Если ни один ключ не совпал, возвращаем false
	return 0, false, nil
}

// Функция добавления аккаунта в базу данных
func addAccountToDB(login, password, contentKey string) error {

	// Проверяем, есть ли уже такой логин
	var exists bool
	err := db.Get(&exists, "SELECT EXISTS(SELECT 1 FROM account WHERE login=$1)", login)
	if err != nil {
		return fmt.Errorf("ошибка проверки логина: %w", err)
	}
	if exists {
		return fmt.Errorf("логин уже существует")
	}

	// Проверяем ключ и получаем keyID
	keyID, isValid, err := verifyKey(contentKey)
	if err != nil {
		return fmt.Errorf("ошибка при проверке ключа: %w", err)
	}
	if !isValid {
		return fmt.Errorf("неверный ключ")
	}
	salt := generateSalt()
	// Хеширование пароля с использованием SHA-256
	h := sha256.New()
	h.Write([]byte(password + salt))
	hashedPassword := hex.EncodeToString(h.Sum(nil))

	// Добавляем пользователя
	query := "INSERT INTO account (login, password, salt, id_key) VALUES (($1), ($2), ($3), ($4))"
	_, err = db.Exec(query, login, hashedPassword, salt, keyID)
	if err != nil {
		return fmt.Errorf("ошибка при добавлении аккаунта: %w", err)
	}

	fmt.Println("Account added in table: ", login)
	return nil
}

// Функция для проверки введенного логина и пароля
func verifyLoginCredentials(login, password string) error {
	// Получаем соль и хеш пароля для данного логина
	var storedSalt, storedHashedPassword string
	err := db.QueryRow("SELECT salt, password FROM account WHERE login = $1", login).Scan(&storedSalt, &storedHashedPassword)
	if err == sql.ErrNoRows {
		return fmt.Errorf("пользователь с таким логином не найден")
	} else if err != nil {
		return fmt.Errorf("ошибка при извлечении данных из базы: %w", err)
	}

	// Хешируем введенный пароль с использованием той же соли
	h := sha256.New()
	h.Write([]byte(password + storedSalt))
	hashedInputPassword := hex.EncodeToString(h.Sum(nil))

	// Сравниваем полученный хеш с тем, который хранится в базе данных
	if storedHashedPassword != hashedInputPassword {
		return fmt.Errorf("неверный пароль")
	}

	return nil
}

//
//
//
//
//
//
//
//
//
//
//
//
//
//
//
//
//
//
//
//
//
//
//
//
//
//
//
//
//
//
//
//
//
//
//
/*// Функция генерации ключа
func generateKey() string {
	keyBytes := make([]byte, 16) // Генерация 16-байтового ключа
	_, err := rand.Read(keyBytes)
	if err != nil {
		log.Fatal("Ошибка генерации ключа:", err)
	}
	return hex.EncodeToString(keyBytes)
}

// Функция для добавления ключей в базу данных
func addKeysToDB(numKeys int) error {
	for i := 0; i < numKeys; i++ {
		// Генерируем новый ключ
		contentKey := generateKey()

		// Генерируем соль
		salt := generateSalt()

		// Хешируем ключ с солью
		h := sha256.New()
		h.Write([]byte(contentKey + salt))
		hashedKey := hex.EncodeToString(h.Sum(nil))

		// Генерация случайного типа доступа (можно заменить на другие значения)
		accessType := i % 3 // например, для демонстрации, случайно выбираем 0, 1 или 2

		// Добавляем ключ в таблицу key
		query := "INSERT INTO key (content_key, salt, access_type) VALUES ($1, $2, $3)"
		_, err := db.Exec(query, hashedKey, salt, accessType)
		if err != nil {
			return fmt.Errorf("ошибка при добавлении ключа: %w", err)
		}

		// Выводим информацию о сгенерированном ключе
		fmt.Printf("Ключ %d: %s (access_type=%d)\n", i+1, contentKey, accessType)
	}
	return nil
}*/
