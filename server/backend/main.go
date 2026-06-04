// package main
//
// import (
// 	"fmt"
// 	"log"
// )
//
// func main() {
// 	fmt.Println("[main] Starting server...")
// 	db, dbErr := InitDB()
// 	if dbErr != nil {
// 		log.Fatalf("db error: %v\n", dbErr)
// 	}
// 	defer db.Close()
//
// 	fmt.Println("[main] bd inited")
//
// 	// Добавляем тестовые данные
// 	// testDefect := defect{"tree", 15}
// 	// lastID := InsertToDB(db, testDefect)
// 	// if lastID > 0 {
// 	// 	fmt.Printf("test data added, id: %d\n", lastID)
// 	// } else {
// 	// 	log.Printf("cannot add data")
// 	// }
//
// }
//
//
//
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

	err := ClrDB(db)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("[main] bd inited")
 	StartServer(db)

}
