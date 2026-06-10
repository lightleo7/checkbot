package main

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"golang.org/x/crypto/bcrypt"
)

var (
	sessionToken string
	sessionMu    sync.RWMutex
	expectedHash string
)

func LoadPasswordHash() {
	path := filepath.Join("password.txt")
	data, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("[AUTH] Не удалось прочитать password.txt: %v. Проверьте, что файл существует в корне проекта.", err)
	}
	expectedHash = strings.TrimSpace(string(data))
	log.Println("[AUTH] Хеш пароля успешно загружен")
}

func generateToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

func GetOrCreateSession() string {
	sessionMu.Lock()
	defer sessionMu.Unlock()
	if sessionToken == "" {
		sessionToken = generateToken()
	}
	return sessionToken
}

func CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(expectedHash), []byte(password))
	return err == nil
}

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		apiToken := r.Header.Get("X-API-Token")

		cookie, err := r.Cookie("session_token")

		sessionMu.RLock()
		currentSession := sessionToken
		sessionMu.RUnlock()

		if currentSession == "" || (apiToken != currentSession && (err != nil || cookie.Value != currentSession)) {
			if strings.HasPrefix(r.URL.Path, "/api/") {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(map[string]string{"error": "Unauthorized"})
				return
			}
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		next(w, r)
	}
}