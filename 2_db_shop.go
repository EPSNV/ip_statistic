package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"

	_ "github.com/lib/pq"
)

type MenuItem struct {
	Name string
	URL  string
}

type tplParams struct {
	Menu       []*MenuItem
	ActiveMenu string
}

type tplParamsWithPage struct {
	Menu       []*MenuItem
	ActiveMenu string
	Page       map[string]interface{}
}

type Msg struct {
	Class string
	Text  string
}

type tplMsg struct {
	Menu       []*MenuItem
	ActiveMenu string
	Msg        Msg
}

var menu = []*MenuItem{
	{Name: "Главная", URL: "/"},
	{Name: "Поставка", URL: "/supply"},
	// {Name: "Цена", URL: "/price"},
	{Name: "Продажа", URL: "/sell"},
	// {Name: "Статистика", URL: "/stat"},
}

var (
	itemsListTmpl = template.Must(
		template.ParseFiles(
			"./templates/index.html",
			"./templates/items_list.html",
		),
	)
	tmpl = template.Must(
		template.ParseFiles(
			"./templates/index.html",
			"./templates/supply_form.html",
		),
	)
	tmplEdit = template.Must(
		template.ParseFiles(
			"./templates/index.html",
			"./templates/edit_form.html",
		),
	)
	tmplSell = template.Must(
		template.ParseFiles(
			"./templates/index.html",
			"./templates/sell_form.html",
		),
	)
	msg = template.Must(
		template.ParseFiles(
			"./templates/index.html",
			"./templates/msg.html",
		),
	)
)

type Item struct {
	ID         int
	Name       string
	StoreCount int
	Price      int
	Cnt        int
	Sum        int
}

type Shop struct {
	db *sql.DB
}

func (s *Shop) Index(w http.ResponseWriter, r *http.Request) {
	err := itemsListTmpl.Execute(w, &tplParams{
		ActiveMenu: r.URL.String(),
		Menu:       menu,
	})
	if err != nil {
		http.Error(w, "error in template", 500)
	}
}

func (s *Shop) getItems() ([]*Item, error) {
	items := []*Item{}
	rows, err := s.db.Query("SELECT id, name, price, cnt, sum, store_count FROM products ORDER BY id ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		it := &Item{}
		err := rows.Scan(&it.ID, &it.Name, &it.Price, &it.Cnt, &it.Sum, &it.StoreCount)
		if err != nil {
			return nil, err
		}
		items = append(items, it)
	}
	return items, nil
}

func (s *Shop) ApiGetItems(w http.ResponseWriter, r *http.Request) {
	items, err := s.getItems()
	if err != nil {
		w.Write([]byte(`{"status": 500, "error": "db error"}`))
		return
	}
	result, _ := json.Marshal(items)
	w.Write(result)
}

func (s *Shop) ApiDeleteItems(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		w.Write([]byte(`{"status": 400, "error: "bad method"}`))
		return
	}
	ID, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		w.Write([]byte(`{"status": 400, "error": "bad id"}`))
		return
	}

	_, err = s.db.Exec("DELETE FROM products WHERE id = $1", ID)
	if err != nil {
		w.Write([]byte(`{"status": 500, "error": "db err"}`))
		return
	}
	w.Write([]byte(`{"status": 200, "error": ""}`))
}

func (s *Shop) Supply(w http.ResponseWriter, r *http.Request) {
	// если это не отправка формы - показываем HTML
	if r.Method != http.MethodPost {
		s.SupplyShowForm(w, r)
		return
	}
	s.SupplyProcessForm(w, r)
}

func (s *Shop) SupplyShowForm(w http.ResponseWriter, r *http.Request) {
	params := tplParams{
		ActiveMenu: r.URL.String(),
		Menu:       menu,
	}

	err := tmpl.Execute(w, params)
	if err != nil {
		// не раскрываем детали ошибки пользователю
		log.Println("error in template:", err)
		http.Error(w, "error in template", 500)
	}
}

func (s *Shop) SupplyProcessForm(w http.ResponseWriter, r *http.Request) {
	inputItemName := r.FormValue("itemName")
	cnt, err := strconv.Atoi(r.FormValue("cnt"))
	if err != nil {
		msg.Execute(w, tplMsg{
			Menu:       menu,
			ActiveMenu: r.URL.String(),
			Msg: Msg{
				Class: "danger",
				Text:  "Введите число в количестве",
			},
		})
		return
	}

	var productID int
	row := s.db.QueryRow(`SELECT id FROM products WHERE "name" = $1`, inputItemName)
	err = row.Scan(&productID)
	if err != sql.ErrNoRows {
		query := `UPDATE products SET store_count = store_count + $1 WHERE id = $2`
		_, err = s.db.Exec(query, cnt, productID)
	} else {
		INSERT
		query := `INSERT INTO products("name", "store_count") VALUES($1, $2)`
		_, err = s.db.Exec(query, inputItemName, cnt)
	}
	params := tplMsg{
		Menu:       menu,
		ActiveMenu: r.URL.String(),
		Msg: Msg{
			Class: "primary",
			Text:  "Товар добавлен",
		},
	}
	if err != nil && err != sql.ErrNoRows {
		params.Msg.Class = "danger"
		params.Msg.Text = "Ошибка добавления в базу"
		log.Println("db err:", err)
	}

	err = msg.Execute(w, params)
	if err != nil {
		// не раскрываем детали ошибки пользователю
		log.Println("error in template:", err)
		http.Error(w, "error in template", 500)
	}
}

