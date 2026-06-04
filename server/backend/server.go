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
	"time"
)

type defectResponse struct {
	ID          int      `json:"id"`
	Type        string   `json:"type"`
	Coordinates string   `json:"coordinates"`
	TimeSpotted string   `json:"time_spotted"`
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
	
	LoadPasswordHash()

	staticDir := "./frontend/static"
	fs := http.FileServer(http.Dir(staticDir))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.Handle("/uploads/", AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		http.StripPrefix("/uploads/", http.FileServer(http.Dir("./uploads"))).ServeHTTP(w, r)
	}))

	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			tmplPath := filepath.Join(staticDir, "login.html")
			tmpl, err := template.ParseFiles(tmplPath)
			if err != nil {
				http.Error(w, "Ошибка загрузки страницы логина", http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			tmpl.Execute(w, nil)
			return
		}

		if r.Method == http.MethodPost {
			password := r.FormValue("password")

			// Используем проверку из auth.go
			if !CheckPassword(password) {
				http.Redirect(w, r, "/login?error=1", http.StatusSeeOther)
				return
			}

			token := GetOrCreateSession()

			http.SetCookie(w, &http.Cookie{
				Name:     "session_token",
				Value:    token,
				Expires:  time.Now().Add(24 * time.Hour),
				HttpOnly: true,
				Path:     "/",
			})

			http.Redirect(w, r, "/", http.StatusSeeOther)
		}
	})

	http.HandleFunc("/api/auth", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			Password string `json:"password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if !CheckPassword(req.Password) {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "Wrong password"})
			return
		}

		token := GetOrCreateSession()

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"token": token})
	})

	// protected

	http.HandleFunc("/", AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
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
	}))

	http.HandleFunc("/defect", AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
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
	}))

	http.HandleFunc("/api/defects/", AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
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
				cleanPath := strings.TrimPrefix(path, "./uploads/")
				d.Images = append(d.Images, "/uploads/"+cleanPath)
			}
		} else {
			d.Images = []string{}
		}

		json.NewEncoder(w).Encode(d)
	}))

	http.HandleFunc("/api/defects", AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
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
			err := r.ParseMultipartForm(32 << 20)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]string{"error": "Invalid multipart form"})
				return
			}

			var newDefect defect
			newDefect.Type = r.FormValue("Type")
			newDefect.Coordinates = r.FormValue("Coordinates")
			newDefect.TimeSpotted = r.FormValue("TimeSpotted")

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

			uploadDir := "./uploads"
			if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
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
					continue
				}
				defer file.Close()

				filename := fmt.Sprintf("defect_%d_img_%s%s", lastID, fileKey, filepath.Ext(fileHeader.Filename))
				targetPath := filepath.Join(uploadDir, filename)

				dst, err := os.Create(targetPath)
				if err != nil {
					continue
				}
				defer dst.Close()

				_, err = io.Copy(dst, file)
				if err != nil {
					continue
				}

				webPath := fmt.Sprintf("./uploads/%s", filename)
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

			report := DamageReport{
				ID:           strconv.FormatInt(lastID, 10),
				Date:         time.Now().Format("2006-01-02"),
				TimeSpotted:  r.FormValue("TimeSpotted"),
				TimeReceived: time.Now().Format("15:04:05"),
				TypeName:     r.FormValue("Type"),
				Coordinates:  r.FormValue("Coordinates"),
				ImagePaths:   savedFilePaths,
			}

			err_report := ExportSingleReport(report, "./reports")
			if err_report != nil {
				fmt.Printf("[REPORTS] Error creating report #%d: %v\n", lastID, err)
				return
			}

		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
		}
	}))

	port := 8080
	log.Printf("[SERVER] Server started at http://localhost:%d", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}
