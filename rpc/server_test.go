package rpc

import (
	"encoding/gob"
	"github.com/DATA-DOG/go-sqlmock"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net"
	"os"
	"testing"
)

func TestGetDSN(t *testing.T) {
	os.Setenv("DB_HOST_IP", "test_ip")
	os.Setenv("DB_USER", "test_user")
	os.Setenv("DB_PASS", "test_pass")
	os.Setenv("DB_NAME", "test_name")

	dsn := getDSN()
	if dsn != "test_user:test_pass@tcp(test_ip)/test_name" {
		t.Error("dsn name not same as expected")
	}
}

func TestInitJWT(t *testing.T) {
	os.Setenv("JWT_SECRET", "secret")

	initJWT()

	if string(jwtSecret) != "secret" {
		t.Error("error initing JWT, failed to get secret from env")
	}
}

func TestHandleRequestRouteLogin(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	client, conn := net.Pipe()
	encoder := gob.NewEncoder(client)
	decoder := gob.NewDecoder(client)

	hash, err := bcrypt.GenerateFromPassword([]byte("mol"), bcrypt.DefaultCost)
	if err != nil {
		t.Error("error bcrypting the password")
	}

	mock.ExpectQuery(
		"SELECT username, nickname, password",
	).WillReturnRows(
		sqlmock.NewRows([]string{
			"username", "nickname", "password",
		}).AddRow("mol", "mol", string(hash)),
	)
	go handleRequest(db, conn)

	request := Request{
		Name: "login",
		Args: []string{"mol", "mol"},
	}
	encoder.Encode(request)

	var response Response
	decoder.Decode(&response)

	if response.Status != "OK" {
		t.Error("error login")
	}
}

func TestHandlerRequestRouteSignUp(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	client, conn := net.Pipe()
	encoder := gob.NewEncoder(client)
	decoder := gob.NewDecoder(client)

	mock.ExpectQuery(
		"SELECT username, nickname, password",
	).WillReturnRows(
		sqlmock.NewRows([]string{"username", "nickname", "password"}),
	)
	mock.ExpectExec("INSERT INTO user_tab").WillReturnResult(sqlmock.NewResult(1, 1))
	go handleRequest(db, conn)

	request := Request{
		Name: "signUp",
		Args: []string{"mol", "mol", "mol"},
	}
	encoder.Encode(request)

	var response Response
	decoder.Decode(&response)
	log.Printf("response = %+v\n", response)

	if response.Status != "OK" {
		t.Error("sign up failed")
	}
}

func TestChangeNickname(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	client, conn := net.Pipe()
	encoder := gob.NewEncoder(client)
	decoder := gob.NewDecoder(client)

	go handleRequest(db, conn)

	hash, err := bcrypt.GenerateFromPassword([]byte("mol"), bcrypt.DefaultCost)
	if err != nil {
		t.Error("error bcrypting the password")
	}
	request := Request{
		Name: "login",
		Args: []string{"mol", "mol"},
	}

	mock.ExpectQuery("SELECT username, nickname, password").
		WillReturnRows(
			sqlmock.NewRows([]string{"username", "nickname", "password"}).
				AddRow("mol", "mol", string(hash)),
		)

	mock.ExpectExec("UPDATE user_tab").WillReturnResult(sqlmock.NewResult(1, 1))

	encoder.Encode(request)

	var response Response
	decoder.Decode(&response)

	if response.Status != "OK" {
		t.Error("login failed")
	}

	token := response.Message
	request = Request{
		Name: "changeNickname",
		Args: []string{"mol", "mol", token},
	}

	encoder.Encode(request)

	handleRequest(db, conn)

	decoder.Decode(&response)

	if response.Status != "OK" {
		t.Error("login failed")
	}
}