// ----------

func (s *Shop) Edit(w http.ResponseWriter, r *http.Request) {
	// если это не отправка формы - показываем HTML
	if r.Method != http.MethodPost {
		s.EditShowForm(w, r)
		return
	}
	s.EditProcessForm(w, r)
}

func (s *Shop) EditShowForm(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.FormValue("id"))
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	product := &Item{}
	row := s.db.QueryRow(`SELECT name, price FROM products WHERE "id" = $1`, id)
	err = row.Scan(&product.Name, &product.Price)
	if err == sql.ErrNoRows {
		http.Error(w, err.Error(), 404)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	params := tplParamsWithPage{
		Menu:       menu,
		ActiveMenu: r.URL.String(),
		Page: map[string]interface{}{
			"ItemName": product.Name,
			"Price":    product.Price,
			"ID":       id,
		},
	}
	err = tmplEdit.Execute(w, params)
	if err != nil {
		// не раскрываем детали ошибки пользователю
		log.Println("error in template:", err)
		http.Error(w, "error in template", 500)
	}
}

func (s *Shop) EditProcessForm(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.FormValue("id"))
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	price, err := strconv.Atoi(r.FormValue("price"))
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	query := `UPDATE products SET price = $1 WHERE id = $2`
	_, err = s.db.Exec(query, price, id)

	params := tplMsg{
		Menu:       menu,
		ActiveMenu: r.URL.String(),
		Msg: Msg{
			Class: "primary",
			Text:  "Цена изменена",
		},
	}
	if err != nil {
		params.Msg.Class = "danger"
		params.Msg.Text = "Ошибка изменения в базу"
		log.Println("db err:", err)
	}

	err = msg.Execute(w, params)
	if err != nil {
		// не раскрываем детали ошибки пользователю
		log.Println("error in template:", err)
		http.Error(w, "error in template", 500)
	}
}

// ----------

func (s *Shop) Sell(w http.ResponseWriter, r *http.Request) {
	// если это не отправка формы - показываем HTML
	if r.Method != http.MethodPost {
		s.SellShowForm(w, r)
		return
	}
	s.SellProcessForm(w, r)
}

func (s *Shop) SellShowForm(w http.ResponseWriter, r *http.Request) {
	items, err := s.getItems()
	if err != nil {
		http.Error(w, err.Error(), 500)
	}
	params := tplParamsWithPage{
		Menu:       menu,
		ActiveMenu: r.URL.String(),
		Page: map[string]interface{}{
			"Items": items,
		},
	}
	err = tmplSell.Execute(w, params)
	if err != nil {
		// не раскрываем детали ошибки пользователю
		log.Println("error in template:", err)
		http.Error(w, "error in template", 500)
	}

}

func (s *Shop) SellProcessForm(w http.ResponseWriter, r *http.Request) {

	r.ParseForm()

	ids := r.Form["ids"]
	cnts := r.Form["cnts"]

	fmt.Println(ids, cnts)

	if len(ids) != len(cnts) {
		http.Error(w, "bad form len", 400)
		return
	}

	upd := []*Item{}
	// плохо если товаров много, надо валидирновать отдельно
	items, err := s.getItems()
	storeCount := map[int]int{}
	for _, it := range items {
		storeCount[it.ID] = it.StoreCount
	}

	for idx, idStr := range ids {
		cnt := cnts[idx]
		if cnt == "" || cnt == "0" {
			continue
		}
		it := &Item{}
		it.ID, err = strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		it.Cnt, err = strconv.Atoi(cnt)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}

		if storeCount[it.ID] < it.Cnt {
			http.Error(w, "Столько нет  на складе", 400)
			return
		}

		upd = append(upd, it)
	}

	for _, it := range upd {
		query := `UPDATE products SET cnt = cnt + $1, sum = sum + ($1 * price), store_count = store_count - $1 WHERE id = $2`
		_, err = s.db.Exec(query, it.Cnt, it.ID)
		if err != nil {
			break
		}
	}

	params := tplMsg{
		Menu:       menu,
		ActiveMenu: r.URL.String(),
		Msg: Msg{
			Class: "primary",
			Text:  "Заказ оформлен",
		},
	}
	if err != nil {
		params.Msg.Class = "danger"
		params.Msg.Text = "Ошибка изменения в баз t"
		log.Println("db err:", err)
	}
	err = msg.Execute(w, params)
	if err != nil {
		// не раскрываем детали ошибки пользователю
		log.Println("error in template:", err)
		http.Error(w, "error in template", 500)
	}

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

	shop := &Shop{db: db}

	http.HandleFunc("/", shop.Index)
	http.HandleFunc("/supply", shop.Supply)
	http.HandleFunc("/edit", shop.Edit)
	http.HandleFunc("/sell", shop.Sell)
	http.HandleFunc("/api/v1/items", shop.ApiGetItems)
	http.HandleFunc("/api/v1/items/delete", shop.ApiDeleteItems)

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
