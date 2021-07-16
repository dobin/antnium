package server

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

type Server struct {
	port int
}

func NewServer(port int) Server {
	w := Server{port}
	return w
}

func (s Server) Serve() {
	myRouter := mux.NewRouter().StrictSlash(true)
	myRouter.HandleFunc("/test", s.test)

	fmt.Println("Serving")
	log.Fatal(http.ListenAndServe("127.0.0.1:4444", myRouter))
}

func (s Server) test(rw http.ResponseWriter, r *http.Request) {
	data := "test"
	fmt.Fprint(rw, data)
}
