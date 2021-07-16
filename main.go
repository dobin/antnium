package main

import (
	"log"
	"mux"
	"net/http"
)

// Webserver contains all data the frontend webserver needs
type Server struct {
	port int
}

// New returns a new webserver instance
func NewServer(port int) Server {
	w := Server{port}
	return w
}

func (s Server) Serve() {
	myRouter := mux.NewRouter().StrictSlash(true)
	myRouter.HandleFunc("/api/{customer}/findRet", w.restFindReg)
	log.Fatal(http.ListenAndServe())
}

func main() {
	s := NewServer(4444)
	s.Serve()
}
