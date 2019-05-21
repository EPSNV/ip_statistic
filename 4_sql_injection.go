package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	_ "github.com/lib/pq"
)

var (
	db *sql.DB
)

var loginFormTmpl = `
<html>
	<body>
	<form action="/login" method="post">
		Login: <input type="text" name="login">
		Password: <input type="password" name="password">
		<input type="submit" value="Login">
	</form>
	</body>
</html>
`

func main() {

	// инициализация базы
	connStr := "user=root dbname=mydb password=123 host=127.0.0.1 port=5432 sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalln("cant open db:", err)
	}

	tablesQuery := `
DROP TABLE IF EXISTS "users";
DROP SEQUENCE IF EXISTS users_id_seq;
CREATE SEQUENCE users_id_seq INCREMENT 1 MINVALUE 1 MAXVALUE 2147483647 START 1 CACHE 1;

CREATE TABLE "users" (
    "id" integer DEFAULT nextval('users_id_seq') NOT NULL,
    "login" character varying(255) NOT NULL,
    "name" character varying(255) NOT NULL
) WITH (oids = false);

CREATE INDEX "users_login" ON "users" USING btree ("login");

INSERT INTO "users" ("id", "login", "name") VALUES
(1,	'admin',	'Администратор'),
(2,	'user',	'Простой пользователь');
`
	db.Exec(tablesQuery)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(loginFormTmpl))
	})

	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		var (
			id          int
			login, body string
		)

		inputLogin := r.FormValue("login")
		body += fmt.Sprintln("inputLogin:", inputLogin)

		// ПЛОХО! НЕ ДЕЛАЙТЕ ТАК!
		// параметры не экранированы должным образом
		// мы подставляем в запрос параметр как есть
		query := fmt.Sprintf("SELECT id, login FROM users WHERE login = '%s' LIMIT 1", inputLogin)
		row := db.QueryRow(query)
		err := row.Scan(&id, &login)
		body += fmt.Sprintln("Sprint query:", query)
		if err == sql.ErrNoRows {
			body += fmt.Sprintln("Sprint case: NOT FOUND")
		} else {
			body += fmt.Sprintln("Sprint case: id:", id, "login:", login)
		}

		// ПРАВИЛЬНО
		// Мы используем плейсхолдеры, там параметры будет экранирован должным образом
		row = db.QueryRow("SELECT id, login FROM users WHERE login = $1 LIMIT 1", inputLogin)
		err = row.Scan(&id, &login)
		if err == sql.ErrNoRows {
			body += fmt.Sprintln("Placeholders case: NOT FOUND")
		} else {
			body += fmt.Sprintln("Placeholders id:", id, "login:", login)
		}

		w.Write([]byte(body))
	})

	fmt.Println("starting server at :8080")
	http.ListenAndServe(":8080", nil)
}
