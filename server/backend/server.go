package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
)

type defectResponse struct {
	ID          int    `json:"id"`
	Type        string `json:"type"`
	Coordinates int    `json:"coordinates"`
}

type pageData struct {
	Title   string
	Defects []defectResponse
}

func StartServer(db *sql.DB) {
	// Настройка обработки статических файлов
	staticDir := "./frontend/static"
	fs := http.FileServer(http.Dir(staticDir))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		tmplPath := filepath.Join(staticDir, "index.html")
		tmpl, err := template.ParseFiles(tmplPath)
		if err != nil {
			log.Printf("Ошибка загрузки шаблона: %v", err)
			http.Error(w, "Ошибка загрузки страницы", http.StatusInternalServerError)
			return
		}

		defects, err := GetAllDefects(db)
		if err != nil {
			log.Printf("Ошибка получения дефектов: %v", err)
			defects = []defectResponse{}
		}

		data := pageData{
			Title:   "Система управления дефектами",
			Defects: defects,
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		err = tmpl.Execute(w, data)
		if err != nil {
			log.Printf("Ошибка выполнения шаблона: %v", err)
		}
	})

	http.HandleFunc("/api/defects", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		
		switch r.Method {
		case http.MethodGet:
			defects, err := GetAllDefects(db)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
				return
			}
			json.NewEncoder(w).Encode(defects)
			
		case http.MethodPost:
			var newDefect defect
			err := json.NewDecoder(r.Body).Decode(&newDefect)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
				return
			}
			
			// InsertToDB возвращает int64
			lastID := InsertToDB(db, newDefect)
			if lastID == 0 {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]string{"error": "Failed to insert defect"})
				return
			}
			
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"message": "Defect added successfully",
				"id": lastID,
			})
			
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
		}
	})

	port := 8080
	fmt.Printf("Server started at http://localhost:%d\n", port)
	
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}
