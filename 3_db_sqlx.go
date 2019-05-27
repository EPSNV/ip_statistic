package main

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type Item struct {
	ID         int
	Name       string
	StoreCount int `db:"store_count"`
	Price      int
	Cnt        int
	Sum        int
}

func main() {
	// инициализация базы
	connStr := "user=root dbname=mydb password=123 host=127.0.0.1 port=5432 sslmode=disable"
	db, err := sqlx.Connect("postgres", connStr)
	if err != nil {
		log.Fatalln("cant open db:", err)
	}

	// реальное подключение произойдет тут
	err = db.Ping()
	if err != nil {
		log.Fatalln("cant connect to db:", err)
	}

	// выборка множества записей
	items := []*Item{}
	query := "SELECT id, name, price, cnt, sum, store_count FROM products ORDER BY id ASC"
	db.Select(&items, query)
	for _, item := range items {
		fmt.Printf("item %#v\n", item)
	}
	fmt.Println("total items", len(items))

	// выборка одной записи
	item := Item{}
	err = db.Get(&item, `SELECT name FROM products WHERE "id" = $1`, 1000000)
	if err == sql.ErrNoRows {
		fmt.Println("Не найдено записей c id 1000000")
	}

	// вставка записи, вариант 1
	var lastInsertID int
	query = `INSERT INTO products("name", "store_count") VALUES($1, $2) RETURNING id`
	err = db.Get(&lastInsertID, query, "test", 1)
	if err == sql.ErrNoRows {
		fmt.Println("Не найдено записей c id 1000000")
	}
	fmt.Println("lastInsertID", lastInsertID, err)

	query = `INSERT INTO products("name", "store_count") VALUES(:name, :store_count) RETURNING id`
	rows, err := db.NamedQuery(query, &Item{Name: "test2", StoreCount: 44})
	if err != nil {
		log.Fatalln(err)
	}
	if rows.Next() {
		rows.Scan(&lastInsertID)
	}
	fmt.Println("lastInsertID from NamedQuery", lastInsertID, err)

	// обновление записи
	query = `UPDATE products SET store_count = store_count + $1 WHERE id = $2`
	res, _ := db.Exec(query, 1, lastInsertID)
	rowsAffected, _ := res.RowsAffected()
	fmt.Println("update rowsAffected", rowsAffected)

	// выборка одной записи
	result := make(map[string]interface{})
	row := db.QueryRowx(`SELECT store_count, name, cnt, sum FROM products WHERE "id" = $1`, lastInsertID)
	err = row.Scan(result)
	if err == sql.ErrNoRows {
		fmt.Println("Не найдено записи c id", lastInsertID)
	}
	fmt.Println("storeCount", result)

	// удаление записи
	res, _ = db.Exec("DELETE FROM products WHERE id = $1", lastInsertID)
	rowsAffected, _ = res.RowsAffected()
	fmt.Println("delete rowsAffected", rowsAffected)

	// я делал 2 INSERT-а, чищу 2-й
	res, _ = db.Exec("DELETE FROM products WHERE id = $1", lastInsertID-1)
	rowsAffected, _ = res.RowsAffected()
	fmt.Println("delete rowsAffected", rowsAffected)
}
