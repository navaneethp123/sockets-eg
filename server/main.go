package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	user := "root"      // replace this
	password := "mysql" // replace this
	dataSourceName := fmt.Sprintf("%s:%s@tcp(127.0.0.1:3306)/restaurant", user, password)

	db, err := sql.Open("mysql", dataSourceName)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	http.HandleFunc("/menu", Curry(db, Menu))
	http.HandleFunc("/order", Curry(db, Order))
	http.HandleFunc("/pushnotifs", PushNotifs)

	log.Println("Server listening on http://localhost:5000 ...")
	http.ListenAndServe(":5000", nil)
}
