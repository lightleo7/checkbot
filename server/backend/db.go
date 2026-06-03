package main

import (
    "database/sql"
    "fmt"
    "log"

    _ "modernc.org/sqlite"
)

type defect struct {
    Type        string `json:"type"`
    Coordinates string `json:"coordinates"`
	TimeSpotted string `json:"time_spotted"`
    Images      string `json:"images"`
}

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
		coordinates TEXT,
		time_spotted TEXT,
		images TEXT
	);`

	_, err = db.Exec(createTableSQL)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func InsertToDB(db *sql.DB, s defect) int64 {
	result, err := db.Exec("INSERT INTO defects (type, coordinates, time_spotted, images) VALUES (?, ?, ?, ?)", s.Type, s.Coordinates, s.TimeSpotted, s.Images)

	if err != nil {
		log.Fatalf("error: %v", err)
	}

	lastID, _ := result.LastInsertId()
	return lastID
}

func SelectFromDB(db *sql.DB, id int) defect {
	var newDefect defect
	err := db.QueryRow("SELECT type, coordinates, time_spotted FROM defects WHERE id = ?", id).Scan(&newDefect.Type, &newDefect.Coordinates, &newDefect.TimeSpotted)

	if err == sql.ErrNoRows {
		fmt.Println("there is no defect with this ID")
	} else if err != nil {
		log.Fatalf("error while executing SELECT: %v", err)
	}

	return newDefect
}

func GetLastID(db *sql.DB) int {
	var id int
	err := db.QueryRow("SELECT id FROM defects ORDER BY id DESC LIMIT 1").Scan(&id)
	if err == sql.ErrNoRows {
		return 0
	} else if err != nil {
		log.Printf("error: %v", err)
		return 0
	}

	return id
}

func GetAllDefects(db *sql.DB) ([]defectResponse, error) {
	var defects []defectResponse

	query := "SELECT id, type, coordinates, time_spotted FROM defects ORDER BY id DESC"
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var d defectResponse
		err := rows.Scan(&d.ID, &d.Type, &d.Coordinates, &d.TimeSpotted)
		if err != nil {
			return nil, err
		}
		defects = append(defects, d)
	}

	return defects, nil
}