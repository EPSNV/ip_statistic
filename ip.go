package main

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
)

type ipStore struct {
	db *sql.DB
}

func main() {
	connStr := "user=root dbname=test password=root host=127.0.0.1 port=5432 sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalln("cant open db:", err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatalln("cant connect to db:", err)
	}

	ipStore := &ipStore{db: db}

	log.Println(ipStore.db)

	// http.HandleFunc("/", shop.Index)
	// http.HandleFunc("/supply", shop.Supply)
	// http.HandleFunc("/edit", shop.Edit)
	// http.HandleFunc("/sell", shop.Sell)
	// http.HandleFunc("/api/v1/items", shop.ApiGetItems)
	// http.HandleFunc("/api/v1/items/delete", shop.ApiDeleteItems)

	/*
		/static/css/bootstrap.css
		после StripPrefix: css/bootstrap.css
		этот файл уже берется из папки static
	*/
	// staticHandler := http.StripPrefix(
	// 	"/static/",
	// 	http.FileServer(http.Dir("./static")),
	// )
	// http.Handle("/static/", staticHandler)

	// fmt.Println("starting server at :8080")
	// err = http.ListenAndServe(":8080", nil)
	// if err != nil {
	// 	log.Fatalln("servger return error:", err)
	// }
}
