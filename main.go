package main

import (
	"fmt"
	"net/http"
)

/*const (
	ExpectedUsername = "user"
	ExpectedPassword = "secret"
)*/

func main() {
	ConnectDB()
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	fmt.Println("Hello world!")
	fmt.Println("Hello world!")
	fmt.Println("Hello world!")

	//регистрирует нашу функцию обработчика  для обработки всех запросов GET
	//http.HandleFunc("/", handler)
	//http.HandleFunc("/login", authHandler) //обработчик авторизации
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/search", searchHandler)
	http.ListenAndServe(":8080", nil)

	//http.HandleFunc("/registration", addAccountHandler) //добавление аккаунта в БД
	//http.HandleFunc("/echo", auth(echoHandler))         //предоставление доступа (к эхо) по авторизации
	//http.HandleFunc("/addecho", auth(addEchoHandler))   //добавить БД в эхо

}
