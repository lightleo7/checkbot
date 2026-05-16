package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "modernc.org/sqlite"
)

// InitDB -- инициализация базы данных
// использование: db, dvErr := InitDB()
// возвращает саму БД и код ошибки
func InitDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite", "database.db")
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}

	createTableSQL := `
	CREATE TABLE IF NOT EXISTS defects (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		type TEXT,
		coordinates INT 
	);`

	_, err = db.Exec(createTableSQL)
	if err != nil {
		return nil, err
	}
	return db, nil
}

// InsertToDB -- добавляет в базу данных структуру
// использование: *int64* := InsertToDB(БД, структура_defect)
// возвращает ID новой строки
func InsertToDB(db *sql.DB, s defect) int64 {
	result, err := db.Exec("INSERT INTO defects (type, coordinates) VALUES (?, ?)", s.Type, s.Coordinates)

	if err != nil {
		log.Fatalf("error: %v", err)
	}

	lastID, _ := result.LastInsertId()
	return lastID
}

// SelectFromDB -- получение последней строки из БД
// возвращает структуру defect
// использование: *defect* := SelectFromDB(БД, ID_строки)
func SelectFromDB(db *sql.DB, id int) defect {
	var newDefect defect
	err := db.QueryRow("SELECT type, coordinates FROM defects WHERE id = ?", id).Scan(&newDefect.Type, &newDefect.Coordinates)

	if err == sql.ErrNoRows {
		fmt.Println("there is no defect with this ID")
	} else if err != nil {
		log.Fatalf("error while executing SELECT: %v", err)
	}

	return newDefect
}

// GetLastID -- получение последнего ID в БД
// использование: *int* := GetLastID(БД)
func GetLastID(db *sql.DB) int {
	var id int
	err := db.QueryRow("SELECT id FROM Defects ORDER BY id DESC LIMIT 1").Scan(&id)
	if err == sql.ErrNoRows {
		return 0
	} else if err != nil {
		log.Printf("error: %v", err)
		return 0
	}

	return id
}
