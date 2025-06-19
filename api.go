package main

import (
	"encoding/json"
	"net/http"

	"github.com/jmoiron/sqlx"
)

func setupAPI(db *sqlx.DB) {
	http.HandleFunc("/api/menu", func(w http.ResponseWriter, r *http.Request) {
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
	})

	http.ListenAndServe(":8080", nil)
}
