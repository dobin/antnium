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
	port     int
	commands []model.Command
}

func NewServer(port int) Server {
	w := Server{port, []model.Command{}}
	return w
}

func (s *Server) Serve() {
	myRouter := mux.NewRouter().StrictSlash(true)
	myRouter.HandleFunc("/admin/listCommands", s.adminListCommands)
	myRouter.HandleFunc("/admin/addCommand", s.adminAddCommand)

	myRouter.HandleFunc("/getCommand", s.getCommand)
	myRouter.HandleFunc("/sendCommand", s.sendCommand)

	fmt.Println("Serving")
	log.Fatal(http.ListenAndServe("127.0.0.1:4444", myRouter))
}

func (s *Server) adminListCommands(rw http.ResponseWriter, r *http.Request) {
	fmt.Fprint(rw, "[")
	for i, command := range s.commands {
		fmt.Fprint(rw, command.Json())
		if i != len(s.commands) {
			fmt.Fprint(rw, ",")
		}
	}
	fmt.Fprint(rw, "]")
}

func (s *Server) adminAddCommand(rw http.ResponseWriter, r *http.Request) {
}

func (s *Server) getCommand(rw http.ResponseWriter, r *http.Request) {
	c := model.NewCommandTest("42", []string{"arg0", "arg1"}, "")
	fmt.Fprint(rw, c.Json())
}

func (s *Server) sendCommand(rw http.ResponseWriter, r *http.Request) {
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatal(err)
	}
	command := model.JsonToCommand(string(reqBody))
	fmt.Println("SendCommand: ")
	fmt.Println(command.Json())
	s.commands = append(s.commands, command)
	fmt.Fprint(rw, "asdf")
}
