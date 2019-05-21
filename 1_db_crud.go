package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

type Item struct {
	ID         int
	Name       string
	StoreCount int
	Price      int
	Cnt        int
	Sum        int
}

func main() {
	// инициализация базы
	connStr := "user=root dbname=mydb password=123 host=127.0.0.1 port=5432 sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalln("cant open db:", err)
	}

	// реальное подключение произойдет тут
	err = db.Ping()
	if err != nil {
		log.Fatalln("cant connect to db:", err)
	}

	tablesQuery := `
	CREATE SEQUENCE IF NOT EXISTS products_id_seq INCREMENT 1 MINVALUE 1 MAXVALUE 2147483647 START 1 CACHE 1;
	CREATE TABLE IF NOT EXISTS "products" (
		"id" integer DEFAULT nextval('products_id_seq') NOT NULL,
		"name" character varying(250) NOT NULL,
		"price" integer DEFAULT '0' NOT NULL,
		"cnt" integer DEFAULT '0' NOT NULL,
		"sum" integer DEFAULT '0' NOT NULL,
		"store_count" integer DEFAULT '0' NOT NULL,
		CONSTRAINT "products_name" UNIQUE ("name")
	) WITH (oids = false);
	`
	db.Exec(tablesQuery)

	// выборка множества записей
	items := []*Item{}
	query := "SELECT id, name, price, cnt, sum, store_count FROM products ORDER BY id ASC"
	rows, err := db.Query(query)
	if err != nil {
		log.Fatalln(err)
	}
	// обязательно закрывать коннект после Query
	defer rows.Close()
	for rows.Next() {
		it := &Item{}
		err := rows.Scan(&it.ID, &it.Name, &it.Price, &it.Cnt, &it.Sum, &it.StoreCount)
		if err != nil {
			log.Fatalln(err)
		}
		items = append(items, it)
		fmt.Printf("item from db: %#v\n", it)

	}
	fmt.Printf("total items: %#v\n", len(items))

	// выборка одной записи
	row := db.QueryRow(`SELECT name FROM products WHERE "id" = $1`, 1000000)
	var productName string
	err = row.Scan(&productName)
	// проверка на то что запись нашлась
	if err == sql.ErrNoRows {
		fmt.Println("Не найдено записей c id 1000000")
	}

	// вставка записи
	var lastInsertID int
	query = `INSERT INTO products("name", "store_count") VALUES($1, $2) RETURNING id`
	row = db.QueryRow(query, "тест", 10)
	err = row.Scan(&lastInsertID)
	fmt.Println("insert:", lastInsertID, "err:", err)

	// обновление записи
	query = `UPDATE products SET store_count = store_count + $1 WHERE id = $2`
	res, _ := db.Exec(query, 1, lastInsertID)
	rowsAffected, _ := res.RowsAffected()
	fmt.Println("update rowsAffected", rowsAffected)

	// выборка одной записи
	var storeCount int
	row = db.QueryRow(`SELECT store_count FROM products WHERE "id" = $1`, lastInsertID)
	row.Scan(&storeCount)
	if err == sql.ErrNoRows {
		fmt.Println("Не найдено записи c id", lastInsertID)
	}
	fmt.Println("storeCount", storeCount)

	// удаление записи
	res, err = db.Exec("DELETE FROM products WHERE id = $1", lastInsertID)
	if err != nil {
		log.Fatalln(err)
	}
	rowsAffected, _ = res.RowsAffected()
	fmt.Println("delete rowsAffected", rowsAffected)
}
