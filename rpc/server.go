package rpc

import (
	"context"
	"database/sql"
	b64 "encoding/base64"
	"encoding/gob"
	"log"
	"net"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/go-redis/redis/v8"
	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
)

const (
	protocol       = "tcp"
	defaultPort    = "6000"
	defaultAddress = "localhost"
	redisPort      = "6739"
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
var ctx = context.Background()
var redisClient *redis.Client

func getDSN() string {
	dbHostIP := os.Getenv("DB_HOST_IP")
	dbHost := "tcp(" + dbHostIP + ")"
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASS")
	dbName := os.Getenv("DB_NAME")
	DSN := dbUser + ":" + dbPass + "@" + dbHost + "/" + dbName

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

func initRedis() {
	dbHostIP := os.Getenv("DB_HOST_IP")
	redisClient = redis.NewClient(&redis.Options{
		Addr:     dbHostIP + ":" + redisPort,
		Password: "",
		DB:       0,
	})

	pong, err := redisClient.Ping(ctx).Result()
	log.Println(pong, err)
}

func (server *Server) ListenAndServe() error {
	initJWT()
	initDB()
	initRedis()

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
	case request.Name == "changePicture":
		changeProfilePic(db, conn, request.Args[0], request.Args[1], request.Args[2], request.Args[3])
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

func changeProfilePic(db *sql.DB, conn net.Conn, username string, base64pic string, fileExt string, token string) {
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

	pic, err := b64.RawStdEncoding.DecodeString(base64pic)
	if err != nil {
		encoder.Encode(failResponse)
	}

	file, err := os.OpenFile("/var/ums/pic/"+username+"."+fileExt, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		encoder.Encode(failResponse)
	}
	defer file.Close()

	_, err = file.Write(pic)
	if err != nil {
		encoder.Encode(failResponse)
	}

	successResponse := Response{
		Status:  "OK",
		Message: "change profile success",
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
