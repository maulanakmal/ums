package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

type LoginRequest struct {
	Username string
	Password string
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

func loginHandlerPost(w http.ResponseWriter, r *http.Request) {
	body, bodyErr := ioutil.ReadAll(r.Body)
	if bodyErr != nil {
		log.Print("bodyErr ", bodyErr.Error())
		http.Error(w, bodyErr.Error(), http.StatusInternalServerError)
		return
	}

	req := &LoginRequest{}
	json.Unmarshal(body, req)

	stmtSelect, err := db.Prepare("SELECT password FROM user_tab WHERE username = ?")
	if err != nil {
		panic(err.Error())
	}
	defer stmtSelect.Close()

	row := stmtSelect.QueryRow(req.Username)

	var password string
	row.Scan(&password)

	if password == req.Password {
		fmt.Fprint(w, "YES")
	} else {
		fmt.Fprint(w, "NO")
	}
}

func signupHandlerPost(w http.ResponseWriter, r *http.Request) {
	log.Print("hit /signup")
	file, header, err := r.FormFile("file")
	defer file.Close()

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	buf := bytes.NewBuffer(nil)
	if _, err := io.Copy(buf, file); err != nil {
		return
	}

	log.Printf("file header %s", header.Filename)
	cnt, _ := buf.ReadString('\n')
	log.Printf("file content %s", cnt)
}
