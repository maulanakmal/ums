package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

type LoginRequest struct {
	Username string
	Password string
}

type RegisterRequest struct {
	Username string
	Nickname string
	Password string
}

type ChangeNicknameRequest struct {
	Username string
	Nickname string
}

type User struct {
	Username string
	Nickname string
	Password string
	PicID    string
}

var db *sql.DB

func main() {
	var err error
	db, err = sql.Open("mysql", getDSN())
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()

	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/signup", signupHandler)
	http.HandleFunc("/change_nickname", changeNicknameHandler)
	http.HandleFunc("/change_pic", changePicHandler)

	fmt.Println("Serving on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func getDSN() string {
	db_host_ip := os.Getenv("DB_HOST_IP")
	db_host := "tcp(" + db_host_ip + ")"
	db_user := os.Getenv("DB_USER")
	db_pass := os.Getenv("DB_PASS")
	db_name := os.Getenv("DB_NAME")
	DSN := db_user + ":" + db_pass + "@" + db_host + "/" + db_name

	return DSN
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello World!")
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		loginHandlerPost(w, r)
	}
}

func signupHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		signupHandlerPost(w, r)
	}
}

func changeNicknameHandler(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == http.MethodPost:
		changeNicknameHandlerPost(w, r)
	}
}

func changeNicknameHandlerPost(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	req := &ChangeNicknameRequest{}
	json.Unmarshal(body, req)

	fmt.Fprintf(w, "change nickname request %v", req)

	_, err = queryUser(req.Username)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	_, err = db.Exec("UPDATE user_tab SET nickname = ? WHERE username = ?", req.Nickname, req.Username)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusOK)
}

func changePicHandler(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == http.MethodPost:
		changePicHandlerPost(w, r)
	}
}

func changePicHandlerPost(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	picFile, header, err := r.FormFile("pic")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = queryUser(username)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	log.Print(header.Filename)

	l := len(header.Filename)
	log.Print(header.Filename[l-4:])
	if err != nil {
		fmt.Println(err)
	}

	fileBytes, err := ioutil.ReadAll(picFile)
	if err != nil {
		fmt.Println(err)
	}

	// TODO get dir from config file
	ioutil.WriteFile("/Users/maulanaakmal/.ums-profile-images/"+username+header.Filename[l-4:], fileBytes, 0644)

	_, err = db.Exec("UPDATE user_tab SET pic_id = ? WHERE username = ?", username, username)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusOK)
}

// TODO set up auth system
func loginHandlerPost(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	req := &LoginRequest{}
	json.Unmarshal(body, req)

	var user User
	user, err = queryUser(req.Username)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func signupHandlerPost(w http.ResponseWriter, r *http.Request) {
	log.Print("hit signup")
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	req := RegisterRequest{}
	json.Unmarshal(body, &req)

	_, err = queryUser(req.Username)
	switch {
	case err == nil:
		w.WriteHeader(http.StatusInternalServerError)
		return
	case err == sql.ErrNoRows:
		err = addUser(req)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		return
	default:
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

}

func queryUser(username string) (User, error) {
	var user User
	err := db.QueryRow("SELECT username, nickname, password, pic_id FROM user_tab WHERE username = ?", username).Scan(&user.Username, &user.Nickname, &user.Password, &user.PicID)

	return user, err
}

func addUser(req RegisterRequest) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	log.Print(hash)
	log.Print(len(hash))
	log.Print(string(hash))
	log.Print(len(string(hash)))
	if err != nil {
		return err
	}

	_, err = db.Exec("INSERT INTO user_tab(username, nickname, password) VALUES(?, ?, ?)", req.Username, req.Nickname, hash)
	if err != nil {
		return err
	}

	return nil
}
