package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type defectResponse struct {
	ID          int      `json:"id"`
	Type        string   `json:"type"`
	Coordinates string   `json:"coordinates"`
	Images      []string `json:"images,omitempty"`
}

type pageData struct {
	Title   string
	Defects []defectResponse
}

func StartServer(db *sql.DB) {
	if err := db.Ping(); err != nil {
		log.Fatalf("[DB] bd connect error: %v", err)
	}

	staticDir := "./frontend/static"
	fs := http.FileServer(http.Dir(staticDir))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir("../uploads"))))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			log.Printf("[HTTP] 404 Not Found: %s %s", r.Method, r.URL.Path)
			http.NotFound(w, r)
			return
		}

		tmplPath := filepath.Join(staticDir, "index.html")
		tmpl, err := template.ParseFiles(tmplPath)
		if err != nil {
			log.Printf("[HTML] Ошибка загрузки шаблона: %v", err)
			http.Error(w, "Ошибка загрузки страницы", http.StatusInternalServerError)
			return
		}

		defects, err := GetAllDefects(db)
		if err != nil {
			log.Printf("[DB] Ошибка получения дефектов: %v", err)
			defects = []defectResponse{}
		}

		data := pageData{
			Title:   "Система управления дефектами",
			Defects: defects,
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		err = tmpl.Execute(w, data)
		if err != nil {
			log.Printf("[HTML] Ошибка рендеринга страницы: %v", err)
			return
		}
		log.Printf("[HTTP] 200 OK: %s %s", r.Method, r.URL.Path)
	})

	http.HandleFunc("/defect", func(w http.ResponseWriter, r *http.Request) {
		idStr := r.URL.Query().Get("id")
		if idStr == "" {
			http.Error(w, "Missing id parameter", http.StatusBadRequest)
			return
		}

		tmplPath := filepath.Join(staticDir, "defect.html")
		tmpl, err := template.ParseFiles(tmplPath)
		if err != nil {
			http.Error(w, "Ошибка загрузки страницы дефекта", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		tmpl.Execute(w, nil)
	})

	http.HandleFunc("/api/defects/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		idStr := strings.TrimPrefix(r.URL.Path, "/api/defects/")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Invalid ID"})
			return
		}

		var d defectResponse
		var imagesStr sql.NullString
		err = db.QueryRow("SELECT id, type, coordinates, images FROM defects WHERE id = ?", id).Scan(&d.ID, &d.Type, &d.Coordinates, &imagesStr)
		if err != nil {
			if err == sql.ErrNoRows {
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(map[string]string{"error": "Defect not found"})
			} else {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			}
			return
		}

		if imagesStr.Valid && imagesStr.String != "" {
			rawPaths := strings.Split(imagesStr.String, ",")
			for _, path := range rawPaths {
				cleanPath := strings.TrimPrefix(path, "../uploads/")
				d.Images = append(d.Images, "/uploads/"+cleanPath)
			}
		} else {
			d.Images = []string{}
		}

		json.NewEncoder(w).Encode(d)
	})

	http.HandleFunc("/api/defects", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		
		switch r.Method {
		case http.MethodGet:
			defects, err := GetAllDefects(db)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
				log.Printf("[API] 500 Internal Server Error: GET /api/defects (bd error: %v)", err)
				return
			}
			json.NewEncoder(w).Encode(defects)
			log.Printf("[API] 200 OK: GET /api/defects")
			
		case http.MethodPost:
			err := r.ParseMultipartForm(32 << 20)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]string{"error": "Invalid multipart form"})
				log.Printf("[API] 400 Bad Request: POST /api/defects (Bad multipart form)")
				return
			}

			var newDefect defect
			newDefect.Type = r.FormValue("Type")
			newDefect.Coordinates = r.FormValue("Coordinates")

			if newDefect.Type == "" || newDefect.Coordinates == "" {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]string{"error": "Missing Type or Coordinates"})
				log.Printf("[API] 400 Bad Request: POST /api/defects (Missing Type или Coordinates)")
				return
			}

			lastID := InsertToDB(db, newDefect)
			if lastID == 0 {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]string{"error": "Failed to insert defect text data"})
				log.Printf("[API] 500 Internal Server Error: POST /api/defects (Failed to insert defect text data)")
				return
			}

			uploadDir := "../uploads"
			if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				log.Printf("[FILE] 500 Error creating folder %s: %v", uploadDir, err)
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
					log.Printf("[FILE] Error opening file %s: %v", fileKey, err)
					continue
				}
				defer file.Close()

				filename := fmt.Sprintf("defect_%d_img_%s%s", lastID, fileKey, filepath.Ext(fileHeader.Filename))
				targetPath := filepath.Join(uploadDir, filename)

				dst, err := os.Create(targetPath)
				if err != nil {
					log.Printf("[FILE] Error creating file %s: %v", targetPath, err)
					continue
				}
				defer dst.Close()

				_, err = io.Copy(dst, file)
				if err != nil {
					log.Printf("[FILE] Error writing file %s: %v", targetPath, err)
					continue
				}

				webPath := fmt.Sprintf("../uploads/%s", filename)
				savedFilePaths = append(savedFilePaths, webPath)
			}

			if len(savedFilePaths) > 0 {
				imagesList := strings.Join(savedFilePaths, ",")
				_, err = db.Exec("UPDATE defects SET images = ? WHERE id = ?", imagesList, lastID)
				if err != nil {
					log.Printf("[DB] Ошибка обновления путей картинок для ID %d: %v", lastID, err)
				}
			}
			
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"message": "Defect and images uploaded successfully",
				"id":      lastID,
				"images":  savedFilePaths,
			})
			log.Printf("[API] 201 Created: POST /api/defects (Created defect ID: %d, files count: %d)", lastID, len(savedFilePaths))
			
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
			log.Printf("[API] 405 Method Not Allowed: %s /api/defects", r.Method)
		}
	})

	port := 8080
	log.Printf("[SERVER] Server started at http://localhost:%d", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}
