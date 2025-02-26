package main

import (
	"encoding/json"
	"html/template"
	"net/http"
)

// Обработчик для главной страницы
func indexHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}

// Обработчик поиска
func searchHandler(w http.ResponseWriter, r *http.Request) {
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
		WHERE game.name_game ILIKE $1;`
	err := db.Select(&categories, sqlQuery, query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(categories)
}
