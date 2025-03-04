package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
)

type ErrorResponse struct {
	Error string
}

// Обработчик для главной страницы
func indexHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}

type Message struct {
	Text string `json:"text"`
}

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

func handleCors(w http.ResponseWriter, r *http.Request) bool {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Methods", "GET,HEAD,OPTIONS,POST,PUT")
	w.Header().Set("Access-Control-Allow-Headers", "Authorization, Origin, Accept, X-Requested-With, Content-Type, Access-Control-Request-Method, Access-Control-Request-Headers")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK) // Здесь важно отправить корректный статус!
		return true
	}
	return false
}

// Обработчик для страницы регистрации
func registerHandler(w http.ResponseWriter, r *http.Request) {
	if handleCors(w, r) {
		return
	}
	tmpl, err := template.ParseFiles("templates/register.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}

// Обработчик для отображения формы авторизации
func handleLoginPage(w http.ResponseWriter, r *http.Request) {
	if handleCors(w, r) {
		return
	}
	tmpl, err := template.ParseFiles("templates/login.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}

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

// Обработчик поиска с лимитом подсказок 3
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
	if r.Method == "POST" {
		login := r.FormValue("login")
		password := r.FormValue("password")

		// Проверяем, существует ли пользователь и правильный ли пароль
		err := verifyLoginCredentials(login, password)
		if err != nil {
			http.Error(w, "Ошибка авторизации: "+err.Error(), http.StatusUnauthorized)
			return
		}

		// Если авторизация прошла успешно, выводим сообщение
		message := "Вы успешно авторизовались"
		tmpl, err := template.ParseFiles("templates/message.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		tmpl.Execute(w, message)
	}
}
