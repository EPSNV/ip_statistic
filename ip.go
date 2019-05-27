package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"

	"github.com/jmoiron/sqlx"

	_ "github.com/lib/pq"
)

var menu = []*MenuItem{
	{Name: "Главная", URL: "/"},
	//{Name: "Добавление", URL: "/supply"},
	// {Name: "Цена", URL: "/price"},
	//{Name: "Продажа", URL: "/sell"},
	// {Name: "Статистика", URL: "/stat"},
}

var itemsListTmpl = template.Must(
	template.ParseFiles(
		"./templates/index.html",
		"./templates/items_list.html",
	),
)

type MenuItem struct {
	Name string
	URL  string
}

type tplParams struct {
	Menu       []*MenuItem
	ActiveMenu string
}

type ipStore struct {
	db *sql.DB
}

type Item struct {
	ID          int
	IpAddress   string
	User        string
	Address     string
	Description string
}

func (s *ipStore) Index(w http.ResponseWriter, r *http.Request) {
	err := itemsListTmpl.Execute(w, &tplParams{
		ActiveMenu: r.URL.String(),
		Menu:       menu,
	})
	if err != nil {
		http.Error(w, "error in template", 500)
	}
}

func (s *ipStore) getItems() ([]*Item, error) {

	items := []*Item{}
	query := "SELECT id, name, price, cnt, sum, store_count FROM products ORDER BY id ASC"
	s.db.Select(&items, query)

	// rows, err := s.db.Query("SELECT id, ip_address, user, address, description FROM ip140 ORDER BY id ASC")
	// if err != nil {
	// 	return nil, err
	// }
	// defer rows.Close()
	// for rows.Next() {
	// 	it := &Item{}
	// 	err := rows.Scan(&it.ID, &it.IpAddress, &it.User, &it.Address, &it.Description)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	items = append(items, it)
	// }
	return items, nil
}

func (s *ipStore) ApiGetItems(w http.ResponseWriter, r *http.Request) {
	items, err := s.getItems()
	fmt.Println(err)
	if err != nil {
		w.Write([]byte(`{"status": 500, "error": "db error"}`))
		return
	}
	result, _ := json.Marshal(items)
	w.Write(result)
}

func (s *ipStore) fillSubnet(ip IP)

func main() {
	connStr := "user=root dbname=ip password=root host=127.0.0.1 port=5432 sslmode=disable"
	db, err := sqlx.Connect("postgres", connStr)
	if err != nil {
		log.Fatalln("cant open db:", err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatalln("cant connect to db:", err)
	}

	ipStore := &ipStore{db: db}

	i := net.ParseIP("10.181.131.1")

	//http.HandleFunc("/fill", ipStore.Fill)
	http.HandleFunc("/", ipStore.Index)

	// http.HandleFunc("/supply", shop.Supply)
	// http.HandleFunc("/edit", shop.Edit)
	// http.HandleFunc("/sell", shop.Sell)
	http.HandleFunc("/api/v1/items", ipStore.ApiGetItems)
	// http.HandleFunc("/api/v1/items/delete", shop.ApiDeleteItems)

	/*
		/static/css/bootstrap.css
		после StripPrefix: css/bootstrap.css
		этот файл уже берется из папки static
	*/
	staticHandler := http.StripPrefix(
		"/static/",
		http.FileServer(http.Dir("./static")),
	)
	http.Handle("/static/", staticHandler)

	fmt.Println("starting server at :8080")
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatalln("servger return error:", err)
	}
}
