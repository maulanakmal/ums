package main

import (
	b64 "encoding/base64"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"./rpc"
)

func main() {
	http.HandleFunc("/", handleHome)
	http.HandleFunc("/login", handleLogin)
	http.HandleFunc("/signUp", handleSignUp)
	http.HandleFunc("/changeNickname", handleChangeNickname)
	http.HandleFunc("/changePicture", handleChangePicture)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleHome(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("hello world"))
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	password := r.FormValue("password")
	client := &rpc.Client{Addr: "localhost:6000"}

	request := rpc.Request{
		Name: "login",
		Args: []string{username, password},
	}

	response, err := client.Call(request)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}

	log.Printf("/login response = %q", response)

}

func handleSignUp(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	password := r.FormValue("password")
	nickname := r.FormValue("nickname")
	log.Printf("%s %s %s", username, password, nickname)
	client := &rpc.Client{Addr: "localhost:6000"}

	request := rpc.Request{
		Name: "signUp",
		Args: []string{username, password, nickname},
	}

	response, err := client.Call(request)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}

	log.Printf("/signUp response = %q", response)

}

func handleChangeNickname(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	nickname := r.FormValue("nickname")
	token := r.FormValue("token")
	client := &rpc.Client{Addr: "localhost:6000"}

	request := rpc.Request{
		Name: "changeNickname",
		Args: []string{username, nickname, token},
	}

	response, err := client.Call(request)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}

	log.Printf("/changeNickname response = %q", response)
}

func handleChangePicture(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	picFile, header, _ := r.FormFile("file")
	defer picFile.Close()

	picBin, err := ioutil.ReadAll(picFile)
	filename := header.Filename
	fileExt := strings.ToLower(filename[len(filename)-3:])

	b64pic := b64.RawStdEncoding.EncodeToString(picBin)
	log.Printf("b64pic %v", b64pic[:100])

	client := &rpc.Client{Addr: "localhost:6000"}
	request := rpc.Request{
		Name: "changePicture",
		Args: []string{username, b64pic, fileExt},
	}

	response, err := client.Call(request)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}

	log.Printf("/changePicture response = %q", response)
}
