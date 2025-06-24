package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	ConnectDB()

	initVosk()
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	fmt.Println("Hello world!")
	fmt.Println("Hello world!")
	fmt.Println("Hello world!")

	// Генерация и добавление 5 ключей
	/*if err := addKeysToDB(5); err != nil {
		log.Fatal("Ошибка при добавлении ключей в БД:", err)
	}*/

	//Добавление картинок в БД:
	//uploadImagesToDB(imageDir)

	/*rootDir := "/home/sofia/Документы/Menu" // путь к корневой папке */
	rootDir := "/home/sofia/Test" // путь к корневой папке
	if err := processAllJsonFiles(rootDir); err != nil {
		log.Fatal("Ошибка обработки файлов:", err)
	}
	//setupAPI(db)

	//http.HandleFunc("/", handler)
	//http.HandleFunc("/login", authHandler) //обработчик авторизации
	//http.HandleFunc("/", indexHandler)
	http.HandleFunc("/search", searchHandler)
	http.HandleFunc("/searchlimit", searchlimitHandler)
	//http.HandleFunc("/register", registerHandler)       // Страница регистрации
	http.HandleFunc("/handle-register", handleRegister) // Обработчик регистрации
	//http.HandleFunc("/login", handleLoginPage)
	http.HandleFunc("/handle-login", handleLogin)
	http.HandleFunc("/files", fileHandler)
	http.HandleFunc("/images", imagesearchHandler)
	http.HandleFunc("/sound", soundsearchHandler)

	http.HandleFunc("/game", gameHandler)
	//http.HandleFunc("/api/message", apiHandler)

	/*http.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir("./uploads"))))
	http.HandleFunc("/api/upload-voice", uploadVoiceHandler)
	*/

	http.HandleFunc("/menu", menuHandler)
	//http.HandleFunc("/api/menu", menuDownloadHandler)
	fmt.Println("Gut")

	http.ListenAndServe(":8080", nil)
}
