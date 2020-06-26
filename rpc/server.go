package rpc

import (
	"database/sql"
	"encoding/gob"
	"github.com/dgrijalva/jwt-go"
	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net"
	"os"
	"time"
)

const (
	protocol       = "tcp"
	defaultPort    = "6000"
	defaultAddress = "localhost"
)

type Server struct {
	Port string
}

type User struct {
	Username string
	Nickname string
	Password string
}

var db *sql.DB
var jwtSecret []byte

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

func initJWT() {
	jwtSecret = []byte(os.Getenv("JWT_SECRET"))
}

func (server *Server) ListenAndServe() error {
	initJWT()
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
		go handleRequest(db, conn)
	}
}

func handleRequest(db *sql.DB, conn net.Conn) {
	decoder := gob.NewDecoder(conn)
	defer conn.Close()
	log.Printf("inside request")

	var request Request
	decoder.Decode(&request)
	log.Printf("inside request")

	log.Printf("request %v", request)

	switch {
	case request.Name == "login":
		login(db, conn, request.Args[0], request.Args[1])
	case request.Name == "signUp":
		singUp(db, conn, request.Args[0], request.Args[1], request.Args[2])
	case request.Name == "changeNickname":
		changeNickname(db, conn, request.Args[0], request.Args[1], request.Args[2])
	}
}

type Claims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

func getJWTToken(username string) (string, error) {
	claims := &Claims{
		Username: username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(60 * time.Minute).Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func decodeJWTToken(token string) (*jwt.Token, error) {
	claims := &Claims{}
	return jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
}

func login(db *sql.DB, conn net.Conn, username string, password string) {
	encoder := gob.NewEncoder(conn)

	failResponse := Response{
		Status:  "ERROR",
		Message: "login failed",
	}
	var user User
	user, err := queryUser(db, username)
	if err != nil {
		encoder.Encode(failResponse)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		encoder.Encode(failResponse)
		return
	}

	tokenString, err := getJWTToken(username)
	successResponse := Response{
		Status:  "OK",
		Message: tokenString,
	}

	encoder.Encode(successResponse)
}

func singUp(db *sql.DB, conn net.Conn, username string, password string, nickname string) {
	log.Println("signup")
	encoder := gob.NewEncoder(conn)

	failResponse := Response{
		Status:  "ERROR",
		Message: "signup failed",
	}

	successResponse := Response{
		Status:  "OK",
		Message: "singup success",
	}

	_, err := queryUser(db, username)
	log.Printf("err = %+v\n", err)
	switch {
	case err == nil:
		encoder.Encode(failResponse)
		log.Printf("log here err nil")
		return
	case err == sql.ErrNoRows:
		err = addUser(db, username, password, nickname)
		log.Printf("err = %+v\n", err)
		if err != nil {
			encoder.Encode(failResponse)
			log.Printf("log here after add user")
			return
		}
		log.Printf("err = %+v\n", err)
		encoder.Encode(successResponse)
	default:
		encoder.Encode(failResponse)
		return
	}

}

func changeNickname(db *sql.DB, conn net.Conn, username string, nickname string, token string) {
	encoder := gob.NewEncoder(conn)

	failResponse := Response{
		Status:  "ERROR",
		Message: "change nickname failed",
	}

	_, err := decodeJWTToken(token)
	if err != nil {
		encoder.Encode(failResponse)
		return
	}
	_, err = queryUser(db, username)
	if err != nil {
		encoder.Encode(failResponse)
		return
	}

	sqlStatement := "UPDATE user_tab SET nickname = ? WHERE username = ?"
	_, err = db.Exec(sqlStatement, nickname, username)
	if err != nil {
		encoder.Encode(failResponse)
		return
	}

	successResponse := Response{
		Status:  "OK",
		Message: "change nickname success",
	}
	encoder.Encode(successResponse)
}

func queryUser(db *sql.DB, username string) (User, error) {
	var user User
	sqlStatement := "SELECT username, nickname, password FROM user_tab WHERE username = ?"
	err := db.QueryRow(sqlStatement, username).Scan(&user.Username, &user.Nickname, &user.Password)

	return user, err
}

func addUser(db *sql.DB, username string, password string, nickname string) error {
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
