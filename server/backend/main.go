package main

import (
	"fmt"
	"log"
)

func main() {
	fmt.Println("[main] Starting server...")
	db, dbErr := InitDB()
	if dbErr != nil {
		log.Fatalf("db error: %v\n", dbErr)
	}
	defer db.Close()

	fmt.Println("[main] bd inited")
 	StartServer(db)

}
