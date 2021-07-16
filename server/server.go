package server

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/dobin/antnium/model"
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
	myRouter.HandleFunc("/getCommand", s.getCommand)
	myRouter.HandleFunc("/sendCommand", s.sendCommand)

	fmt.Println("Serving")
	log.Fatal(http.ListenAndServe("127.0.0.1:4444", myRouter))
}

func (s Server) getCommand(rw http.ResponseWriter, r *http.Request) {
	c := model.NewCommandTest([]string{"arg0", "arg1"}, "")
	fmt.Fprint(rw, c.Json())
}

func (s Server) sendCommand(rw http.ResponseWriter, r *http.Request) {
	//	c := model.NewCommandTest([]string{"arg0", "arg1"})
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatal(err)
	}
	command := model.JsonToCommand(string(reqBody))
	fmt.Println("SendCommand: ")
	fmt.Println(command.Json())
	fmt.Fprint(rw, "asdf")
}
