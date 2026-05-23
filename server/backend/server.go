package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"io"
	"os"
	"strings"
)

type defectResponse struct {
	ID          int    `json:"id"`
	Type        string `json:"type"`
	Coordinates string `json:"coordinates"`
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
			// ограничение в 32 мб
			err := r.ParseMultipartForm(32 << 20)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]string{"error": "Invalid multipart form"})
				return
			}

			// текст
			var newDefect defect
			newDefect.Type = r.FormValue("Type")
			newDefect.Coordinates = r.FormValue("Coordinates")

			if newDefect.Type == "" || newDefect.Coordinates == "" {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]string{"error": "Missing Type or Coordinates"})
				return
			}

			lastID := InsertToDB(db, newDefect)
			if lastID == 0 {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]string{"error": "Failed to insert defect text data"})
				return
			}

			uploadDir := "../uploads"
			if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
				log.Printf("Ошибка создания директории: %v", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			var savedFilePaths []string

			form := r.MultipartForm
			for fileKey, fileHeaders := range form.File {
				if len(fileHeaders) == 0 {
					continue
				}
				fileHeader := fileHeaders[0]

				file, err := fileHeader.Open()
				if err != nil {
					log.Printf("Ошибка открытия файла %s: %v", fileKey, err)
					continue
				}
				defer file.Close()

				filename := fmt.Sprintf("defect_%d_img_%s%s", lastID, fileKey, filepath.Ext(fileHeader.Filename))
				targetPath := filepath.Join(uploadDir, filename)

				dst, err := os.Create(targetPath)
				if err != nil {
					log.Printf("Ошибка создания файла на диске: %v", err)
					continue
				}
				defer dst.Close()

				_, err = io.Copy(dst, file)
				if err != nil {
					log.Printf("Ошибка копирования байт: %v", err)
					continue
				}

				webPath := fmt.Sprintf("../uploads/%s", filename)
				savedFilePaths = append(savedFilePaths, webPath)
			}

			if len(savedFilePaths) > 0 {
				imagesList := strings.Join(savedFilePaths, ",")
				_, err = db.Exec("UPDATE defects SET images = ? WHERE id = ?", imagesList, lastID)
				if err != nil {
					log.Printf("Ошибка при обновлении путей к изображениям в БД: %v", err)
				}
			}
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"message": "Defect and images uploaded successfully",
				"id":      lastID,
				"images":  savedFilePaths,
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
