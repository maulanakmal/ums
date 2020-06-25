package main

import (
	"database/sql"
	"encoding/gob"
	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net"
	"os"
)

const (
	protocol       = "tcp"
	defaultPort    = "6000"
	defaultAddress = "localhost"
)

type Server struct {
	Port string
}

type Request struct {
	Name string
	Args []string
}

type Response struct {
	Status string
}

type User struct {
	Username string
	Nickname string
	Password string
	PicID    string
}

var db *sql.DB

func getDSN() string {
	db_host_ip := os.Getenv("DB_HOST_IP")
	db_host := "tcp(" + db_host_ip + ")"
	db_user := os.Getenv("DB_USER")
	db_pass := os.Getenv("DB_PASS")
	db_name := os.Getenv("DB_NAME")
	DSN := db_user + ":" + db_pass + "@" + db_host + "/" + db_name

	return DSN
}

func initDB() {
	var err error
	db, err = sql.Open("mysql", getDSN())
	if err != nil {
		panic(err.Error())
	}
}

func (server *Server) ListenAndServe() error {
	initDB()

	listener, err := net.Listen(protocol, defaultAddress+":"+server.Port)

	log.Printf("listening on %s:%s", defaultAddress, server.Port)
	if err != nil {
		panic(err.Error())
	}

	for {
		log.Printf("incoming")
		conn, err := listener.Accept()
		if err != nil {
			panic(err.Error())
		}
		go handleRequest(conn)
	}

	return nil
}

func handleRequest(conn net.Conn) {
	decoder := gob.NewDecoder(conn)
	defer conn.Close()
	log.Printf("inside request")

	var request Request
	decoder.Decode(&request)
	log.Printf("inside request")

	log.Printf("request %v", request)

	switch {
	case request.Name == "login":
	case request.Name == "signup":
	case request.Name == "changeNickname":
	}
}

func login(username string, password string) {
	var user User
	user, err := queryUser(username)
	if err != nil {
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return
	}
}

func singUp(username string, password string, nickname string) {
	_, err := queryUser(username)
	switch {
	case err == nil:
		return
	case err == sql.ErrNoRows:
		err = addUser(username, password, nickname)
		if err != nil {
			return
		}
		return
	default:
		return
	}

}

func changeNickname(username string, nickname string) {
	_, err := queryUser(username)
	if err != nil {
		return
	}

	sqlStatement := "UPDATE user_tab SET nickname = ? WHERE username = ?"
	_, err = db.Exec(sqlStatement, nickname, username)
	if err != nil {
		return
	}
}

func queryUser(username string) (User, error) {
	var user User
	sqlStatement := "SELECT username, nickname, password, pic_id FROM user_tab WHERE username = ?"
	err := db.QueryRow(sqlStatement, username).Scan(&user.Username, &user.Nickname, &user.Password, &user.PicID)

	return user, err
}

func addUser(username string, password string, nickname string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	sqlStatement := "INSERT INTO user_tab(username, nickname, password) VALUES(?, ?, ?)"
	_, err = db.Exec(sqlStatement, username, nickname, hash)
	if err != nil {
		return err
	}

	return nil
}
