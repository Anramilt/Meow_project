package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
)

type ErrorResponse struct {
	Error string
}

// Функия ответа на запрос с ошибкой.
func respondWithError(text string, w http.ResponseWriter) {
	res := ErrorResponse{text}
	toWrite, err := json.MarshalIndent(res, "", "\t")
	if err != nil {
		log.Println("Error marshalling data")
		return
	}
	_, err = w.Write(toWrite)
	if err != nil {
		log.Println("Error writing the data")
		return
	}
}

// Функция для предоставления доступа по авторизации
func auth(fn func(w http.ResponseWriter, r *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := strings.TrimSpace(strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer"))
		if !validateToken(token) {
			w.WriteHeader(400)
			respondWithError("Invalid token", w)
			return
		}
		fn(w, r)
	}
}

// r.Method - медот HTTP(GET, POST, PUT) и др.
// r.URL.Path - часть пути URL-адреса
// r.RemoteAddr - IP-адрес клиента, сделавшего запрос
func authHandler(w http.ResponseWriter, r *http.Request) {
	// /login
	logger := log.New(os.Stdout, "http: ", log.LstdFlags)
	logger.Printf("Received request: %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)

	if r.Method != "POST" {
		w.WriteHeader(403)
		respondWithError("Bad method", w)
		return
	}

	var user struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Message  string `json:"message"`
	}
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil || user.Username == "" || user.Password == "" /*|| user.Message == "" */ {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	//Здесь должна быть проверка пользователя в БД
	//
	/*if user.Username != ExpectedUsername && user.Password != ExpectedPassword {
		http.Error(w, "Unauthorization", http.StatusUnauthorized)
		return
	}*/

	//Проверка пользователя в БД
	exists, err := userExist(user.Username, user.Password)
	if err != nil {
		logger.Printf("Database error: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	if !exists {
		http.Error(w, "Unauthorized. Please register at /registration", http.StatusUnauthorized)
		return
	}

	token, err := generateToken(user.Username)
	if err != nil {
		http.Error(w, "Could not create token", http.StatusInternalServerError)
		return
	}
	w.Write([]byte(token))
}

func addAccountHandler(w http.ResponseWriter, r *http.Request) {
	logger := log.New(os.Stdout, "http: ", log.LstdFlags)
	logger.Printf("Received request: %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
	if r.Method != "POST" {
		w.WriteHeader(403)
		respondWithError("Bad method", w)
		return
	}
	var user struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil || user.Username == "" || user.Password == "" {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	err = addAccounttoDB(user.Username, user.Password)
	if err != nil {
		//logger.Printf("Error writing to DB: %v", err)
		http.Error(w, "Error write in DB", http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "Account saved in DB: %s\n", user.Username)
}

func echoHandler(w http.ResponseWriter, r *http.Request) {
	// /echo
	logger := log.New(os.Stdout, "http: ", log.LstdFlags)
	logger.Printf("Received request: %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)

	if r.Method == http.MethodGet {
		/*token := r.Header.Get("Authorization")
		token = strings.TrimPrefix(token, "Bearer ")
		log.Println(token)
		if !validateToken(token) {
			http.Error(w, "Unauthorized!", http.StatusUnauthorized)
			return
		}*/

		message, err := getEchomessageList()
		if err != nil {
			http.Error(w, "Error read in DB", http.StatusInternalServerError)
			return
		}

		var response string
		for _, msg := range message {
			response += fmt.Sprintf("ID: %d, Message: %s \n", msg.ID, msg.Message)
		}

		w.Write([]byte(response))
		return
	}

}
func addEchoHandler(w http.ResponseWriter, r *http.Request) {
	// /add echo
	logger := log.New(os.Stdout, "http: ", log.LstdFlags)
	logger.Printf("Received request: %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)

	if r.Method != "POST" {
		w.WriteHeader(403)
		respondWithError("Bad method", w)
		return
	}

	var user struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Message  string `json:"message"`
	}
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil || user.Username == "" || user.Password == "" || user.Message == "" {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}
	//нужна ли проверка ли обертке?
	/*token := r.Header.Get("Authorization")
	token = strings.TrimPrefix(token, "Bearer ")
	log.Println(token)
	if !validateToken(token) {
		http.Error(w, "Unauthorized!", http.StatusUnauthorized)
		return
	}*/

	message := user.Message
	err = addEchotoDB(message)
	if err != nil {
		http.Error(w, "Error write in DB", http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Message saved in DB: %s\n", message)
}
