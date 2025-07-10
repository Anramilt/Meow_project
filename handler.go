package main

import (
	"bytes"
	"crypto/sha256"
	"crypto/tls"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"math/rand"
	"net"
	"net/http"
	"net/mail"
	"net/smtp"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// const imageDir = "/home/sofia/Документы/Menu" // путь к корневой папке
const imageDir = "/home/sofia/Test"

type ErrorResponse struct {
	Error string
}

// Обработчик для главной страницы
/*func indexHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}

type Message struct {
	Text string `json:"text"`
}*/

func handleCors(w http.ResponseWriter, r *http.Request) bool {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Methods", "GET,HEAD,OPTIONS,POST,PUT")
	w.Header().Set("Access-Control-Allow-Headers", "Authorization, Origin, Accept, X-Requested-With, Content-Type, Access-Control-Request-Method, Access-Control-Request-Headers")
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return true
	}
	return false
}

// Обработчик для страницы регистрации
/*func registerHandler(w http.ResponseWriter, r *http.Request) {
	if handleCors(w, r) {
		return
	}
	tmpl, err := template.ParseFiles("templates/register.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}*/

// Обработчик для отображения формы авторизации
/*func handleLoginPage(w http.ResponseWriter, r *http.Request) {
	if handleCors(w, r) {
		return
	}
	tmpl, err := template.ParseFiles("templates/login.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}*/

