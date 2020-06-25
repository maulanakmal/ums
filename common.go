package main

type Request struct {
	Name string
	Args []string
}

type Response struct {
	Status  string
	Message string
}
