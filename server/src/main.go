package main

import "fmt"

type defect struct {
	Type        string
	Coordinates int
}

func main() {
	db, dbErr := InitDB() // обязательно в начале программы
	if dbErr != nil {
		fmt.Printf("db error: %v\n", dbErr)
		return
	}

	// далее я просто тестил базы данных

	var newDefect = defect{"tree", 15} // создание структуры

	InsertToDB(db, newDefect) // добавление ее в БД

	lastID := GetLastID(db) // получение последнего ID в БД
	var oldDefect = SelectFromDB(db, lastID) // получение последней строки в БД

	// вывод последней строки в БД и последнего ID
	fmt.Printf("the last row in the DB: %v. last ID: %d\n", oldDefect, lastID)
}
