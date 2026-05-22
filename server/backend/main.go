package main

import (
	"fmt"
	"log"
)

func main() {
	db, dbErr := InitDB()
	if dbErr != nil {
		log.Fatalf("db error: %v\n", dbErr)
	}
	defer db.Close()

	fmt.Println("bd inited")

	// Добавляем тестовые данные
	// testDefect := defect{"tree", 15}
	// lastID := InsertToDB(db, testDefect)
	// if lastID > 0 {
	// 	fmt.Printf("test data added, id: %d\n", lastID)
	// } else {
	// 	log.Printf("cannot add data")
	// }

	StartServer(db)
}