// Обработчик поиска
func searchHandler(w http.ResponseWriter, r *http.Request) {
	if handleCors(w, r) {
		return
	}

	query := "%" + r.URL.Query().Get("q") + "%"
	if query == "" {
		http.Error(w, "Query parameter is required", http.StatusBadRequest)
		return
	}

	// Поиск игр и их изображений
	rows, err := db.Query(`
		SELECT g.id_game, g.name_game, g.type, g.icon, i.image_name
		FROM games g
		LEFT JOIN images i ON g.id_game = i.id_game
		WHERE g.name_game ILIKE $1`, "%"+query+"%")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Структура ответа
	type GameWithImages struct {
		ID     int      `json:"id"`
		Name   string   `json:"name"`
		Type   string   `json:"type"`
		Icon   string   `json:"icon"`
		Images []string `json:"images"`
	}

	gameMap := make(map[string]*GameWithImages)

	// Обрабатываем результаты запроса
	for rows.Next() {
		var id int
		var name, gameType, icon, imagePath string
		if err := rows.Scan(&id, &name, &gameType, &icon, &imagePath); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Если игра уже есть в map, просто добавляем изображение
		if _, exists := gameMap[name]; exists {
			gameMap[name].Images = append(gameMap[name].Images, imagePath)
		} else {
			gameMap[name] = &GameWithImages{
				ID:     id,
				Name:   name,
				Type:   gameType,
				Icon:   icon,
				Images: []string{imagePath},
			}
		}
	}

	// Преобразуем map в JSON
	response := make([]GameWithImages, 0, len(gameMap))
	for _, game := range gameMap {
		response = append(response, *game)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Поиск с Левентшейном
func searchlimitHandler(w http.ResponseWriter, r *http.Request) {
	if handleCors(w, r) {
		return
	}

	query := "%" + r.URL.Query().Get("q") + "%"
	if query == "" {
		http.Error(w, "Query parameter is required", http.StatusBadRequest)
		return
	}

	var games []string
	sqlQuery := `
        SELECT DISTINCT name_game
        FROM games
        WHERE name_game ILIKE $1
        LIMIT 6;`
	err := db.Select(&games, sqlQuery, "%"+query+"%")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Если найдено меньше 2-х подсказок, добавляем результаты по Левенштейну
	if len(games) < 2 {
		levenshteinQuery := `
            SELECT DISTINCT name_game
            FROM games
            WHERE levenshtein(name_game, $1) <= 9
            LIMIT 6;`
		var levenshteinGames []string
		err = db.Select(&levenshteinGames, levenshteinQuery, query)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Объединяем результаты, удаляем дубликаты
		games = append(games, levenshteinGames...)
		games = unique(games)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(games)
}

// Функция для удаления дубликатов
func unique(strings []string) []string {
	keys := make(map[string]bool)
	var list []string
	for _, entry := range strings {
		if _, exists := keys[entry]; !exists {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

/*
func imagesearchHandler(w http.ResponseWriter, r *http.Request) {
	if handleCors(w, r) {
		return
	}

	imageName := r.URL.Query().Get("name")
	if imageName == "" {
		http.Error(w, "Параметр 'name' обязателен", http.StatusBadRequest)
		return
	}

	// Корень
	//baseDir := "/home/sofia/Документы/Menu"
	baseDir := "/home/sofia/Test"

	// Поиск файла везде
	var foundPath string
	err := filepath.Walk(baseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.Contains(info.Name(), imageName) {
			foundPath = path
		}
		return nil
	})

	if err != nil {
		http.Error(w, "Ошибка при поиске файла", http.StatusInternalServerError)
		return
	}

	if foundPath == "" {
		http.Error(w, "Файл не найден", http.StatusNotFound)
		return
	}

	// Файл пользователю ->
	http.ServeFile(w, r, foundPath)
}*/

func imagesearchHandler(w http.ResponseWriter, r *http.Request) {
	if handleCors(w, r) {
		return
	}

	baseDir := "/home/sofia/Test"

	// 1. Если передан полный путь
	imagePath := r.URL.Query().Get("path")
	if imagePath != "" {
		fullPath := filepath.Join(baseDir, imagePath)
		if _, err := os.Stat(fullPath); err == nil {
			http.ServeFile(w, r, fullPath)
			return
		}
		http.Error(w, "Файл не найден (по path)", http.StatusNotFound)
		return
	}

	// 2. Если передано просто имя
	imageName := r.URL.Query().Get("name")
	if imageName != "" {
		var found string
		_ = filepath.Walk(baseDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() && info.Name() == imageName {
				found = path
				return io.EOF // останавливаем Walk
			}
			return nil
		})
		if found != "" {
			http.ServeFile(w, r, found)
			return
		}
		http.Error(w, "Файл не найден (по name)", http.StatusNotFound)
		return
	}

	// Если ни path, ни name не указаны
	http.Error(w, "Параметр 'path' или 'name' обязателен", http.StatusBadRequest)
}

func soundsearchHandler(w http.ResponseWriter, r *http.Request) {
	if handleCors(w, r) {
		return
	}

	soundName := r.URL.Query().Get("name")
	if soundName == "" {
		http.Error(w, "Параметр 'name' обязателен", http.StatusBadRequest)
		return
	}

	// Корень
	//baseDir := "/home/sofia/Документы/Menu"
	baseDir := "/home/sofia/Test"

	// Поиск файла везде
	var foundPath string
	err := filepath.Walk(baseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.Contains(info.Name(), soundName) {
			foundPath = path
		}
		return nil
	})

	if err != nil {
		http.Error(w, "Ошибка при поиске файла", http.StatusInternalServerError)
		return
	}

	if foundPath == "" {
		http.Error(w, "Файл не найден", http.StatusNotFound)
		return
	}

	// Файл пользователю ->
	http.ServeFile(w, r, foundPath)
}

// Обработчик для регистрации нового пользователя
func handleRegister(w http.ResponseWriter, r *http.Request) {
	if handleCors(w, r) {
		return
	}
	if r.Method == "POST" {
		login := r.FormValue("login")
		password := r.FormValue("password")
		contentKey := r.FormValue("key")

		fmt.Println(contentKey)
		// Если ключ правильный, добавляем аккаунт в базу данных
		err := addAccountToDB(login, password, contentKey)
		if err != nil {
			fmt.Println("Варнинг3")
			http.Error(w, "Ошибка регистрации: "+err.Error(), http.StatusInternalServerError)
			return
		}
		// Выводим сообщение, что пользователь зарегистрирован
		tmpl, err := template.ParseFiles("templates/message.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		message := "Пользователь зарегистрирован"
		tmpl.Execute(w, message)

	}
}

// Обработчик для авторизации пользователя
func handleLogin(w http.ResponseWriter, r *http.Request) {
	if handleCors(w, r) {
		return
	}
	if r.Method != "POST" {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	login := r.FormValue("login")
	password := r.FormValue("password")

	var id int
	var idKey sql.NullInt64
	err := db.QueryRow(`SELECT id_account, id_key FROM account WHERE login = $1`, login).Scan(&id, &idKey)

	if err != nil {
		http.Error(w, "Ошибка авторизации: ", http.StatusUnauthorized)
		return
	}

	// Существует ли пользователь и правильный ли пароль
	err = verifyLoginCredentials(login, password)
	if err != nil {
		http.Error(w, "Ошибка авторизации: "+err.Error(), http.StatusUnauthorized)
		return
	}
	userData, err := getFullUserProfile(id, login, idKey)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(userData)

	/*var firstName, lastName, email, subStatus sql.NullString
	err = db.QueryRow(`
		SELECT first_name, last_name, email, subscription_status
		FROM user_profile WHERE id_account = $1`, id).Scan(&firstName, &lastName, &email, &subStatus)

	if err == sql.ErrNoRows {
		_, err = db.Exec("INSERT INTO user_profile (id_account, subscription_status) VALUES ($1, 'inactive')", id)
		if err != nil {
			http.Error(w, "Ошибка создания профиля", http.StatusInternalServerError)
			return
		}
		firstName, lastName, email, subStatus = sql.NullString{}, sql.NullString{}, sql.NullString{}, sql.NullString{String: "inactive", Valid: true}
	} else if err != nil {
		http.Error(w, "Ошибка при получении данных пользователя: "+err.Error(), http.StatusInternalServerError)
		return
	}

	var accessType *int
	if idKey.Valid {
		var at int
		err = db.QueryRow(`SELECT access_type FROM key WHERE id_key = $1`, idKey.Int64).Scan(&at)
		if err == nil {
			accessType = &at

			// Обновление на "active"
			_, _ = db.Exec(`UPDATE user_profile SET subscription_status = 'active' WHERE id_account = $1`, id)
			subStatus = sql.NullString{String: "active", Valid: true}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id_account":          id,
		"login":               login,
		"first_name":          firstName.String,
		"last_name":           lastName.String,
		"email":               email.String,
		"subscription_status": subStatus.String,
		"access_type":         accessType,
	})*/
}

// Заполнение профиля
func updateProfile(w http.ResponseWriter, r *http.Request) {
	if handleCors(w, r) {
		return
	}
	if r.Method != "POST" {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	var data struct {
		ID        int    `json:"id_account"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		//Email     string `json:"email"`
	}

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Некорректный JSON", http.StatusBadRequest)
		return
	}

	_, err := db.Exec(`UPDATE user_profile SET first_name = $1, last_name = $2 WHERE id_account = $3`,
		data.FirstName, data.LastName, data.ID)
	/*_, err := db.Exec(`UPDATE user_profile SET first_name = $1, last_name = $2, email = $3 WHERE id_account = $4`,
	data.FirstName, data.LastName, data.Email, data.ID)*/
	if err != nil {
		http.Error(w, "Ошибка обновления профиля", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// Возвращает полные данные профиля по ID
func getFullUserProfile(id int, login string, idKey sql.NullInt64) (map[string]interface{}, error) {

	var firstName, lastName, email, subStatus sql.NullString
	err := db.QueryRow(`
		SELECT first_name, last_name, email, subscription_status 
		FROM user_profile WHERE id_account = $1`, id).Scan(&firstName, &lastName, &email, &subStatus)

	if err == sql.ErrNoRows {
		_, err = db.Exec(`INSERT INTO user_profile (id_account, subscription_status) VALUES ($1, 'inactive')`, id)
		if err != nil {
			return nil, fmt.Errorf("ошибка создания профиля: %w", err)
		}
		firstName, lastName, email, subStatus = sql.NullString{}, sql.NullString{}, sql.NullString{}, sql.NullString{String: "inactive", Valid: true}
	} else if err != nil {
		return nil, fmt.Errorf("ошибка при получении данных пользователя: %w", err)
	}

	var accessType *int
	if idKey.Valid {
		var at int
		err = db.QueryRow(`SELECT access_type FROM key WHERE id_key = $1`, idKey.Int64).Scan(&at)
		if err == nil {
			accessType = &at
			_, _ = db.Exec(`UPDATE user_profile SET subscription_status = 'active' WHERE id_account = $1`, id)
			subStatus = sql.NullString{String: "active", Valid: true}
		}
	}

	return map[string]interface{}{
		"id_account":          id,
		"login":               login,
		"first_name":          firstName.String,
		"last_name":           lastName.String,
		"email":               email.String,
		"subscription_status": subStatus.String,
		"access_type":         accessType,
	}, nil
}

func handleGetProfile(w http.ResponseWriter, r *http.Request) {
	if handleCors(w, r) {
		return
	}
	if r.Method != "POST" {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		ID    int    `json:"id_account"`
		Login string `json:"login"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Некорректный JSON", http.StatusBadRequest)
		return
	}

	var idKey sql.NullInt64
	err := db.QueryRow(`SELECT id_key FROM account WHERE id_account = $1`, req.ID).Scan(&idKey)
	if err != nil {
		http.Error(w, "Пользователь не найден", http.StatusNotFound)
		return
	}

	userData, err := getFullUserProfile(req.ID, req.Login, idKey)
	if err != nil {
		http.Error(w, "Ошибка получения профиля: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(userData)
}

// Обработчик для получения данных пользователя
/*func handleGetUser(w http.ResponseWriter, r *http.Request) {
	if handleCors(w, r) {
		return
	}
	if r.Method != "GET" {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}
	id := r.URL.Query().Get("id")

	var firstName, lastName, email, subStatus sql.NullString
	err := db.QueryRow(`
		SELECT first_name, last_name, email, subscription_status
		FROM user_profile WHERE id_account = $1`, id).Scan(&firstName, &lastName, &email, &subStatus)
	if err != nil {
		http.Error(w, "Ошибка при получении данных пользователя", http.StatusInternalServerError)
		return
	}


}*/

// Обработчик изменения пароля
func handleChangePassword(w http.ResponseWriter, r *http.Request) {
	if handleCors(w, r) {
		return
	}
	if r.Method != "POST" {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	var data struct {
		ID          int    `json:"id_account"`
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Некорректный JSON", http.StatusBadRequest)
		return
	}

	var login, hashedPassword, salt string
	err := db.QueryRow(`SELECT login, password, salt FROM account WHERE id_account = $1`, data.ID).
		Scan(&login, &hashedPassword, &salt)
	if err != nil {
		http.Error(w, "Ошибка при получении данных пользователя", http.StatusInternalServerError)
		return
	}

	fmt.Println("Пароль старый:", data.OldPassword)
	fmt.Println("Пароль новый:", data.NewPassword)

	oldHash := hashPassword(data.OldPassword, salt)
	fmt.Println("Хеш из базы:", hashedPassword)
	fmt.Println("Введённый хеш:", oldHash)

	if oldHash != hashedPassword {
		http.Error(w, "Неверный старый пароль", http.StatusUnauthorized)
		return
	}

	newSalt := generateSalt()
	newHash := hashPassword(data.NewPassword, newSalt)

	_, err = db.Exec(`UPDATE account SET password = $1, salt = $2 WHERE id_account = $3`,
		newHash, newSalt, data.ID)
	if err != nil {
		http.Error(w, "Ошибка при обновлении пароля", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
func hashPassword(password, salt string) string {
	h := sha256.New()
	h.Write([]byte(password + salt))
	return hex.EncodeToString(h.Sum(nil))
}

// Обработчик отправка ссылки и запись токена
func handleSendEmail(w http.ResponseWriter, r *http.Request) {
	if handleCors(w, r) {
		return
	}
	if r.Method != "POST" {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	var data struct {
		ID    int    `json:"id_account"`
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Некорректный JSON", http.StatusBadRequest)
		return
	}

	var oldEmail sql.NullString
	err := db.QueryRow(`SELECT email FROM user_profile WHERE id_account = $1`, data.ID).Scan(&oldEmail)
	if err != nil {
		http.Error(w, "Ошибка при получении текущей почты", http.StatusInternalServerError)
		return
	}

	var toSend string
	if oldEmail.Valid && oldEmail.String != "" {
		toSend = oldEmail.String
	} else {
		toSend = data.Email
	}

	token := generateToken()
	_, err = db.Exec(`UPDATE user_profile SET email_confirm_token = $1 WHERE id_account = $2`, token, data.ID)
	if err != nil {
		http.Error(w, "Ошибка при обновлении токена", http.StatusInternalServerError)
		return
	}

	link := fmt.Sprintf("http://localhost:8080/confirm-email?token=%s&email=%s", token, url.QueryEscape(toSend))

	err = sendMail(toSend, "Для подтверждения почты перейдите по ссылке: "+link, "Подтверждение почты")
	if err != nil {
		http.Error(w, "Ошибка при отправке письма: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func generateToken() string {
	token := make([]byte, 16)
	_, err := rand.Read(token)
	if err != nil {
		panic(err)
	}
	return hex.EncodeToString(token)
}

type SmtpDetails struct {
	Mail     string
	Password string
	Host     string
}

var config struct {
	MailDetails SmtpDetails
}

func loadSMTPConfigFromEnv() {
	_ = godotenv.Load()

	config.MailDetails = SmtpDetails{
		Mail:     os.Getenv("SMTP_MAIL"),
		Password: os.Getenv("SMTP_PASSWORD"),
		Host:     os.Getenv("SMTP_HOST"),
	}
}

func sendMail(recipient string, content string, subject string) error {
	from := mail.Address{Address: config.MailDetails.Mail}
	to := mail.Address{Address: recipient}
	host, _, _ := net.SplitHostPort(config.MailDetails.Host)
	auth := smtp.PlainAuth("", config.MailDetails.Mail, config.MailDetails.Password, host)
	dial, err := tls.Dial("tcp", config.MailDetails.Host, &tls.Config{InsecureSkipVerify: true, ServerName: host})
	if err != nil {
		return err
	}
	c, err := smtp.NewClient(dial, host)
	if err != nil {
		return err
	}
	if err = c.Auth(auth); err != nil {
		return err
	}
	if err = c.Mail(from.Address); err != nil {
		return err
	}
	if err = c.Rcpt(to.Address); err != nil {
		return err
	}
	w, err := c.Data()
	if err != nil {
		return err
	}

	message := "From: " + from.Address + "\r\n"
	message += "To: " + to.Address + "\r\n"
	message += "Subject: " + subject + "\r\n"
	message += "MIME-Version: 1.0\r\n"
	message += "Content-Type: text/plain; charset=\"utf-8\"\r\n"
	message += "\r\n"

	headerBytes := []byte(message)
	_, err = w.Write(append(headerBytes, []byte(content)...))
	if err != nil {
		return err
	}
	err = w.Close()
	if err != nil {
		return err
	}
	_ = c.Quit()
	return nil
}

func handleConfirmEmail(w http.ResponseWriter, r *http.Request) {
	if handleCors(w, r) {
		return
	}
	if r.Method != "GET" {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	token := r.URL.Query().Get("token")
	email := r.URL.Query().Get("email")

	if token == "" || email == "" {
		http.Error(w, "Некорректные параметры", http.StatusBadRequest)
		return
	}

	res, err := db.Exec(`UPDATE user_profile SET email=$1, email_confirm_token = NULL WHERE email_confirm_token = $2`, email, token)
	if err != nil {
		http.Error(w, "Ошибка при обновлении токена", http.StatusInternalServerError)
		return
	}

	rows, _ := res.RowsAffected()
	if err != nil || rows == 0 {
		http.Error(w, "Неверный токен", http.StatusUnauthorized)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Email подтвержден"))
}

// Обработчик для работы с файлами (загрузка и список)
func fileHandler(w http.ResponseWriter, r *http.Request) {
	if handleCors(w, r) {
		return
	}

	switch r.Method {
	case "GET":
		// Получение списка файлов
		files, err := listFiles(imageDir)
		if err != nil {
			http.Error(w, "Ошибка при получении списка файлов: "+err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(files)

	case "POST":
		// Загрузка файла
		err := r.ParseMultipartForm(10 << 20) // 10 MB limit
		if err != nil {
			http.Error(w, "Размер файла слишком большой", http.StatusBadRequest)
			return
		}

		file, handler, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "Ошибка загрузки файла", http.StatusBadRequest)
			return
		}
		defer file.Close()

		// Проверка расширения файла
		ext := filepath.Ext(handler.Filename)
		allowedExtensions := map[string]bool{".png": true, ".jpg": true, ".jpeg": true, ".webp": true}
		if !allowedExtensions[ext] {
			http.Error(w, "Недопустимый тип файла", http.StatusBadRequest)
			return
		}

		// Создание папки, если её нет
		if _, err := os.Stat(imageDir); os.IsNotExist(err) {
			os.Mkdir(imageDir, os.ModePerm)
		}

		// Сохранение файла
		filePath := filepath.Join(imageDir, handler.Filename)
		dst, err := os.Create(filePath)
		if err != nil {
			http.Error(w, "Ошибка сохранения файла", http.StatusInternalServerError)
			return
		}
		defer dst.Close()

		if _, err := io.Copy(dst, file); err != nil {
			http.Error(w, "Ошибка копирования файла", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Файл %s успешно загружен!", handler.Filename)

	default:
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
	}
}

// Функция получения списка файлов
func listFiles(directory string) ([]string, error) {
	files := []string{}
	entries, err := os.ReadDir(directory)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			files = append(files, entry.Name())
		}
	}
	return files, nil
}

// Получение json файла игры по запросу
func gameHandler(w http.ResponseWriter, r *http.Request) {
	if handleCors(w, r) {
		return
	}

	gameName := r.URL.Query().Get("name")
	if gameName == "" {
		http.Error(w, "Параметр 'name' обязателен", http.StatusBadRequest)
		return
	}

	var jsonPath string
	err := db.QueryRow("SELECT json_path FROM games WHERE name_game = $1", gameName).Scan(&jsonPath)
	if err != nil {
		http.Error(w, "Игра не найдена", http.StatusNotFound)
		return
	}

	content, err := os.ReadFile(jsonPath)
	if err != nil {
		http.Error(w, "Не удалось прочитать JSON-файл", http.StatusInternalServerError)
		return
	}

	var gameData map[string]interface{}
	err = json.Unmarshal(content, &gameData)
	if err != nil {
		http.Error(w, "Не удалось прочитать JSON-файл", http.StatusInternalServerError)
		return
	}
	//рандом 10 записей
	pagesRaw, ok := gameData["Pages"].([]interface{})
	if ok && len(pagesRaw) > 10 {
		rand.Seed(time.Now().UnixNano())
		rand.Shuffle(len(pagesRaw), func(i, j int) {
			pagesRaw[i], pagesRaw[j] = pagesRaw[j], pagesRaw[i]
		})
		gameData["Pages"] = pagesRaw[:10]
	}

	modifiedContent, err := json.Marshal(gameData)
	if err != nil {
		http.Error(w, "Не удалось преобразовать данные в JSON", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	//w.Write(content)
	w.Write(modifiedContent)
}

// Распознавание голоса
func uploadVoiceHandler(w http.ResponseWriter, r *http.Request) {
	if handleCors(w, r) {
		return
	}

	if r.Method != "POST" {
		sendErrorResponse(w, "Только POST", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		sendErrorResponse(w, "Ошибка чтения формы", http.StatusBadRequest)
		return
	}

	file, _, err := r.FormFile("voice")
	if err != nil {
		sendErrorResponse(w, "Не удалось получить файл", http.StatusBadRequest)
		return
	}
	defer file.Close()

	audioData, err := io.ReadAll(file)
	if err != nil {
		sendErrorResponse(w, "Ошибка чтения аудио", http.StatusInternalServerError)
		return
	}
	log.Printf("Получено аудио, размер: %d байт", len(audioData))

	// Конвертируем аудио в нужный формат
	convertedData, err := convertAudioToWav(audioData)
	if err != nil {
		log.Printf("Ошибка конвертации аудио: %v", err)
		sendErrorResponse(w, "Ошибка конвертации аудио", http.StatusInternalServerError)
		return
	}
	log.Printf("Аудио конвертировано, размер: %d байт", len(convertedData))

	answersJson := r.FormValue("answers")
	//log.Println("Лог1", r)
	var correctAnswers []string
	//log.Println("Лог2", correctAnswers)
	err = json.Unmarshal([]byte(answersJson), &correctAnswers)
	if err != nil {
		//log.Println(err)
		sendErrorResponse(w, "Ошибка парсинга ответов", http.StatusBadRequest)
		return
	}
	log.Printf("Полученные ответы: %v", correctAnswers)

	result := Recognize(correctAnswers, convertedData)
	if len(result) == 0 {
		log.Println("Распознавание не удалось")
		sendErrorResponse(w, "Ничего не распознано", http.StatusOK)
		return
	}

	isCorrect := false
	userSaid := strings.TrimSpace(strings.ToLower(result[0].Text))
	for _, ans := range correctAnswers {
		if userSaid == strings.TrimSpace(strings.ToLower(ans)) {
			isCorrect = true
			break
		}
	}

	resp := map[string]interface{}{
		"correct":  isCorrect,
		"userSaid": userSaid,
		"gopScore": result[0].Value,
		"category": result[0].Category,
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Println("Ошибка отправки ответа:", err)
		sendErrorResponse(w, "Ошибка сервера", http.StatusInternalServerError)
	}
}

// Конвертация аудио с использованием временных файлов
func convertAudioToWav(inputData []byte) ([]byte, error) {
	// Создаём временный входной файл
	inputFile, err := os.CreateTemp("", "input_audio_*.wav")
	if err != nil {
		log.Printf("Ошибка создания временного входного файла: %v", err)
		return nil, err
	}
	inputPath := inputFile.Name()
	defer os.Remove(inputPath) // Удаляем файл после использования

	if _, err := inputFile.Write(inputData); err != nil {
		log.Printf("Ошибка записи во временный входной файл: %v", err)
		inputFile.Close()
		return nil, err
	}
	inputFile.Close()

	// Создаём временный выходной файл
	outputFile, err := os.CreateTemp("", "output_audio_*.wav")
	if err != nil {
		log.Printf("Ошибка создания временного выходного файла: %v", err)
		return nil, err
	}
	outputPath := outputFile.Name()
	outputFile.Close()
	defer os.Remove(outputPath) // Удаляем файл после использования

	// Выполняем конвертацию с помощью ffmpeg
	cmd := exec.Command("ffmpeg", "-y", "-i", inputPath, "-ar", "16000", "-ac", "1", "-f", "wav", outputPath)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		log.Printf("FFmpeg error: %s", stderr.String())
		return nil, err
	}

	// Читаем конвертированные данные
	convertedData, err := os.ReadFile(outputPath)
	if err != nil {
		log.Printf("Ошибка чтения конвертированного файла: %v", err)
		return nil, err
	}

	return convertedData, nil
}

// Вспомогательная функция для отправки JSON-ошибок
func sendErrorResponse(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	w.WriteHeader(statusCode)
	resp := map[string]string{"error": message}
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Println("Ошибка отправки JSON-ошибки:", err)
	}
}

func menuHandler(w http.ResponseWriter, r *http.Request) {
	if handleCors(w, r) {
		return
	}

	// Путь к New_menu.json
	menuPath := "/home/sofia/Test/New_menu.json"

	// Читаем файл
	content, err := os.ReadFile(menuPath)
	if err != nil {
		http.Error(w, "Не удалось прочитать файл меню", http.StatusInternalServerError)
		return
	}

	// Устанавливаем заголовки и отправляем JSON
	w.Header().Set("Content-Type", "application/json")
	w.Write(content)
}

/*
func menuDownloadHandler(w http.ResponseWriter, r *http.Request) {
	if handleCors(w, r) {
		return
	}

	rows, err := db.Queryx(`
			WITH RECURSIVE category_tree AS (
				SELECT id_category, tag, icon, parent_id, 1 AS level
				FROM category
				WHERE parent_id IS NULL OR parent_id = 1
				UNION ALL
				SELECT c.id_category, c.tag, c.icon, c.parent_id, ct.level + 1
				FROM category c
				JOIN category_tree ct ON c.parent_id = ct.id_category
			)
			SELECT
				ct.id_category, ct.tag, ct.icon, ct.parent_id, ct.level,
				g.id_game, g.name_game, g.type, g.icon AS game_icon, g.json_path
			FROM category_tree ct
			LEFT JOIN game_category gc ON ct.id_category = gc.id_category
			LEFT JOIN games g ON gc.id_game = g.id_game
			ORDER BY ct.level, ct.tag, g.name_game
		`)
	if err != nil {
		http.Error(w, "Ошибка запроса к БД", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type MenuItem struct {
		IDCategory int     `json:"id_category" db:"id_category"`
		Tag        string  `json:"tag" db:"tag"`
		Icon       string  `json:"icon" db:"icon"`
		ParentID   *int    `json:"parent_id" db:"parent_id"`
		Level      int     `json:"level" db:"level"`
		IDGame     *int    `json:"id_game" db:"id_game"`
		NameGame   *string `json:"name_game" db:"name_game"`
		Type       *string `json:"type" db:"type"`
		GameIcon   *string `json:"game_icon" db:"game_icon"`
		JSONPath   *string `json:"json_path" db:"json_path"`
	}

	var items []MenuItem
	for rows.Next() {
		var item MenuItem
		if err := rows.StructScan(&item); err != nil {
			http.Error(w, "Ошибка сканирования данных", http.StatusInternalServerError)
			return
		}
		items = append(items, item)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}
*/
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
//
//
////
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
/*func apiHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000") // Разрешаем запросы с React
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json")

	response := Message{Text: "Привет из Go!"}
	json.NewEncoder(w).Encode(response)
}*/

/*func handleCors(w http.ResponseWriter, r *http.Request) bool {
	(w).Header().Set("Access-Control-Allow-Origin", "*")
	if r.Method == "OPTIONS" {
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Methods", "GET,HEAD,OPTIONS,POST,PUT")
		w.Header().Set("Access-Control-Allow-Headers", "Access-Control-Allow-Headers, Authorization, Origin,Accept, X-Requested-With, Content-Type, Access-Control-Request-Method, Access-Control-Request-Headers")
		return true
	}
	return false
}*/

// Обработчик для получения конкретного изображения
/*func imageHandler(w http.ResponseWriter, r *http.Request) {
	if handleCors(w, r) {
		return
	}

	imageName := r.URL.Query().Get("name")
	if imageName == "" {
		http.Error(w, "Параметр 'name' обязателен", http.StatusBadRequest)
		return
	}

	imagePath := filepath.Join(imageDir, imageName)
	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		http.Error(w, "Файл не найден", http.StatusNotFound)
		return
	}

	http.ServeFile(w, r, imagePath)

}*/
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
/*
func searchlimitHandler(w http.ResponseWriter, r *http.Request) {
	if handleCors(w, r) {
		return
	}
	query := r.URL.Query().Get("q") + "%"
	if query == "" {
		http.Error(w, "Query parameter is required", http.StatusBadRequest)
		return
	}

	var games []string
	sqlQuery := `
	  	SELECT DISTINCT name_game
		FROM games
		WHERE name_game ILIKE $1
		LIMIT 3;`
	err := db.Select(&games, sqlQuery, query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Если найдено меньше 3-х подсказок, добавляем результаты по Левенштейну
	if len(games) < 3 {
		query_one := r.URL.Query().Get("q")
		levenshteinQuery := `
            SELECT DISTINCT name_game
            FROM games
            WHERE levenshtein(name_game, $1) <= 4
            LIMIT 3;`
		var levenshteinGames []string
		err = db.Select(&levenshteinGames, levenshteinQuery, query_one)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Объединяем результаты, удаляем дубликаты
		games = append(games, levenshteinGames...)
		games = unique(games)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(games)
}

// Функция для удаления дубликатов
func unique(strings []string) []string {
	keys := make(map[string]bool)
	var list []string
	for _, entry := range strings {
		if _, exists := keys[entry]; !exists {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}*/

/*
// Обработчик поиска
func searchHandler(w http.ResponseWriter, r *http.Request) {
	if handleCors(w, r) {
		return
	}
	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, "Query parameter is required", http.StatusBadRequest)
		return
	}

	var categories []string
	sqlQuery := `
	  	SELECT DISTINCT category.tag
		FROM category
		JOIN accordance_game_category ON category.id_category = accordance_game_category.id_category
		JOIN game ON accordance_game_category.id_game = game.id_game
		WHERE game.name_game ILIKE $1;`
	err := db.Select(&categories, sqlQuery, query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl, err := template.ParseFiles("templates/search.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl.Execute(w, categories)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(categories)
}
*/

// Обработчик поиска с лимитом подсказок 3
/*
func searchlimitHandler(w http.ResponseWriter, r *http.Request) {
	if handleCors(w, r) {
		return
	}
	query := r.URL.Query().Get("q") + "%"
	if query == "" {
		http.Error(w, "Query parameter is required", http.StatusBadRequest)
		return
	}

	var categories []string
	sqlQuery := `
	  	SELECT DISTINCT category.tag
		FROM category
		JOIN accordance_game_category ON category.id_category = accordance_game_category.id_category
		JOIN game ON accordance_game_category.id_game = game.id_game
		WHERE game.name_game ILIKE $1
		LIMIT 3;`
	err := db.Select(&categories, sqlQuery, query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Если найдено меньше 3-х подсказок, выполняем второй запрос с использованием расстояния Левенштейна
	if len(categories) < 3 {
		query_one := r.URL.Query().Get("q")
		levenshteinQuery := `
            SELECT DISTINCT category.tag
            FROM category
            JOIN accordance_game_category ON category.id_category = accordance_game_category.id_category
            JOIN game ON accordance_game_category.id_game = game.id_game
            WHERE levenshtein(game.name_game, $1) <= 4
            LIMIT 3;`

		var levenshteinCategories []string
		err = db.Select(&levenshteinCategories, levenshteinQuery, query_one)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Объединяем результаты
		categories = append(categories, levenshteinCategories...)
		// Удаляем дубликаты
		categories = unique(categories)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(categories)
}

// Функция для удаления дубликатов из среза строк
func unique(strings []string) []string {
	keys := make(map[string]bool)
	var list []string
	for _, entry := range strings {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}
*/
