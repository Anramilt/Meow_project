package main

import (
	"fmt"
	//"html/template"
	"net/http"
)

var categories = []string{"Категория 1", "Категория 2", "Категория 3"}

/*const (
	ExpectedUsername = "user"
	ExpectedPassword = "secret"
)*/

func main() {
	//ConnectDB()

	fmt.Println("Hello world!")
	fmt.Println("Hello world!")
	fmt.Println("Hello world!")

	//logger := log.New(os.Stdout, "http: ", log.LstdFlags)

	//регистрирует нашу функцию обработчика  для обработки всех запросов GET
	//http.HandleFunc("/", handler)
	//http.HandleFunc("/login", authHandler) //обработчик авторизации
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/sort", sortHandler)
	http.ListenAndServe(":8080", nil)
	//http.HandleFunc("/registration", addAccountHandler) //добавление аккаунта в БД
	//http.HandleFunc("/echo", auth(echoHandler))         //предоставление доступа (к эхо) по авторизации
	//http.HandleFunc("/addecho", auth(addEchoHandler))   //добавить БД в эхо

}
