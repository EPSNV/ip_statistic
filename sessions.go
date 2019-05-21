package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"time"
)

func GetTemplate(url, text string) []byte {
	return []byte(`
	<html>
		<body>
		<form action="` + url + `" method="post" autocomplete="off">
			Login: <input type="text" name="login">
			Password: <input type="password" name="password">
			<input type="submit" value="` + text + `">
		</form>
		</body>
	</html>
	`)
}

type User struct {
	Login    string
	Password string
}

var (
	users    = make(map[string]*User)
	sessions = make(map[string]string)
)

// аутентификация - проверяем подлинность сессии
func mainPage(w http.ResponseWriter, r *http.Request) {
	sessionCookie, err := r.Cookie("session_id")
	var (
		sessionLogin string
		loggedIn     bool
	)
	if err != http.ErrNoCookie {
		sessionLogin, loggedIn = sessions[sessionCookie.Value]
	}

	w.Header().Set("Content-Type", "text/html")
	if loggedIn {
		user := users[sessionLogin]
		fmt.Fprintln(w, `Hello, `+user.Login)
		fmt.Fprintln(w, `<br><a href="/logout">logout</a>`)
		fmt.Fprintln(w, "<br>Your session: "+sessionCookie.Value)
	} else {
		fmt.Fprintln(w, `<a href="/register">register</a>`)
		fmt.Fprintln(w, `<a href="/login">login</a>`)
	}
}

//  регистрация - создание нового пользователя
func registerPage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Write(GetTemplate("/register", "Register"))
		return
	}

	user := &User{
		Login:    r.FormValue("login"),
		Password: r.FormValue("password"),
	}

	if user.Login == "" {
		http.Error(w, "login cannot be empty", http.StatusBadRequest)
		return
	}
	if _, exists := users[user.Login]; exists {
		http.Error(w, "user already exists", http.StatusBadRequest)
		return
	}

	users[user.Login] = user
	http.Redirect(w, r, "/", http.StatusFound)
}

// авторизация - проверяем что пользователь тот за кого себя выдает
func loginPage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Write(GetTemplate("/login", "Login"))
		return
	}

	user := &User{
		Login:    r.FormValue("login"),
		Password: r.FormValue("password"),
	}

	dbUser, exists := users[user.Login]
	if !exists {
		http.Error(w, "user not exists", http.StatusBadRequest)
		return
	}

	if user.Password != dbUser.Password {
		http.Error(w, "bad password", http.StatusBadRequest)
		return
	}

	sessionID := RandStringRunes(32)
	// опасная операция - я могу перезаписать уже существующую сессию с таким номером
	// правильно проверять что такой номер еще никому не выдан
	sessions[sessionID] = user.Login

	expiration := time.Now().Add(10 * time.Hour)
	cookie := http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		Expires:  expiration,
		HttpOnly: true,
	}
	http.SetCookie(w, &cookie)
	http.Redirect(w, r, "/", http.StatusFound)
}

func logoutPage(w http.ResponseWriter, r *http.Request) {
	session, err := r.Cookie("session_id")
	if err == http.ErrNoCookie {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	session.Expires = time.Now().AddDate(0, 0, -1)
	http.SetCookie(w, session)

	delete(sessions, session.Value)

	http.Redirect(w, r, "/", http.StatusFound)
}

func main() {
	http.HandleFunc("/", mainPage)
	http.HandleFunc("/login", loginPage)
	http.HandleFunc("/logout", logoutPage)
	http.HandleFunc("/register", registerPage)

	fmt.Println("starting server at :8080")
	http.ListenAndServe(":8080", nil)
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
