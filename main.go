package main

import (
	"bytes"
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

func main() {
	db, err := sql.Open("mysql", getDSN())
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()

	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/singup", homeHandler)
	http.HandleFunc("/edit", homeHandler)

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

func loginHandlerPost(w http.ResponseWriter, r *http.Request) {
	buf, bodyErr := ioutil.ReadAll(r.Body)
	if bodyErr != nil {
		log.Print("bodyErr ", bodyErr.Error())
		http.Error(w, bodyErr.Error(), http.StatusInternalServerError)
		return
	}

	rdr1 := ioutil.NopCloser(bytes.NewBuffer(buf))
	log.Printf("BODY: %q", rdr1)
}
